package mappers

import (
	"encoding/base64"
	"time"

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
	createdAt := msg.CreatedAt
	if createdAt.IsZero() && msg.Timestamp != 0 {
		createdAt = time.Unix(msg.Timestamp, 0).UTC()
	}
	timestamp := msg.Timestamp
	if timestamp == 0 && !createdAt.IsZero() {
		timestamp = createdAt.Unix()
	}

	decrypted := domain.DecryptedMessageWithReply{
		DecryptedMessage: domain.DecryptedMessage{
			ID:        msg.ID,
			Content:   string(msg.Content),
			AuthorID:  msg.AuthorID,
			Timestamp: timestamp,
			CreatedAt: createdAt,
			Seen:      msg.Seen,
		},
	}

	if msg.ReplyToMessage != nil {
		replyCreatedAt := msg.ReplyToMessage.CreatedAt
		if replyCreatedAt.IsZero() && msg.ReplyToMessage.Timestamp != 0 {
			replyCreatedAt = time.Unix(msg.ReplyToMessage.Timestamp, 0).UTC()
		}
		replyTimestamp := msg.ReplyToMessage.Timestamp
		if replyTimestamp == 0 && !replyCreatedAt.IsZero() {
			replyTimestamp = replyCreatedAt.Unix()
		}

		decrypted.ReplyTo = &domain.DecryptedMessage{
			ID:        msg.ReplyToMessage.ID,
			Content:   string(msg.ReplyToMessage.Content),
			AuthorID:  msg.ReplyToMessage.AuthorID,
			Timestamp: replyTimestamp,
			CreatedAt: replyCreatedAt,
			Seen:      msg.ReplyToMessage.Seen,
		}
	}

	return decrypted
}
