package chat

import (
	"context"

	"github.com/slipe-fun/bloom-kit/api/chat"
	"github.com/slipe-fun/bloom-kit/domain"
	"github.com/slipe-fun/skid-v4/pkg/identity"
)

func (c *ChatManager) Create(
	ctx context.Context,
	sender *identity.User,
	receiver *identity.User,
	secretKeys *identity.SecretKeys,
) (*domain.Chat, *identity.HandshakePayload, []byte, []byte, error) {
	hanshake, chatKey, syncKey, err := identity.InitiateKeyExchange(sender, secretKeys, receiver)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	createChatRequest := chat.NewCreateChatRequest(receiver.ID, hanshake)
	createChatResponse, err := c.chatClient.Create(ctx, createChatRequest)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	return createChatResponse, hanshake, chatKey, syncKey, nil
}
