package services

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"
	"zapmeow/configs"
	"zapmeow/models"
	"zapmeow/queues"
	"zapmeow/utils"

	"github.com/vincent-petithory/dataurl"
	"go.mau.fi/whatsmeow"
	waProto "go.mau.fi/whatsmeow/binary/proto"
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
	proxyService   ProxyService
}

type ContactInfo struct {
	Phone   string
	Name    string
	Status  string
	Picture string
}

type SendMessageResponse struct {
	ID        string
	Sender    types.JID
	Timestamp time.Time
}

type DownloadedMedia struct {
	Data     []byte
	Type     MediaType
	Mimetype string
}

type ParsedEventMessage struct {
	InstanceID string
	Body       string
	SenderJID  string
	ChatJID    string
	MessageID  string
	Timestamp  time.Time
	FromMe     bool
	MediaType  *MediaType
	Media      *[]byte
	Mimetype   *string
}

type MediaType int

const (
	Audio MediaType = iota
	Image
	Document
	Sticker
	Video
)

func (m MediaType) String() string {
	switch m {
	case Audio:
		return "audio"
	case Document:
		return "document"
	case Sticker:
		return "sticker"
	case Video:
		return "video"
	case Image:
		return "image"
	}
	return "unknown"
}

type WppService interface {
	GetInstance(instanceID string) (*configs.Instance, error)
	GetAuthenticatedInstance(instanceID string) (*configs.Instance, error)
	GetContactInfo(instanceID string, jid types.JID) (*ContactInfo, error)
	SendMessage(instanceID string, jid types.JID, message *waProto.Message) (SendMessageResponse, error)
	SendTextMessage(instanceID string, jid types.JID, message string) (SendMessageResponse, error)
	SendAudioMessage(instanceID string, jid types.JID, audio *dataurl.DataURL, mimitype string) (SendMessageResponse, error)
	SendImageMessage(instanceID string, jid types.JID, image *dataurl.DataURL, mimitype string) (SendMessageResponse, error)
	ParseEventMessage(instance *configs.Instance, message *events.Message) (ParsedEventMessage, error)
	Logout(instanceID string) error
	destroyInstance(instanceID string) error
	getTextMessage(message *waProto.Message) string
	downloadMedia(instance *configs.Instance, message *waProto.Message) (*DownloadedMedia, error)
	uploadMedia(instanceID string, media *dataurl.DataURL, mediaType MediaType) (*whatsmeow.UploadResponse, error)
}

func NewWppService(
	app *configs.ZapMeow,
	messageService MessageService,
	accountService AccountService,
	proxyService ProxyService,
) *wppService {
	return &wppService{
		app:            app,
		messageService: messageService,
		accountService: accountService,
		proxyService:   proxyService,
	}
}

func (w *wppService) GetInstance(instanceID string) (*configs.Instance, error) {
	instance := w.app.LoadInstance(instanceID)

	if instance != nil {
		return instance, nil
	}

	client, err := w.getClient(instanceID)
	if err != nil {
		return nil, err
	}

	w.app.StoreInstance(instanceID, &configs.Instance{
		ID:      instanceID,
		Client:  client,
		ProxyID: 0,
	})

	instance = w.app.LoadInstance(instanceID)

	instance.Client.AddEventHandler(func(evt interface{}) {
		w.eventHandler(instanceID, evt)
	})

	if instance.Client.Store.ID == nil {
		go w.qrcode(instanceID)
	} else {
		err := instance.Client.Connect()
		if err != nil {
			return nil, err
		}

		if !instance.Client.WaitForConnection(5 * time.Second) {
			return nil, errors.New("websocket didn't reconnect within 5 seconds of failed")
		}
	}

	return instance, nil
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

func (w *wppService) SendTextMessage(instanceID string, jid types.JID, text string) (SendMessageResponse, error) {
	message := &waProto.Message{
		ExtendedTextMessage: &waProto.ExtendedTextMessage{
			Text: &text,
		},
	}

	return w.SendMessage(instanceID, jid, message)
}

func (w *wppService) SendAudioMessage(instanceID string, jid types.JID, audioURL *dataurl.DataURL, mimitype string) (SendMessageResponse, error) {
	uploaded, err := w.uploadMedia(instanceID, audioURL, Audio)
	if err != nil {
		return SendMessageResponse{}, err
	}

	message := &waProto.Message{
		AudioMessage: &waProto.AudioMessage{
			Ptt:           proto.Bool(true),
			Url:           proto.String(uploaded.URL),
			DirectPath:    proto.String(uploaded.DirectPath),
			MediaKey:      uploaded.MediaKey,
			Mimetype:      proto.String(mimitype),
			FileEncSha256: uploaded.FileEncSHA256,
			FileSha256:    uploaded.FileSHA256,
			FileLength:    proto.Uint64(uint64(len(audioURL.Data))),
		},
	}

	return w.SendMessage(instanceID, jid, message)
}

func (w *wppService) SendImageMessage(instanceID string, jid types.JID, imageURL *dataurl.DataURL, mimitype string) (SendMessageResponse, error) {
	uploaded, err := w.uploadMedia(instanceID, imageURL, Image)
	if err != nil {
		return SendMessageResponse{}, err
	}

	message := &waProto.Message{
		ImageMessage: &waProto.ImageMessage{
			Url:           proto.String(uploaded.URL),
			DirectPath:    proto.String(uploaded.DirectPath),
			MediaKey:      uploaded.MediaKey,
			Mimetype:      proto.String(mimitype),
			FileEncSha256: uploaded.FileEncSHA256,
			FileSha256:    uploaded.FileSHA256,
			FileLength:    proto.Uint64(uint64(len(imageURL.Data))),
		},
	}

	return w.SendMessage(instanceID, jid, message)
}

func (w *wppService) SendMessage(instanceID string, jid types.JID, message *waProto.Message) (SendMessageResponse, error) {
	instance, err := w.GetAuthenticatedInstance(instanceID)
	if err != nil {
		return SendMessageResponse{}, err
	}

	proxy, _ := w.proxyService.GetProxy(instance.ProxyID)

	if proxy.Ranking == 0 {
		proxy, err := w.proxyService.GetProxyWithHighestRanking()
		if err != nil {
			return SendMessageResponse{}, err
		}
		instance.ProxyID = proxy.ID

		instance.Client.SetProxy(http.ProxyURL(&url.URL{
			Scheme: proxy.Scheme,
			Host:   proxy.Ip + ":" + proxy.Port,
		}))
		instance.Client.Disconnect()
		instance.Client.Connect()
	}

	resp, err := instance.Client.SendMessage(context.Background(), jid, message)
	if err != nil {
		return SendMessageResponse{}, err
	}

	return SendMessageResponse{
		ID:        resp.ID,
		Sender:    *instance.Client.Store.ID,
		Timestamp: resp.Timestamp,
	}, nil
}

func (w *wppService) ParseEventMessage(instance *configs.Instance, message *events.Message) (ParsedEventMessage, error) {
	media, err := w.downloadMedia(
		instance,
		message.Message,
	)

	if err != nil && media == nil {
		return ParsedEventMessage{}, err
	}

	text := w.getTextMessage(message.Message)
	base := ParsedEventMessage{
		InstanceID: instance.ID,
		Body:       text,
		MessageID:  message.Info.ID,
		ChatJID:    message.Info.Chat.User,
		SenderJID:  message.Info.Sender.User,
		FromMe:     message.Info.MessageSource.IsFromMe,
		Timestamp:  message.Info.Timestamp,
	}

	if media != nil && err == nil {
		base.MediaType = &media.Type
		base.Mimetype = &media.Mimetype
		base.Media = &media.Data
		return base, nil
	}

	return base, nil
}

func (w *wppService) uploadMedia(instanceID string, media *dataurl.DataURL, mediaType MediaType) (*whatsmeow.UploadResponse, error) {
	instance, err := w.GetAuthenticatedInstance(instanceID)
	if err != nil {
		return nil, err
	}

	var mType whatsmeow.MediaType
	switch mediaType {
	case Image:
		mType = whatsmeow.MediaImage
	case Audio:
		mType = whatsmeow.MediaAudio
	default:
		return nil, errors.New("unknown media type")
	}

	uploaded, err := instance.Client.Upload(context.Background(), media.Data, mType)
	if err != nil {
		return nil, err
	}

	return &uploaded, nil
}

func (m *wppService) downloadMedia(instance *configs.Instance, message *waProto.Message) (*DownloadedMedia, error) {
	dirPath := utils.MakeAccountStoragePath(instance.ID)
	err := os.MkdirAll(dirPath, 0751)
	if err != nil {
		return nil, err
	}

	document := message.GetDocumentMessage()
	if document != nil {
		data, err := instance.Client.Download(document)
		if err != nil {
			return &DownloadedMedia{Type: Document}, err
		}

		return &DownloadedMedia{
			Data:     data,
			Type:     Document,
			Mimetype: document.GetMimetype(),
		}, nil
	}

	audio := message.GetAudioMessage()
	if audio != nil {
		data, err := instance.Client.Download(audio)
		if err != nil {
			return &DownloadedMedia{Type: Audio}, err
		}

		return &DownloadedMedia{
			Data:     data,
			Type:     Audio,
			Mimetype: audio.GetMimetype(),
		}, nil
	}

	image := message.GetImageMessage()
	if image != nil {
		data, err := instance.Client.Download(image)
		if err != nil {
			return &DownloadedMedia{Type: Image}, err
		}

		return &DownloadedMedia{
			Data:     data,
			Type:     Image,
			Mimetype: image.GetMimetype(),
		}, nil
	}

	sticker := message.GetStickerMessage()
	if sticker != nil {
		data, err := instance.Client.Download(sticker)
		if err != nil {
			return &DownloadedMedia{Type: Sticker}, err
		}

		return &DownloadedMedia{
			Data:     data,
			Type:     Sticker,
			Mimetype: sticker.GetMimetype(),
		}, nil
	}

	video := message.GetVideoMessage()
	if video != nil {
		data, err := instance.Client.Download(video)
		if err != nil {
			return &DownloadedMedia{Type: Video}, err
		}

		return &DownloadedMedia{
			Data:     data,
			Type:     Video,
			Mimetype: video.GetMimetype(),
		}, nil
	}

	return nil, nil
}

func (m *wppService) getTextMessage(message *waProto.Message) string {
	extendedTextMessage := message.GetExtendedTextMessage()
	if extendedTextMessage != nil {
		return *extendedTextMessage.Text
	}
	return message.GetConversation()
}

func (w *wppService) destroyInstance(instanceID string) error {
	instance, err := w.GetInstance(instanceID)
	if err != nil {
		return err
	}

	err = w.accountService.DeleteAccountMessages(instanceID)
	if err != nil {
		return err
	}

	instance.Client.Disconnect()
	w.app.DeleteInstance(instanceID)

	return nil
}

func (w *wppService) Logout(instanceID string) error {
	instance, err := w.GetAuthenticatedInstance(instanceID)
	if err != nil {
		return err
	}

	err = instance.Client.Logout()
	if err != nil {
		return err
	}

	err = w.accountService.UpdateAccount(instanceID, map[string]interface{}{
		"Status": "UNPAIRED",
	})
	if err != nil {
		return err
	}

	return w.destroyInstance(instanceID)
}

func (w *wppService) GetContactInfo(instanceID string, jid types.JID) (*ContactInfo, error) {
	instance, err := w.GetAuthenticatedInstance(instanceID)
	if err != nil {
		return nil, err
	}

	userInfo, err := instance.Client.GetUserInfo([]types.JID{jid})
	if err != nil {
		return nil, err
	}

	ctInfo, err := instance.Client.Store.Contacts.GetContact(jid)
	if err != nil {
		return nil, err
	}

	profilePictureInfo, err := instance.Client.GetProfilePictureInfo(
		jid,
		&whatsmeow.GetProfilePictureParams{},
	)

	profilePictureURL := ""
	if profilePictureInfo != nil {
		profilePictureURL = profilePictureInfo.URL
	}

	return &ContactInfo{
		Phone:   jid.User,
		Name:    ctInfo.PushName,
		Status:  userInfo[jid].Status,
		Picture: profilePictureURL,
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
		return w.createClient(
			w.app.WhatsmeowContainer.NewDevice(),
		), nil
	} else if account.Status != "CONNECTED" {
		return w.createClient(
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
		return w.createClient(
			device,
		), nil
	}

	device = w.app.WhatsmeowContainer.NewDevice()
	return w.createClient(
		device,
	), nil
}

func (w *wppService) qrcode(instanceID string) {
	instance := w.app.LoadInstance(instanceID)
	if instance.Client.Store.ID == nil {
		qrChan, err := instance.Client.GetQRChannel(context.Background())
		if err != nil {
			if !errors.Is(err, whatsmeow.ErrQRStoreContainsID) {
				if w.app.Config.Env != "production" {
					fmt.Println("failed to get qr channel")
				}
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
					return
				case "timeout":
					{
						// w.app.Mutex.Lock()
						// defer w.app.Mutex.Unlock()
						err := w.accountService.UpdateAccount(instanceID, map[string]interface{}{
							"QrCode": "",
							"Status": "TIMEOUT",
						})
						if err != nil {
							fmt.Println("[qrcode]: ", err)
						}

						w.app.DeleteInstance(instanceID)
					}
				case "code":
					{
						// w.app.Mutex.Lock()
						// defer w.app.Mutex.Unlock()
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
	case *events.TempBanReason:
	case *events.TemporaryBan:
		w.handleBan(instanceID)
	}
}

func (w *wppService) handleBan(instanceID string) {
	var instance = w.app.LoadInstance(instanceID)
	proxy, _ := w.proxyService.GetProxy(instance.ProxyID)

	w.proxyService.UpdateProxy(proxy.ID, map[string]interface{}{
		"Ranking": proxy.Ranking - 1,
	})
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
	var instance = w.app.LoadInstance(instanceID)
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
	err := w.destroyInstance(instanceID)
	if err != nil {
		fmt.Println("Error", err)
	}

	err = w.accountService.UpdateAccount(instanceID, map[string]interface{}{
		"Status": "UNPAIRED",
	})
	if err != nil {
		fmt.Println("Error", err)
	}
}

func (w *wppService) handleMessage(instanceId string, evt *events.Message) {
	instance := w.app.LoadInstance(instanceId)
	parsedEventMessage, err := w.ParseEventMessage(instance, evt)

	if err != nil {
		return
	}

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

	err = w.messageService.CreateMessage(&message)
	if err != nil {
		fmt.Println(err)
	}

	body := map[string]interface{}{
		"InstanceId": instanceId,
		"Message":    w.messageService.ToJSON(message),
	}

	err = utils.Request(w.app.Config.WebhookURL, body)

	if err != nil {
		fmt.Println("Error when send request:", err)
	}
}

func (w *wppService) createClient(deviceStore *store.Device) *whatsmeow.Client {
	var client *whatsmeow.Client
	if w.app.Config.Env == "production" {
		client = whatsmeow.NewClient(deviceStore, nil)
	} else {
		log := waLog.Stdout("Client", "DEBUG", true)
		client = whatsmeow.NewClient(deviceStore, log)
	}

	proxy, _ := w.proxyService.GetProxyWithHighestRanking()
	fmt.Println("proxy =>>>", proxy)
	client.SetProxy(http.ProxyURL(&url.URL{
		Scheme: proxy.Scheme,
		Host:   proxy.Ip + ":" + proxy.Port,
	}))

	return client
}
