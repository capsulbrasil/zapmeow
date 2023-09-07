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
		case <-q.app.StopCh:
			return
		default:
			data, err := queue.Dequeue()
			if err == nil && data == nil {
				time.Sleep(time.Second)
				continue
			} else if err != nil {
				fmt.Println(err)
				continue
			}

			var evt waProto.HistorySync
			err = proto.Unmarshal(data.History, &evt)
			if err != nil {
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
				err := q.wppService.DeleteInstanceMessages(account.InstanceID)
				if err != nil {
					fmt.Println(err)
					continue
				}

				err = q.accountService.UpdateAccount(account.InstanceID, map[string]interface{}{
					"WasSynced": true,
				})
				if err != nil {
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

				var historySyncMsgs = conv.GetMessages()
				sort.Slice(historySyncMsgs, func(i, j int) bool {
					message1, _ := instance.Client.ParseWebMessage(chatJID, historySyncMsgs[i].GetMessage())
					message2, _ := instance.Client.ParseWebMessage(chatJID, historySyncMsgs[j].GetMessage())
					return message1.Info.Timestamp.After(message2.Info.Timestamp)
				})

				var limit = utils.Min(q.app.Config.MessageLimit, len(historySyncMsgs))
				var slice = historySyncMsgs[:limit]
				for _, historySyncMsg := range slice {
					eventMessage, _ := instance.Client.ParseWebMessage(chatJID, historySyncMsg.GetMessage())
					message := q.messageService.Parse(instance, eventMessage)
					if message != nil {
						messages = append(messages, *message)
					}
				}
			}

			err = q.messageService.CreateMessages(&messages)
			if err != nil {
				fmt.Println(err)
			}
			continue
		}
	}
}
