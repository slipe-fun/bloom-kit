package domain

import "time"

type RawMessage struct {
	ID         int        `json:"id"`
	Ciphertext string     `json:"ciphertext"`
	Nonce      string     `json:"nonce"`
	Salt       string     `json:"salt"`
	ChatID     int        `json:"chat_id"`
	Seen       *time.Time `json:"seen,omitempty"`
	ReplyTo    *int       `json:"reply_to,omitempty"`
}

type Message struct {
	RawMessage
	ReplyToMessage *RawMessage `json:"reply_to,omitempty"`
}
