package workers

import (
	"sort"
	"time"
	"zapmeow/configs"
	"zapmeow/models"
	"zapmeow/queues"
	"zapmeow/services"
	"zapmeow/utils"

	waProto "go.mau.fi/whatsmeow/binary/proto"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	"google.golang.org/protobuf/proto"
)

type historySyncWorker struct {
	app            *configs.ZapMeow
	messageService services.MessageService
	accountService services.AccountService
	wppService     services.WppService
	log            configs.Logger
}

type HistorySyncWorker interface {
	ProcessQueue()
}

func NewHistorySyncWorker(
	app *configs.ZapMeow,
	messageService services.MessageService,
	accountService services.AccountService,
	wppService services.WppService,
	log configs.Logger,
) *historySyncWorker {
	return &historySyncWorker{
		messageService: messageService,
		accountService: accountService,
		wppService:     wppService,
		app:            app,
		log:            log,
	}
}

func (q *historySyncWorker) ProcessQueue() {
	queue := queues.NewHistorySyncQueue(q.app, q.log)

	defer q.app.Wg.Done()
	for {
		select {
		case <-*q.app.StopCh:
			return
		default:
			if err := q.processHistorySync(queue); err != nil {
				q.log.Error("Error processing history sync. ", err)
			}
		}

		time.Sleep(3 * time.Second)
	}
}

func (q *historySyncWorker) processHistorySync(queue queues.HistorySyncQueue) error {
	data, err := queue.Dequeue()
	if err != nil {
		return err
	}

	if data == nil {
		return nil
	}

	historySync, err := q.parseHistorySync(data.History)
	if err != nil {
		return err
	}

	instance, err := q.wppService.GetInstance(data.InstanceID)
	if err != nil {
		return err
	}

	account, err := q.accountService.GetAccountByInstanceID(data.InstanceID)
	if err != nil {
		return err
	}

	if !account.WasSynced {
		if err := q.accountService.DeleteAccountMessages(account.InstanceID); err != nil {
			return err
		}

		if err := q.accountService.UpdateAccount(account.InstanceID, map[string]interface{}{
			"WasSynced": true,
		}); err != nil {
			return err
		}
	}

	messages, err := q.processMessages(historySync, account, instance)
	if err != nil {
		return err
	}

	if err := q.messageService.CreateMessages(&messages); err != nil {
		return err
	}

	return nil
}

func (q *historySyncWorker) parseHistorySync(history []byte) (*waProto.HistorySync, error) {
	var data waProto.HistorySync
	if err := proto.Unmarshal(history, &data); err != nil {
		return nil, err
	}
	return &data, nil
}

func (q *historySyncWorker) processMessages(evt *waProto.HistorySync, account *models.Account, instance *configs.Instance) ([]models.Message, error) {
	var messages []models.Message

	for _, conv := range evt.GetConversations() {
		chatJID, _ := types.ParseJID(conv.GetId())

		count, err := q.messageService.CountChatMessages(account.InstanceID, chatJID.User)
		if err != nil {
			return nil, err
		}

		if count > int64(q.app.Config.MaxMessagesPerInstance) {
			continue
		}

		historySyncMsgs := conv.GetMessages()
		if historySyncMsgs == nil || len(historySyncMsgs) == 0 {
			continue
		}

		eventsMessage, err := q.processConversation(conv, chatJID, instance)
		if err != nil {
			return nil, err
		}

		sort.Slice(eventsMessage, func(i, j int) bool {
			return eventsMessage[i].Info.Timestamp.After(eventsMessage[j].Info.Timestamp)
		})

		maxMessages := utils.Min(q.app.Config.MaxMessagesPerInstance, len(eventsMessage))
		slice := eventsMessage[:maxMessages]

		for _, evtMessage := range slice {
			parsedEvtMesage, err := q.wppService.ParseEventMessage(instance, evtMessage)
			if err != nil {
				continue
			}

			message, err := q.makeMessage(instance, parsedEvtMesage)
			if err != nil {
				continue
			}
			messages = append(messages, *message)
		}
	}

	return messages, nil
}

func (q *historySyncWorker) processConversation(conv *waProto.Conversation, chatJID types.JID, instance *configs.Instance) ([]*events.Message, error) {
	var eventsMessage []*events.Message
	for _, msg := range conv.GetMessages() {
		parsedMessage, err := instance.Client.ParseWebMessage(chatJID, msg.GetMessage())
		if err != nil {
			continue
		}
		eventsMessage = append(eventsMessage, parsedMessage)
	}
	return eventsMessage, nil
}

func (q *historySyncWorker) makeMessage(instance *configs.Instance, parsedMessage services.ParsedEventMessage) (*models.Message, error) {
	message := models.Message{
		SenderJID:  parsedMessage.SenderJID,
		ChatJID:    parsedMessage.ChatJID,
		InstanceID: parsedMessage.InstanceID,
		MessageID:  parsedMessage.MessageID,
		Timestamp:  parsedMessage.Timestamp,
		Body:       parsedMessage.Body,
		FromMe:     parsedMessage.FromMe,
	}

	if parsedMessage.MediaType != nil {
		path, err := utils.SaveMedia(
			instance.ID,
			parsedMessage.MessageID,
			*parsedMessage.Media,
			*parsedMessage.Mimetype,
		)

		if err != nil {
			return nil, err
		}

		message.MediaType = parsedMessage.MediaType.String()
		message.MediaPath = path
	}

	return &message, nil
}
