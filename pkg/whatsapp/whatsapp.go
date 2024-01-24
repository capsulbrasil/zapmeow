package whatsapp

import (
	"context"
	"errors"
	"fmt"
	"time"
	"zapmeow/config"
	"zapmeow/pkg/logger"

	"github.com/vincent-petithory/dataurl"
	"go.mau.fi/whatsmeow"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	"go.mau.fi/whatsmeow/store"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	waLog "go.mau.fi/whatsmeow/util/log"
	"google.golang.org/protobuf/proto"
)

type Client = whatsmeow.Client

type JID = types.JID

type Instance struct {
	ID              string
	Client          *Client
	QrCodeRateLimit uint16
}

type Message struct {
	InstanceID string
	Body       string
	SenderJID  string
	ChatJID    string
	MessageID  string
	FromMe     bool
	Timestamp  time.Time
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

type ContactInfo struct {
	Phone   string `json:"phone"`
	Name    string `json:"name"`
	Status  string `json:"status"`
	Picture string `json:"picture"`
}

type MessageResponse struct {
	ID        string
	Sender    JID
	Timestamp time.Time
}

type DownloadResponse struct {
	Data     []byte
	Type     MediaType
	Mimetype string
}

type UploadResponse struct {
	URL           string
	DirectPath    string
	Mimetype      MediaType
	MediaKey      []byte
	FileEncSHA256 []byte
	FileSHA256    []byte
	FileLength    uint64
}

type IsOnWhatsAppResponse struct {
	Query        string `json:"query"`
	Phone        string `json:"phone"`
	IsRegistered bool   `json:"is_registered"`
}

type WhatsApp interface {
	CreateInstance(id string) *Instance
	CreateInstanceFromDevice(id string, jid JID) *Instance
	IsLoggedIn(instance *Instance) bool
	IsConnected(instance *Instance) bool
	Disconnect(instance *Instance)
	Logout(instance *Instance) error
	EventHandler(instance *Instance, handler func(evt interface{}))
	InitInstance(instance *Instance, qrcodeHandler func(evt string, qrcode string, err error)) error
	SendTextMessage(instance *Instance, jid JID, text string) (MessageResponse, error)
	SendAudioMessage(instance *Instance, jid JID, audioURL *dataurl.DataURL, mimitype string) (MessageResponse, error)
	SendImageMessage(instance *Instance, jid JID, imageURL *dataurl.DataURL, mimitype string) (MessageResponse, error)
	SendDocumentMessage(instance *Instance, jid JID, documentURL *dataurl.DataURL, mimitype string, filename string) (MessageResponse, error)
	GetContactInfo(instance *Instance, jid JID) (*ContactInfo, error)
	ParseEventMessage(instance *Instance, message *events.Message) (Message, error)
	IsOnWhatsApp(instance *Instance, phones []string) ([]IsOnWhatsAppResponse, error)
}

type whatsApp struct {
	container *sqlstore.Container
}

func NewWhatsApp(databasePath string) *whatsApp {
	cfg := config.Load()

	var level = "DEBUG"
	if cfg.Environment == config.Production {
		level = "ERROR"
	}
	dbLog := waLog.Stdout("Database", level, true)

	container, err := sqlstore.New("sqlite3", "file:"+databasePath+"?_foreign_keys=on", dbLog)
	if err != nil {
		logger.Fatal(err)
	}
	return &whatsApp{container: container}
}

func (w *whatsApp) CreateInstance(id string) *Instance {
	client := w.createClient(w.container.NewDevice())
	return &Instance{
		ID:              id,
		Client:          client,
		QrCodeRateLimit: 10,
	}
}

func (w *whatsApp) CreateInstanceFromDevice(id string, jid JID) *Instance {
	device, _ := w.container.GetDevice(JID{
		User:   jid.User,
		Agent:  jid.Agent,
		Device: jid.Device,
		Server: jid.Server,
		AD:     jid.AD,
	})
	if device != nil {
		client := w.createClient(device)
		return &Instance{
			ID:              id,
			Client:          client,
			QrCodeRateLimit: 10,
		}
	}
	return w.CreateInstance(id)
}

func (w *whatsApp) IsLoggedIn(instance *Instance) bool {
	return instance.Client.IsLoggedIn()
}

func (w *whatsApp) IsConnected(instance *Instance) bool {
	return instance.Client.IsConnected()
}

func (w *whatsApp) Disconnect(instance *Instance) {
	instance.Client.Disconnect()
}

func (w *whatsApp) Connect(instance *Instance) {
	instance.Client.Disconnect()
}

func (w *whatsApp) Logout(instance *Instance) error {
	return instance.Client.Logout()
}

func (w *whatsApp) EventHandler(instance *Instance, handler func(evt interface{})) {
	instance.Client.AddEventHandler(handler)
}

func (w *whatsApp) InitInstance(instance *Instance, qrcodeHandler func(evt string, qrcode string, err error)) error {
	if instance.Client.Store.ID == nil {
		go w.generateQrcode(instance, qrcodeHandler)
	} else {
		err := instance.Client.Connect()
		if err != nil {
			return err
		}

		if !instance.Client.WaitForConnection(5 * time.Second) {
			return errors.New("websocket didn't reconnect within 5 seconds of failed")
		}
	}

	return nil
}

func (w *whatsApp) SendTextMessage(instance *Instance, jid JID, text string) (MessageResponse, error) {
	message := &waProto.Message{
		ExtendedTextMessage: &waProto.ExtendedTextMessage{
			Text: &text,
		},
	}
	return w.sendMessage(instance, jid, message)
}

func (w *whatsApp) SendAudioMessage(instance *Instance, jid JID, audioURL *dataurl.DataURL, mimitype string) (MessageResponse, error) {
	uploaded, err := w.uploadMedia(instance, audioURL, Audio)
	if err != nil {
		return MessageResponse{}, err
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
	return w.sendMessage(instance, jid, message)
}

func (w *whatsApp) SendImageMessage(instance *Instance, jid JID, imageURL *dataurl.DataURL, mimitype string) (MessageResponse, error) {
	uploaded, err := w.uploadMedia(instance, imageURL, Image)
	if err != nil {
		return MessageResponse{}, err
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
	return w.sendMessage(instance, jid, message)
}

func (w *whatsApp) SendDocumentMessage(
	instance *Instance, jid JID, documentURL *dataurl.DataURL, mimitype string, filename string) (MessageResponse, error) {
	uploaded, err := w.uploadMedia(instance, documentURL, Document)
	if err != nil {
		return MessageResponse{}, err
	}

	message := &waProto.Message{
		DocumentMessage: &waProto.DocumentMessage{
			Url:           proto.String(uploaded.URL),
			FileName:      &filename,
			DirectPath:    proto.String(uploaded.DirectPath),
			MediaKey:      uploaded.MediaKey,
			Mimetype:      proto.String(mimitype),
			FileEncSha256: uploaded.FileEncSHA256,
			FileSha256:    uploaded.FileSHA256,
			FileLength:    proto.Uint64(uint64(len(documentURL.Data))),
		},
	}
	return w.sendMessage(instance, jid, message)
}

func (w *whatsApp) IsOnWhatsApp(instance *Instance, phones []string) ([]IsOnWhatsAppResponse, error) {
	isOnWhatsAppResponse, err := instance.Client.IsOnWhatsApp(phones)
	if err != nil {
		return nil, err
	}

	data := make([]IsOnWhatsAppResponse, len(isOnWhatsAppResponse))
	for _, resp := range isOnWhatsAppResponse {
		data = append(data, IsOnWhatsAppResponse{
			Query:        resp.Query,
			IsRegistered: resp.IsIn,
			Phone:        resp.JID.User,
		})
	}

	return data, nil
}

func (w *whatsApp) sendMessage(instance *Instance, jid JID, message *waProto.Message) (MessageResponse, error) {
	resp, err := instance.Client.SendMessage(context.Background(), jid, message)
	if err != nil {
		return MessageResponse{}, err
	}

	return MessageResponse{
		ID:        resp.ID,
		Sender:    *instance.Client.Store.ID,
		Timestamp: resp.Timestamp,
	}, nil
}

func (w *whatsApp) GetContactInfo(instance *Instance, jid JID) (*ContactInfo, error) {
	userInfo, err := instance.Client.GetUserInfo([]JID{jid})
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

	profilePictureURL := ""
	if profilePictureInfo != nil {
		profilePictureURL = profilePictureInfo.URL
	}

	return &ContactInfo{
		Phone:   jid.User,
		Name:    contactInfo.PushName,
		Status:  userInfo[jid].Status,
		Picture: profilePictureURL,
	}, nil
}

func (w *whatsApp) ParseEventMessage(instance *Instance, message *events.Message) (Message, error) {
	media, err := w.downloadMedia(
		instance,
		message.Message,
	)

	if err != nil && media == nil {
		return Message{}, err
	}

	text := w.getTextMessage(message.Message)
	base := Message{
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

func (w *whatsApp) createClient(deviceStore *store.Device) *whatsmeow.Client {
	cfg := config.Load()

	var level = "DEBUG"
	if cfg.Environment == config.Production {
		level = "ERROR"
	}
	log := waLog.Stdout("Client", level, true)
	return whatsmeow.NewClient(deviceStore, log)
}

func (w *whatsApp) uploadMedia(instance *Instance, media *dataurl.DataURL, mediaType MediaType) (*UploadResponse, error) {
	var mType whatsmeow.MediaType
	switch mediaType {
	case Image:
		mType = whatsmeow.MediaImage
	case Audio:
		mType = whatsmeow.MediaAudio
	case Document:
		mType = whatsmeow.MediaDocument
	default:
		return nil, errors.New("unknown media type")
	}

	uploaded, err := instance.Client.Upload(context.Background(), media.Data, mType)
	if err != nil {
		return nil, err
	}

	return &UploadResponse{
		URL:           uploaded.URL,
		Mimetype:      mediaType,
		DirectPath:    uploaded.DirectPath,
		MediaKey:      uploaded.MediaKey,
		FileEncSHA256: uploaded.FileEncSHA256,
		FileSHA256:    uploaded.FileSHA256,
		FileLength:    uploaded.FileLength,
	}, nil
}

func (w *whatsApp) downloadMedia(instance *Instance, message *waProto.Message) (*DownloadResponse, error) {
	document := message.GetDocumentMessage()
	if document != nil {
		data, err := instance.Client.Download(document)
		if err != nil {
			return &DownloadResponse{Type: Document}, err
		}

		return &DownloadResponse{
			Data:     data,
			Type:     Document,
			Mimetype: document.GetMimetype(),
		}, nil
	}

	audio := message.GetAudioMessage()
	if audio != nil {
		data, err := instance.Client.Download(audio)
		if err != nil {
			return &DownloadResponse{Type: Audio}, err
		}

		return &DownloadResponse{
			Data:     data,
			Type:     Audio,
			Mimetype: audio.GetMimetype(),
		}, nil
	}

	image := message.GetImageMessage()
	if image != nil {
		data, err := instance.Client.Download(image)
		if err != nil {
			return &DownloadResponse{Type: Image}, err
		}

		return &DownloadResponse{
			Data:     data,
			Type:     Image,
			Mimetype: image.GetMimetype(),
		}, nil
	}

	sticker := message.GetStickerMessage()
	if sticker != nil {
		data, err := instance.Client.Download(sticker)
		if err != nil {
			return &DownloadResponse{Type: Sticker}, err
		}

		return &DownloadResponse{
			Data:     data,
			Type:     Sticker,
			Mimetype: sticker.GetMimetype(),
		}, nil
	}

	video := message.GetVideoMessage()
	if video != nil {
		data, err := instance.Client.Download(video)
		if err != nil {
			return &DownloadResponse{Type: Video}, err
		}

		return &DownloadResponse{
			Data:     data,
			Type:     Video,
			Mimetype: video.GetMimetype(),
		}, nil
	}

	return nil, nil
}

func (w *whatsApp) getTextMessage(message *waProto.Message) string {
	extendedTextMessage := message.GetExtendedTextMessage()
	if extendedTextMessage != nil {
		return *extendedTextMessage.Text
	}
	return message.GetConversation()
}

func (w *whatsApp) generateQrcode(instance *Instance, qrcodeHandler func(evt string, qrcode string, err error)) {
	qrChan, err := instance.Client.GetQRChannel(context.Background())
	if err != nil {
		if !errors.Is(err, whatsmeow.ErrQRStoreContainsID) {
			errMessage := fmt.Sprintf("Failed to get qr channel. %s", err)
			qrcodeHandler("error", "", errors.New(errMessage))
		}
	} else {
		err = instance.Client.Connect()
		if err != nil {
			errMessage := fmt.Sprintf("Failed to connect client to WhatsApp websocket. %s", err)
			qrcodeHandler("error", "", errors.New(errMessage))
		} else {
			for evt := range qrChan {
				if instance.QrCodeRateLimit == 0 {
					qrcodeHandler("rate-limit", "", nil)
					return
				}

				switch evt.Event {
				case "code":
					instance.QrCodeRateLimit -= 1
					qrcodeHandler("code", evt.Code, nil)
				default:
					qrcodeHandler(evt.Event, "", evt.Error)
				}
			}
		}
	}
}
