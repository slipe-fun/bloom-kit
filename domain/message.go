package domain

import (
	"time"

	"github.com/slipe-fun/skid-v4/pkg/messages"
)

type RawMessage struct {
	ID         int        `json:"id"`
	Ciphertext string     `json:"ciphertext"`
	Nonce      string     `json:"nonce"`
	Salt       string     `json:"salt"`
	ChatID     int        `json:"chat_id"`
	Seen       *time.Time `json:"seen,omitempty"`
	ReplyTo    *int       `json:"reply_to,omitempty"`
}

type RawMessageWithReply struct {
	RawMessage
	ReplyToMessage *RawMessage `json:"reply_to,omitempty"`
}

type MessageWithDecryptedData struct {
	RawMessageWithReply
	messages.Message
}

type Message struct {
	MessageWithDecryptedData
	ReplyToMessage *MessageWithDecryptedData `json:"reply_to,omitempty"`
}

type SendMessageRequest struct {
	Ciphertext string `json:"ciphertext"`
	Nonce      string `json:"nonce"`
	Salt       string `json:"salt"`
	ChatID     int    `json:"chat_id"`
	ReplyTo    *int   `json:"reply_to,omitempty"`
}
