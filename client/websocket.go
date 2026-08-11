package client

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gorilla/websocket"
	"github.com/slipe-fun/bloom-kit/domain"
	"github.com/slipe-fun/bloom-kit/mappers"
	"github.com/slipe-fun/skid-v4/pkg/identity"
)

type WsEvent struct {
	Type string `json:"type"`
	Data []byte `json:"data"`
}

type ChatNewEvent struct {
	domain.Chat
	Type              string                    `json:"type"`
	ChatType          string                    `json:"chat_type"`
	UserID            string                    `json:"user_id"`
	Role              *string                   `json:"role,omitempty"`
	InvitedBy         *domain.User              `json:"invited_by,omitempty"`
	Handshake         *domain.Handshake         `json:"handshake,omitempty"`
	EncryptedGroupKey *domain.EncryptedGroupKey `json:"encrypted_group_key,omitempty"`
}

type MessageNewEvent struct {
	Type    string             `json:"type"`
	ID      int                `json:"id"`
	UserID  string             `json:"user_id"`
	ReplyTo *domain.RawMessage `json:"reply_to,omitempty"`
	*domain.RawMessageWithReply
}

func (m *MessageNewEvent) UnmarshalJSON(data []byte) error {
	var helper struct {
		Type    string             `json:"type"`
		ID      int                `json:"id"`
		UserID  string             `json:"user_id"`
		ReplyTo *domain.RawMessage `json:"reply_to,omitempty"`
	}

	if err := json.Unmarshal(data, &helper); err != nil {
		return err
	}

	var raw domain.RawMessageWithReply
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	m.Type = helper.Type
	m.ID = helper.ID
	m.UserID = helper.UserID
	m.ReplyTo = helper.ReplyTo
	m.RawMessageWithReply = &raw

	return nil
}

func (c *BloomClient) startWebSocket(ctx context.Context, wsURL string) {
	go func() {
		backoff := 0 * time.Second

		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(backoff):
				err := c.connectAndListen(ctx, fmt.Sprintf("%s?token=%s", wsURL, c.credentials.Token))
				if err != nil {
					if backoff < 30*time.Second {
						backoff *= 2
					}
					continue
				}
				backoff = time.Second
			}
		}
	}()
}

func (c *BloomClient) connectAndListen(ctx context.Context, url string) error {
	conn, _, err := websocket.DefaultDialer.DialContext(ctx, url, nil)
	if err != nil {
		return fmt.Errorf("websocket dial error: %w", err)
	}
	defer conn.Close()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			_, message, err := conn.ReadMessage()
			if err != nil {
				return fmt.Errorf("websocket read error: %w", err)
			}

			var event WsEvent
			if err := json.Unmarshal(message, &event); err != nil {
				continue
			}

			switch event.Type {
			case "chat.new":
				var event ChatNewEvent
				if err := json.Unmarshal(message, &event); err != nil {
					continue
				}

				c.handleNewChatEvent(&event)
			case "message.new":
				var event MessageNewEvent
				if err := json.Unmarshal(message, &event); err != nil {
					fmt.Println(err)
					continue
				}

				c.handleNewMessageEvent(&event)
			}
		}
	}
}

func (c *BloomClient) handleNewChatEvent(chatEvent *ChatNewEvent) {
	chat := chatEvent.Chat

	creds, err := c.loadCredentials()
	if err != nil {
		return
	}

	var me *domain.User
	if err = json.Unmarshal(c.credentials.UserJSON, &me); err != nil {
		return
	}

	userIdentity := mappers.ConvertUserToIdentity(me)
	secretKeys, err := mappers.UnmapSecretKeys(creds.SecretKeys)
	if err != nil {
		return
	}
	defer secretKeys.Wipe()

	switch chatEvent.ChatType {
	case "private":
		recipient := getChatOtherMember(&chat, c.credentials.UserID)
		if recipient == nil {
			return
		}

		recipientIdentity := mappers.ConvertUserToIdentity(recipient)

		handshakePayload, err := mappers.DecodeHandshake(chat.Handshake)
		if err != nil {
			return
		}

		chatKey, syncKey, err := identity.FinalizeKeyExchange(handshakePayload, userIdentity, secretKeys, recipientIdentity, nil, true)
		if err != nil {
			chatKey, syncKey, err = identity.FinalizeKeyExchange(handshakePayload, recipientIdentity, nil, userIdentity, secretKeys, false)
			if err != nil {
				return
			}
		}

		err = c.database.SaveChat(chat, chatKey, syncKey)
		if err != nil {
			return
		}

		_ = c.database.SaveUser(recipient)
	case "group":
		// recipientIdentity := mappers.ConvertUserToIdentity(chatEvent.InvitedBy)

		// handshakePayload, err := mappers.DecodeHandshake(chatEvent.Handshake)
		// if err != nil {
		// 	return
		// }

		// chatKey, syncKey, err := identity.FinalizeKeyExchange(handshakePayload, userIdentity, secretKeys, recipientIdentity, nil, true)
	}

	c.notifyChatsUpdated()
}

func (c *BloomClient) handleNewMessageEvent(messageEvent *MessageNewEvent) {
	chat, err := c.database.GetChat(messageEvent.ChatID)
	if err != nil {
		return
	}

	recipient := getChatOtherMember(&domain.Chat{
		RawChat: chat.RawChat,
	}, c.credentials.UserID)
	recipientIdentity := mappers.ConvertUserToIdentity(recipient)

	user := getChatOtherMember(&domain.Chat{
		RawChat: chat.RawChat,
	}, recipient.ID)
	userIdentity := mappers.ConvertUserToIdentity(user)

	decryptedMessage, err := c.messageManager.DecryptMessage(messageEvent.RawMessageWithReply, chat.ChatKey, chat.SyncKey, userIdentity, recipientIdentity)
	if err != nil {
		return
	}
	if decryptedMessage == nil {
		return
	}

	var replyToMessage *domain.MessageWithDecryptedData
	if messageEvent.ReplyTo != nil {
		decryptedReply, err := c.messageManager.DecryptMessage(&domain.RawMessageWithReply{
			RawMessage: *messageEvent.ReplyTo,
		}, chat.ChatKey, chat.SyncKey, userIdentity, recipientIdentity)
		if err != nil {
			return
		}
		if decryptedReply == nil {
			return
		}

		replyToMessage = &domain.MessageWithDecryptedData{
			RawMessageWithReply: domain.RawMessageWithReply{
				RawMessage: *messageEvent.ReplyTo,
			},
			Message: *decryptedReply,
		}
	}

	msg := domain.Message{
		MessageWithDecryptedData: domain.MessageWithDecryptedData{
			RawMessageWithReply: *messageEvent.RawMessageWithReply,
			Message:             *decryptedMessage,
		},
		ReplyToMessage: replyToMessage,
	}

	err = c.database.SaveMessage(&msg)
	if err != nil {
		return
	}

	decryptedWithReply := mappers.MapDomainMessageToDecrypted(&msg)
	messageJSON, err := json.Marshal(decryptedWithReply)
	if err != nil {
		return
	}

	c.listenerMu.RLock()
	messagesListener := c.messagesListener
	c.listenerMu.RUnlock()

	if messagesListener != nil {
		messagesListener.OnNewMessage(messageJSON)
	}

	c.notifyChatsUpdated()
}
