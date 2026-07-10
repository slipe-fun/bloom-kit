package message

import (
	"context"
	"time"

	"github.com/slipe-fun/bloom-kit/api/message"
	"github.com/slipe-fun/bloom-kit/domain"
	"github.com/slipe-fun/skid-v4/pkg/identity"
	"github.com/slipe-fun/skid-v4/pkg/messages"
)

func (m *MessageManager) Send(
	ctx context.Context,
	content string,
	chatID int,
	replyToID *int,
	chatKey []byte,
	syncKey []byte,
	userIdentity *identity.User,
	receiverIdentity *identity.User,
) (*domain.Message, error) {
	encryptedMessage, err := messages.Encrypt(chatKey, []byte(content), syncKey, userIdentity, receiverIdentity)
	if err != nil {
		return nil, err
	}

	sendMessageRequest := message.NewSendMessageRequest(chatID, replyToID, *encryptedMessage)

	sentMessage, err := m.messageClient.Send(ctx, sendMessageRequest)
	if err != nil {
		return nil, err
	}

	var replyToMessage *domain.MessageWithDecryptedData
	if sentMessage.ReplyToMessage != nil {
		replyToMessage = &domain.MessageWithDecryptedData{
			RawMessageWithReply: domain.RawMessageWithReply{
				RawMessage: *sentMessage.ReplyToMessage,
			},
		}
	}

	return &domain.Message{
		MessageWithDecryptedData: domain.MessageWithDecryptedData{
			RawMessageWithReply: *sentMessage,
			Message: messages.Message{
				Content:   []byte(content),
				AuthorID:  userIdentity.ID,
				SyncTag:   syncKey,
				Timestamp: time.Now().Unix(),
			},
		},
		ReplyToMessage: replyToMessage,
	}, nil
}
