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
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

type ChatNewEvent struct {
	Type   string `json:"type"`
	UserID string `json:"user_id"`
	domain.Chat
}

func (c *BloomClient) startWebSocket(ctx context.Context, wsURL string) {
	go func() {
		backoff := time.Second

		for {
			select {
			case <-ctx.Done():
				return
			default:
				err := c.connectAndListen(ctx, fmt.Sprintf("%s?token=%s", wsURL, c.credentials.Token))
				if err != nil {
					time.Sleep(backoff)
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

	recipient := getChatOtherMember(&chat, c.credentials.UserID)
	if recipient == nil {
		return
	}

	me := getChatOtherMember(&chat, recipient.ID)
	if recipient == nil {
		return
	}

	recipientIdentity := mappers.ConvertUserToIdentity(recipient)
	userIdentity := mappers.ConvertUserToIdentity(me)

	secretKeys, err := mappers.UnmapSecretKeys(creds.SecretKeys)
	if err != nil {
		return
	}
	defer secretKeys.Wipe()

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

	c.notifyChatsUpdated()
}
