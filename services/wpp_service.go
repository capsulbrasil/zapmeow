package services

import (
	"context"
	"errors"
	"fmt"
	"time"
	"zapmeow/configs"
	"zapmeow/models"
	"zapmeow/queues"
	"zapmeow/utils"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	waLog "go.mau.fi/whatsmeow/util/log"
	"google.golang.org/protobuf/proto"
)

type WppService struct {
	app            *configs.App
	messageService MessageService
	accountService AccountService
}

func NewWppService(
	app *configs.App,
	messageService MessageService,
	accountService AccountService,
) *WppService {
	return &WppService{
		app:            app,
		messageService: messageService,
		accountService: accountService,
	}
}

func (w *WppService) GetInstance(instanceID string) (*whatsmeow.Client, error) {
	instance, ok := w.app.Instances[instanceID]

	if ok && instance != nil {
		return instance, nil
	}

	client, err := w.getClient(instanceID)
	if err != nil {
		return nil, err
	}

	w.app.Instances[instanceID] = client
	w.app.Instances[instanceID].AddEventHandler(func(evt interface{}) {
		w.eventHandler(instanceID, evt)
	})

	if w.app.Instances[instanceID].Store.ID == nil {
		go w.qrcode(instanceID)
	} else {
		err := w.app.Instances[instanceID].Connect()
		if err != nil {
			return nil, err
		}

		if !w.app.Instances[instanceID].WaitForConnection(5 * time.Second) {
			return nil, errors.New("websocket didn't reconnect within 5 seconds of failed")
		}
	}

	return w.app.Instances[instanceID], nil
}

func (w *WppService) GetAuthenticatedInstance(instanceID string) (*whatsmeow.Client, error) {
	instance, err := w.GetInstance(instanceID)
	if err != nil {
		return nil, err
	}

	if !instance.IsConnected() {
		return nil, errors.New("instance not connected")
	}

	if !instance.IsLoggedIn() {
		return nil, errors.New("inauthenticated instance")
	}

	return instance, nil
}

func (w *WppService) getClient(instanceID string) (*whatsmeow.Client, error) {
	account, err := w.accountService.GetAccountByInstanceID(instanceID)
	if err != nil {
		return nil, err
	}

	if account == nil {
		err := w.accountService.CreateAccount(&models.Account{
			InstanceID: instanceID,
		})

		if err != nil {
			return nil, err
		}
		return createClient(
			w.app.WhatsmeowContainer.NewDevice(),
		), nil
	} else if account.Status != "CONNECTED" {
		return createClient(
			w.app.WhatsmeowContainer.NewDevice(),
		), nil
	}

	device, err := w.app.WhatsmeowContainer.GetDevice(types.JID{
		User:   account.User,
		Agent:  account.Agent,
		Device: account.Device,
		Server: account.Server,
		AD:     account.AD,
	})
	if err != nil {
		return nil, err
	}

	if device != nil {
		return createClient(device), nil
	}

	device = w.app.WhatsmeowContainer.NewDevice()
	return createClient(device), nil
}

func (w *WppService) qrcode(instanceID string) {
	client := w.app.Instances[instanceID]
	if client.Store.ID == nil {
		qrChan, err := client.GetQRChannel(context.Background())
		if err != nil {
			if !errors.Is(err, whatsmeow.ErrQRStoreContainsID) {
				fmt.Println("failed to get qr channel")
			}
		} else {
			err = client.Connect()
			if err != nil {
				fmt.Println("[qrcode]: ", err)
				return
			}
			for evt := range qrChan {
				switch evt.Event {
				case "success":
					fmt.Println("[qrcode]: success")
					return
				case "timeout":
					fmt.Println("[qrcode]: timeout error")
					err := w.accountService.UpdateAccount(instanceID, map[string]interface{}{
						"QrCode": "",
						"Status": "TIMEOUT",
					})
					if err != nil {
						fmt.Println("[qrcode]: ", err)
					}

					delete(w.app.Instances, instanceID)
				case "code":
					w.accountService.UpdateAccount(instanceID, map[string]interface{}{
						"QrCode":    evt.Code,
						"Status":    "UNPAIRED",
						"WasSynced": false,
					})
					if err != nil {
						fmt.Println("[qrcode]: ", err)
					}
				}
			}
		}
	}
}

func (w *WppService) eventHandler(instanceID string, rawEvt interface{}) {
	switch evt := rawEvt.(type) {
	case *events.Message:
		w.handleMessage(instanceID, evt)
	case *events.HistorySync:
		w.handleHistorySync(instanceID, evt)
	case *events.Connected:
		w.handleConnected(instanceID)
	case *events.LoggedOut:
		w.handleLoggedOut(instanceID)
	}
}

func (w *WppService) handleHistorySync(instanceID string, evt *events.HistorySync) {
	history, _ := proto.Marshal(evt.Data)

	queue := queues.NewHistorySyncQueue(w.app)
	err := queue.Enqueue(queues.HistorySyncQueueData{
		History:    history,
		InstanceID: instanceID,
	})

	if err != nil {
		fmt.Println("Error adding item to queue: ", err)
	}
}

func (w *WppService) handleConnected(instanceID string) {
	var instance = w.app.Instances[instanceID]
	err := w.accountService.UpdateAccount(instanceID, map[string]interface{}{
		"User":       instance.Store.ID.User,
		"Agent":      instance.Store.ID.Agent,
		"Device":     instance.Store.ID.Device,
		"Server":     instance.Store.ID.Server,
		"AD":         instance.Store.ID.AD,
		"Status":     "CONNECTED",
		"QrCode":     "",
		"InstanceID": instanceID,
		"WasSynced":  false,
	})

	if err != nil {
		fmt.Println("Error creating account:", err)
		return
	}
}

func (w *WppService) handleLoggedOut(instanceID string) {
	instance := w.app.Instances[instanceID]

	_, err := w.accountService.GetAccountByInstanceID(instanceID)
	if err != nil {
		fmt.Println(err)
		return
	}

	err = w.accountService.UpdateAccount(instanceID, map[string]interface{}{
		"Status": "UNPAIRED",
	})

	if err != nil {
		fmt.Println("Error", err)
		return
	}

	instance.Disconnect()
	delete(w.app.Instances, instanceID)
}

func (w *WppService) handleMessage(instanceId string, evt *events.Message) {
	instance := w.app.Instances[instanceId]
	message := w.messageService.Parse(instance, evt)

	if message == nil {
		return
	}

	err := w.messageService.CreateMessage(message)
	if err != nil {
		fmt.Println(err)
	}

	body := map[string]interface{}{
		"InstanceId": instanceId,
		"Message":    w.messageService.ToJSON(*message),
	}

	err = utils.Request(w.app.Config.WebhookURL, body)

	if err != nil {
		fmt.Println("Error when send request:", err)
	}
}

func createClient(deviceStore *store.Device) *whatsmeow.Client {
	log := waLog.Stdout("Client", "DEBUG", true)
	return whatsmeow.NewClient(deviceStore, log)
}
