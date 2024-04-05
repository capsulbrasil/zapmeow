package service

import (
	"zapmeow/api/helper"
	"zapmeow/api/model"
	"zapmeow/api/queue"
	"zapmeow/api/response"
	"zapmeow/pkg/http"
	"zapmeow/pkg/logger"
	"zapmeow/pkg/whatsapp"
	"zapmeow/pkg/zapmeow"

	"github.com/vincent-petithory/dataurl"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	"google.golang.org/protobuf/proto"
)

type whatsAppService struct {
	app            *zapmeow.ZapMeow
	messageService MessageService
	accountService AccountService
	whatsApp       whatsapp.WhatsApp
}

type WhatsAppService interface {
	GetInstance(instanceID string) (*whatsapp.Instance, error)
	IsAuthenticated(instance *whatsapp.Instance) bool
	Logout(instance *whatsapp.Instance) error
	SendTextMessage(instance *whatsapp.Instance, jid whatsapp.JID, text string) (whatsapp.MessageResponse, error)
	SendAudioMessage(instance *whatsapp.Instance, jid whatsapp.JID, audioURL *dataurl.DataURL, mimitype string) (whatsapp.MessageResponse, error)
	SendDocumentMessage(instance *whatsapp.Instance, jid whatsapp.JID, documentURL *dataurl.DataURL, mimitype string, filename string) (whatsapp.MessageResponse, error)
	SendImageMessage(instance *whatsapp.Instance, jid whatsapp.JID, imageURL *dataurl.DataURL, mimitype string) (whatsapp.MessageResponse, error)
	GetContactInfo(instance *whatsapp.Instance, jid whatsapp.JID) (*whatsapp.ContactInfo, error)
	ParseEventMessage(instance *whatsapp.Instance, message *events.Message) (whatsapp.Message, error)
	IsOnWhatsApp(instance *whatsapp.Instance, phones []string) ([]whatsapp.IsOnWhatsAppResponse, error)
}

func NewWhatsAppService(
	app *zapmeow.ZapMeow,
	messageService MessageService,
	accountService AccountService,
	whatsApp whatsapp.WhatsApp,
) *whatsAppService {
	return &whatsAppService{
		app:            app,
		messageService: messageService,
		accountService: accountService,
		whatsApp:       whatsApp,
	}
}

func (w *whatsAppService) SendTextMessage(
	instance *whatsapp.Instance,
	jid whatsapp.JID,
	text string,
) (whatsapp.MessageResponse, error) {
	return w.whatsApp.SendTextMessage(instance, jid, text)
}

func (w *whatsAppService) SendDocumentMessage(
	instance *whatsapp.Instance,
	jid whatsapp.JID,
	documentURL *dataurl.DataURL,
	mimitype string,
	filename string,
) (whatsapp.MessageResponse, error) {
	return w.whatsApp.SendDocumentMessage(instance, jid, documentURL, mimitype, filename)
}

func (w *whatsAppService) SendAudioMessage(
	instance *whatsapp.Instance,
	jid whatsapp.JID,
	audioURL *dataurl.DataURL,
	mimitype string,
) (whatsapp.MessageResponse, error) {
	return w.whatsApp.SendAudioMessage(instance, jid, audioURL, mimitype)
}

func (w *whatsAppService) SendImageMessage(
	instance *whatsapp.Instance,
	jid whatsapp.JID,
	imageURL *dataurl.DataURL,
	mimitype string,
) (whatsapp.MessageResponse, error) {
	return w.whatsApp.SendImageMessage(instance, jid, imageURL, mimitype)
}

func (w *whatsAppService) GetContactInfo(instance *whatsapp.Instance, jid whatsapp.JID) (*whatsapp.ContactInfo, error) {
	return w.whatsApp.GetContactInfo(instance, jid)
}

func (w *whatsAppService) ParseEventMessage(instance *whatsapp.Instance, message *events.Message) (whatsapp.Message, error) {
	return w.whatsApp.ParseEventMessage(instance, message)
}

func (w *whatsAppService) IsOnWhatsApp(instance *whatsapp.Instance, phones []string) ([]whatsapp.IsOnWhatsAppResponse, error) {
	return w.whatsApp.IsOnWhatsApp(instance, phones)
}

func (w *whatsAppService) GetInstance(instanceID string) (*whatsapp.Instance, error) {
	instance := w.app.LoadInstance(instanceID)
	if instance != nil {
		return instance, nil
	}

	instance, err := w.gerOrCreateInstance(instanceID)
	if err != nil {
		return nil, err
	}
	w.app.StoreInstance(instanceID, instance)

	instance = w.app.LoadInstance(instanceID)
	instance.Client.AddEventHandler(func(evt interface{}) {
		w.eventHandler(instanceID, evt)
	})

	err = w.whatsApp.InitInstance(instance, func(event string, code string, err error) {
		switch event {
		case "code":
			{
				instance.QrCodeRateLimit -= 1
				err = w.accountService.UpdateAccount(instanceID, map[string]interface{}{
					"QrCode":    code,
					"Status":    "UNPAIRED",
					"WasSynced": false,
				})
				if err != nil {
					logger.Error("Failed to update account. ", err)
				}
			}
		case "error":
			{
			}
		case "rate-limit":
			{
				err := w.deleteInstance(instance)
				if err != nil {
					logger.Info("a")
					// logger.Error("Failed to destroy instance. ", err)
				}
				return
			}
		case "timeout":
			{
				err := w.accountService.UpdateAccount(instanceID, map[string]interface{}{
					"QrCode": "",
					"Status": "TIMEOUT",
				})
				if err != nil {
					logger.Error("Failed to update account. ", err)
				}

				w.deleteInstance(instance)
			}

		}
	})
	if err != nil {
		return nil, err
	}

	return instance, nil
}

func (w *whatsAppService) IsAuthenticated(instance *whatsapp.Instance) bool {
	return w.whatsApp.IsConnected(instance) && w.whatsApp.IsLoggedIn(instance)
}

func (w *whatsAppService) Logout(instance *whatsapp.Instance) error {
	err := w.whatsApp.Logout(instance)
	if err != nil {
		return err
	}

	err = w.accountService.UpdateAccount(instance.ID, map[string]interface{}{
		"Status": "UNPAIRED",
	})
	if err != nil {
		return err
	}

	return w.deleteInstance(instance)
}

func (w *whatsAppService) gerOrCreateInstance(instanceID string) (*whatsapp.Instance, error) {
	account, err := w.accountService.GetAccountByInstanceID(instanceID)
	if err != nil {
		return nil, err
	}

	if account == nil || (account != nil && account.Status != "CONNECTED") {
		instance := w.whatsApp.CreateInstance(instanceID)

		err := w.accountService.CreateAccount(&model.Account{
			InstanceID: instanceID,
		})
		if err != nil {
			return nil, err
		}
		return instance, nil
	}

	jid := types.JID{
		User:       account.User,
		RawAgent:   account.RawAgent,
		Device:     account.Device,
		Integrator: account.Integrator,
		Server:     account.Server,
	}
	instance := w.whatsApp.CreateInstanceFromDevice(
		instanceID,
		jid,
	)
	return instance, nil
}

func (w *whatsAppService) deleteInstance(instance *whatsapp.Instance) error {
	err := w.accountService.DeleteAccountMessages(instance.ID)
	if err != nil {
		return err
	}

	w.whatsApp.Disconnect(instance)
	w.app.DeleteInstance(instance.ID)
	return nil
}

func (w *whatsAppService) eventHandler(instanceID string, rawEvt interface{}) {
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

func (w *whatsAppService) handleHistorySync(instanceID string, evt *events.HistorySync) {
	if !w.app.Config.HistorySync {
		return
	}
	history, _ := proto.Marshal(evt.Data)

	q := queue.NewHistorySyncQueue(w.app)
	err := q.Enqueue(queue.HistorySyncQueueData{
		History:    history,
		InstanceID: instanceID,
	})

	if err != nil {
		logger.Error("Failed to add history sync to queue. ", err)
	}
}

func (w *whatsAppService) handleConnected(instanceID string) {
	var instance = w.app.LoadInstance(instanceID)
	err := w.accountService.UpdateAccount(instanceID, map[string]interface{}{
		"User":       instance.Client.Store.ID.User,
		"RawAgent":   instance.Client.Store.ID.RawAgent,
		"Device":     instance.Client.Store.ID.Device,
		"Server":     instance.Client.Store.ID.Server,
		"Integrator": instance.Client.Store.ID.Integrator,
		"InstanceID": instance.ID,
		"Status":     "CONNECTED",
		"QrCode":     "",
		"WasSynced":  false,
	})

	if err != nil {
		logger.Error("Failed to update account. ", err)
	}
}

func (w *whatsAppService) handleLoggedOut(instanceID string) {
	instance, err := w.GetInstance(instanceID)
	if err != nil {
		logger.Error(err)
		return
	}

	err = w.deleteInstance(instance)
	if err != nil {
		logger.Error(err)
	}

	err = w.accountService.UpdateAccount(instanceID, map[string]interface{}{
		"Status": "UNPAIRED",
	})
	if err != nil {
		logger.Error("Failed to update account. ", err)
	}
}

func (w *whatsAppService) handleMessage(instanceId string, evt *events.Message) {
	instance := w.app.LoadInstance(instanceId)
	parsedEventMessage, err := w.whatsApp.ParseEventMessage(instance, evt)

	if err != nil {
		logger.Error(err)
		return
	}

	message := model.Message{
		SenderJID:  parsedEventMessage.SenderJID,
		ChatJID:    parsedEventMessage.ChatJID,
		InstanceID: parsedEventMessage.InstanceID,
		MessageID:  parsedEventMessage.MessageID,
		Timestamp:  parsedEventMessage.Timestamp,
		Body:       parsedEventMessage.Body,
		FromMe:     parsedEventMessage.FromMe,
	}

	if parsedEventMessage.MediaType != nil {
		path, err := helper.SaveMedia(
			instance.ID,
			parsedEventMessage.MessageID,
			*parsedEventMessage.Media,
			*parsedEventMessage.Mimetype,
		)

		if err != nil {
			logger.Error("Failed to save media. ", err)
		}

		message.MediaType = parsedEventMessage.MediaType.String()
		message.MediaPath = path
	}

	err = w.messageService.CreateMessage(&message)
	if err != nil {
		logger.Error("Failed to create message. ", err)
		return
	}

	body := map[string]interface{}{
		"instanceId": instanceId,
		"message":    response.NewMessageResponse(message),
	}

	err = http.Request(w.app.Config.WebhookURL, body)
	if err != nil {
		logger.Error("Failed to send webhook request. ", err)
	}
}
