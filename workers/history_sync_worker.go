package workers

import (
	"fmt"
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
}

type HistorySyncWorker interface {
	ProcessQueue()
}

func NewHistorySyncWorker(
	app *configs.ZapMeow,
	messageService services.MessageService,
	accountService services.AccountService,
	wppService services.WppService,
) *historySyncWorker {
	return &historySyncWorker{
		messageService: messageService,
		accountService: accountService,
		wppService:     wppService,
		app:            app,
	}
}

func (q *historySyncWorker) ProcessQueue() {
	queue := queues.NewHistorySyncQueue(q.app)

	defer q.app.Wg.Done()
	for {
		select {
		case <-*q.app.StopCh:
			return
		default:
			data, err := queue.Dequeue()
			if err != nil {
				fmt.Println(err)
				continue
			}

			if data == nil {
				time.Sleep(time.Second)
				continue
			}

			var evt waProto.HistorySync
			if err := proto.Unmarshal(data.History, &evt); err != nil {
				fmt.Println("proto error ", err)
				continue
			}

			instance := q.app.Instances[data.InstanceID]
			account, err := q.accountService.GetAccountByInstanceID(data.InstanceID)
			if err != nil {
				fmt.Println(err)
				continue
			}

			if !account.WasSynced {
				if err := q.accountService.DeleteAccountMessages(account.InstanceID); err != nil {
					fmt.Println(err)
					continue
				}

				if err := q.accountService.UpdateAccount(account.InstanceID, map[string]interface{}{
					"WasSynced": true,
				}); err != nil {
					fmt.Println(err)
					continue
				}
			}

			var messages []models.Message
			for _, conv := range evt.GetConversations() {
				chatJID, _ := types.ParseJID(conv.GetId())

				count, err := q.messageService.CountChatMessages(account.InstanceID, chatJID.User)
				if err != nil {
					fmt.Println(err)
					continue
				}

				if count > int64(q.app.Config.MessageLimit) {
					continue
				}

				historySyncMsgs := conv.GetMessages()
				if historySyncMsgs == nil || len(historySyncMsgs) == 0 {
					continue
				}

				var eventsMessage []*events.Message
				for _, msg := range historySyncMsgs {
					parsedMessage, err := instance.Client.ParseWebMessage(chatJID, msg.GetMessage())
					if err != nil {
						continue
					}
					eventsMessage = append(eventsMessage, parsedMessage)
				}

				sort.Slice(eventsMessage, func(i, j int) bool {
					return eventsMessage[i].Info.Timestamp.After(eventsMessage[j].Info.Timestamp)
				})

				limit := utils.Min(q.app.Config.MessageLimit, len(eventsMessage))
				slice := eventsMessage[:limit]

				for _, eventMessage := range slice {
					parsedEventMessage, err := q.wppService.ParseEventMessage(instance, eventMessage)

					message := models.Message{
						SenderJID:  parsedEventMessage.SenderJID,
						ChatJID:    parsedEventMessage.ChatJID,
						InstanceID: parsedEventMessage.InstanceID,
						MessageID:  parsedEventMessage.MessageID,
						Timestamp:  parsedEventMessage.Timestamp,
						Body:       parsedEventMessage.Body,
						FromMe:     parsedEventMessage.FromMe,
					}

					if parsedEventMessage.MediaType != nil {
						path, err := utils.SaveMedia(
							instance.ID,
							parsedEventMessage.MessageID,
							*parsedEventMessage.Media,
							*parsedEventMessage.Mimetype,
						)

						if err != nil {
							fmt.Println(err)
						}

						message.MediaType = parsedEventMessage.MediaType.String()
						message.MediaPath = path
					}

					if err != nil {
						messages = append(messages, message)
					}
				}
			}

			if err := q.messageService.CreateMessages(&messages); err != nil {
				fmt.Println(err)
			}
		}
	}
}
