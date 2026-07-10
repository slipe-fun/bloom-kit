package message

import (
	"context"
	"errors"

	"github.com/slipe-fun/bloom-kit/domain"
	"github.com/slipe-fun/skid-v4/pkg/identity"
)

type Direction string

const (
	DirectionBefore Direction = "before"
	DirectionAfter  Direction = "after"
)

func (m *MessageManager) GetMessages(
	ctx context.Context,
	chatID int,
	messageID int,
	direction Direction,
	encryptionKey []byte,
	syncKey []byte,
	userIdentity *identity.User,
	recipientIdentity *identity.User,
) ([]domain.Message, error) {
	var messages *[]domain.RawMessageWithReply
	var err error
	if direction == DirectionAfter {
		messages, err = m.messageClient.GetMessagesAfter(ctx, chatID, messageID)
		if err != nil {
			return nil, err
		}
	} else if direction == DirectionBefore {
		messages, err = m.messageClient.GetMessagesBefore(ctx, chatID, messageID)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, errors.New("unknown direction")
	}

	if messages == nil {
		return []domain.Message{}, nil
	}

	var result []domain.Message
	for i := range *messages {
		msg := &(*messages)[i]

		decryptedMessage, err := m.DecryptMessage(msg, encryptionKey, syncKey, userIdentity, recipientIdentity)
		if err != nil {
			return nil, err
		}
		if decryptedMessage == nil {
			return nil, errors.New("decrypted message is nil")
		}

		var replyToMessage *domain.MessageWithDecryptedData
		if msg.ReplyToMessage != nil {
			decryptedReply, err := m.DecryptMessage(&domain.RawMessageWithReply{
				RawMessage: *msg.ReplyToMessage,
			}, encryptionKey, syncKey, userIdentity, recipientIdentity)
			if err != nil {
				return nil, err
			}
			if decryptedReply == nil {
				return nil, errors.New("decrypted reply is nil")
			}

			replyToMessage = &domain.MessageWithDecryptedData{
				RawMessageWithReply: domain.RawMessageWithReply{
					RawMessage: *msg.ReplyToMessage,
				},
				Message: *decryptedReply,
			}
		}

		result = append(result, domain.Message{
			MessageWithDecryptedData: domain.MessageWithDecryptedData{
				RawMessageWithReply: *msg,
				Message:             *decryptedMessage,
			},
			ReplyToMessage: replyToMessage,
		})
	}

	return result, nil
}
