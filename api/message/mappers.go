package message

import (
	"encoding/base64"

	"github.com/slipe-fun/bloom-kit/domain"
	"github.com/slipe-fun/skid-v4/pkg/messages"
)

func NewSendMessageRequest(chatID int, replyToID *int, encryptedMessage messages.EncryptedMessage) *domain.SendMessageRequest {
	return &domain.SendMessageRequest{
		Ciphertext: base64.StdEncoding.EncodeToString(encryptedMessage.Ciphertext),
		Nonce:      base64.StdEncoding.EncodeToString(encryptedMessage.Nonce),
		Salt:       base64.StdEncoding.EncodeToString(encryptedMessage.Salt),
		ChatID:     chatID,
		ReplyTo:    replyToID,
	}
}
