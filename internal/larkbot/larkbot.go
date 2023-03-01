package larkbot

import (
	"context"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/larksuite/oapi-sdk-gin"
	"github.com/larksuite/oapi-sdk-go/v3"
	"github.com/larksuite/oapi-sdk-go/v3/event/dispatcher"
	"github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
)

type Config struct {
	AppID             string
	AppSecret         string
	VerificationToken string
	EventEncryptKey   string
}

type MessageType string

const (
	PrivateChat MessageType = "PrivateChat"
	GroupChat   MessageType = "GroupChat"
)

type Message struct {
	ID       string
	Type     MessageType
	SenderID string
	Content  string
}

type Bot struct {
	conf Config
	cli  *lark.Client
}

func New(c Config) *Bot {
	return &Bot{
		conf: c,
		cli:  lark.NewClient(c.AppID, c.AppSecret),
	}
}

func (b *Bot) Run(handlerFunc func(msg Message)) error {
	r := gin.Default()
	handler := dispatcher.NewEventDispatcher(b.conf.VerificationToken, b.conf.EventEncryptKey).
		OnP2MessageReceiveV1(func(ctx context.Context, e *larkim.P2MessageReceiveV1) error {
			go func() {
				msg := Message{
					ID:       *e.Event.Message.MessageId,
					Type:     PrivateChat,
					SenderID: *e.Event.Sender.SenderId.OpenId,
					Content:  *e.Event.Message.Content,
				}
				if *e.Event.Message.ChatType != "p2p" {
					msg.Type = GroupChat
				}
				handlerFunc(msg)
			}()
			return nil
		})

	r.POST("/webhook/event", sdkginext.NewEventHandlerFunc(handler))
	return r.Run()
}

type msgContent struct {
	Text string `json:"text"`
}

func (b *Bot) Reply(messageID, content string) error {
	msg := msgContent{Text: content}
	msgByte, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	_, err = b.cli.Im.Message.Reply(context.Background(), larkim.NewReplyMessageReqBuilder().
		MessageId(messageID).
		Body(larkim.NewReplyMessageReqBodyBuilder().
			MsgType(larkim.MsgTypeText).
			Content(string(msgByte)).
			Build()).
		Build(),
	)
	return err
}
