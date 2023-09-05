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

type wppService struct {
	app            *configs.ZapMeow
	messageService MessageService
	accountService AccountService
}

type WppService interface {
	GetInstance(instanceID string) (*configs.Instance, error)
	GetAuthenticatedInstance(instanceID string) (*configs.Instance, error)
	GetContactInfo(instanceID string, jid types.JID) (map[string]interface{}, error)
}

func NewWppService(
	app *configs.ZapMeow,
	messageService MessageService,
	accountService AccountService,
) *wppService {
	return &wppService{
		app:            app,
		messageService: messageService,
		accountService: accountService,
	}
}

func (w *wppService) GetInstance(instanceID string) (*configs.Instance, error) {
	instance, ok := w.app.Instances[instanceID]

	if ok && instance != nil {
		return instance, nil
	}

	client, err := w.getClient(instanceID)
	if err != nil {
		return nil, err
	}

	w.app.Instances[instanceID] = &configs.Instance{
		ID:     instanceID,
		Client: client,
	}
	w.app.Instances[instanceID].Client.AddEventHandler(func(evt interface{}) {
		w.eventHandler(instanceID, evt)
	})

	if w.app.Instances[instanceID].Client.Store.ID == nil {
		go w.qrcode(instanceID)
	} else {
		err := w.app.Instances[instanceID].Client.Connect()
		if err != nil {
			return nil, err
		}

		if !w.app.Instances[instanceID].Client.WaitForConnection(5 * time.Second) {
			return nil, errors.New("websocket didn't reconnect within 5 seconds of failed")
		}
	}

	return w.app.Instances[instanceID], nil
}

func (w *wppService) GetAuthenticatedInstance(instanceID string) (*configs.Instance, error) {
	instance, err := w.GetInstance(instanceID)
	if err != nil {
		return nil, err
	}

	if !instance.Client.IsConnected() {
		return nil, errors.New("instance not connected")
	}

	if !instance.Client.IsLoggedIn() {
		return nil, errors.New("inauthenticated instance")
	}

	return instance, nil
}

func (w *wppService) GetContactInfo(instanceID string, jid types.JID) (map[string]interface{}, error) {
	instance, err := w.GetAuthenticatedInstance(instanceID)
	if err != nil {
		return nil, err
	}

	userInfo, err := instance.Client.GetUserInfo([]types.JID{jid})
	if err != nil {
		return nil, err
	}

	contactInfo, err := instance.Client.Store.Contacts.GetContact(jid)
	if err != nil {
		return nil, err
	}

	profilePictureInfo, err := instance.Client.GetProfilePictureInfo(
		jid,
		&whatsmeow.GetProfilePictureParams{},
	)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"Phone":   jid.User,
		"Name":    contactInfo.PushName,
		"Status":  userInfo[jid].Status,
		"Picture": profilePictureInfo.URL,
	}, nil
}

func (w *wppService) getClient(instanceID string) (*whatsmeow.Client, error) {
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

func (w *wppService) qrcode(instanceID string) {
	instance := w.app.Instances[instanceID]
	if instance.Client.Store.ID == nil {
		qrChan, err := instance.Client.GetQRChannel(context.Background())
		if err != nil {
			if !errors.Is(err, whatsmeow.ErrQRStoreContainsID) {
				fmt.Println("failed to get qr channel")
			}
		} else {
			err = instance.Client.Connect()
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

func (w *wppService) eventHandler(instanceID string, rawEvt interface{}) {
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

func (w *wppService) handleHistorySync(instanceID string, evt *events.HistorySync) {
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

func (w *wppService) handleConnected(instanceID string) {
	var instance = w.app.Instances[instanceID]
	err := w.accountService.UpdateAccount(instanceID, map[string]interface{}{
		"User":       instance.Client.Store.ID.User,
		"Agent":      instance.Client.Store.ID.Agent,
		"Device":     instance.Client.Store.ID.Device,
		"Server":     instance.Client.Store.ID.Server,
		"AD":         instance.Client.Store.ID.AD,
		"InstanceID": instance.ID,
		"Status":     "CONNECTED",
		"QrCode":     "",
		"WasSynced":  false,
	})

	if err != nil {
		fmt.Println("Error creating account:", err)
		return
	}
}

func (w *wppService) handleLoggedOut(instanceID string) {
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

	err = w.accountService.DeleteAccountInfos(instanceID)
	if err != nil {
		fmt.Println(err)
		return
	}

	instance.Client.Disconnect()
	delete(w.app.Instances, instanceID)
}

func (w *wppService) handleMessage(instanceId string, evt *events.Message) {
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
