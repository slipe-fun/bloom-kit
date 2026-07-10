package mappers

import (
	"encoding/base64"

	"github.com/slipe-fun/bloom-kit/domain"
	"github.com/slipe-fun/skid-v4/pkg/messages"
)

func ConvertRawMessageToEncryptedMessage(ciphertext, nonce, salt string) (*messages.EncryptedMessage, error) {
	ciphertextBytes, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return nil, err
	}

	nonceBytes, err := base64.StdEncoding.DecodeString(nonce)
	if err != nil {
		return nil, err
	}

	saltBytes, err := base64.StdEncoding.DecodeString(salt)
	if err != nil {
		return nil, err
	}

	return &messages.EncryptedMessage{
		Ciphertext: ciphertextBytes,
		Nonce:      nonceBytes,
		Salt:       saltBytes,
	}, nil
}

func MapDomainMessageToDecrypted(msg *domain.Message) domain.DecryptedMessageWithReply {
	decrypted := domain.DecryptedMessageWithReply{
		DecryptedMessage: domain.DecryptedMessage{
			ID:        msg.ID,
			Content:   string(msg.Content),
			AuthorID:  msg.AuthorID,
			Timestamp: msg.Timestamp,
			Seen:      msg.Seen,
		},
	}

	if msg.ReplyToMessage != nil {
		decrypted.ReplyTo = &domain.DecryptedMessage{
			ID:        msg.ReplyToMessage.ID,
			Content:   string(msg.ReplyToMessage.Content),
			AuthorID:  msg.ReplyToMessage.AuthorID,
			Timestamp: msg.ReplyToMessage.Timestamp,
			Seen:      msg.ReplyToMessage.Seen,
		}
	}

	return decrypted
}
