package workers

import (
	"fmt"
	"sort"
	"time"
	"zapmeow/configs"
	"zapmeow/models"
	"zapmeow/queues"
	"zapmeow/repositories"
	"zapmeow/services"
	"zapmeow/utils"

	waProto "go.mau.fi/whatsmeow/binary/proto"
	"go.mau.fi/whatsmeow/types"
	"google.golang.org/protobuf/proto"
)

type HistorySyncWorker struct {
	messageService services.MessageService
	accountRepo    repositories.AccountRepository
	messageRepo    repositories.MessageRepository
	app            *configs.App
}

func NewHistorySyncWorker(
	app *configs.App,
	messageService services.MessageService,
	accountRepo repositories.AccountRepository,
	messageRepo repositories.MessageRepository,
) *HistorySyncWorker {
	return &HistorySyncWorker{
		messageService: messageService,
		accountRepo:    accountRepo,
		messageRepo:    messageRepo,
		app:            app,
	}
}

func (q *HistorySyncWorker) ProcessQueue() {
	queue := queues.NewHistorySyncQueue(q.app)

	defer q.app.Wg.Done()
	for {
		select {
		case <-q.app.StopCh:
			return
		default:
			data, err := queue.Dequeue()
			if err == nil {
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
			account, err := q.accountRepo.GetAccountByInstanceID(data.InstanceID)
			if err != nil {
				fmt.Println(err)
				continue
			}

			if !account.WasSynced {
				err = q.accountRepo.UpdateAccount(account.InstanceID, map[string]interface{}{
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

				count, err := q.messageRepo.CountMessages(account.User, chatJID.User)
				if err != nil {
					fmt.Println(err)
					continue
				}

				if count > int64(q.app.Config.MessageLimit) {
					continue
				}

				var historySyncMsgs = conv.GetMessages()
				sort.Slice(historySyncMsgs, func(i, j int) bool {
					message1, _ := instance.ParseWebMessage(chatJID, historySyncMsgs[i].GetMessage())
					message2, _ := instance.ParseWebMessage(chatJID, historySyncMsgs[j].GetMessage())
					return message1.Info.Timestamp.After(message2.Info.Timestamp)
				})

				var limit = utils.Min(q.app.Config.MessageLimit, len(historySyncMsgs))
				var slice = historySyncMsgs[:limit]
				for _, historySyncMsg := range slice {
					eventMessage, _ := instance.ParseWebMessage(chatJID, historySyncMsg.GetMessage())
					message := q.messageService.Parse(instance, eventMessage)
					if message != nil {
						messages = append(messages, *message)
					}
				}
			}

			err = q.messageRepo.CreateMessages(&messages)
			if err != nil {
				fmt.Println(err)
			}
			continue
		}
	}
}
