package chat

import (
	"context"

	"github.com/slipe-fun/bloom-kit/api"
	"github.com/slipe-fun/bloom-kit/domain"
)

func (c *ChatClient) Create(ctx context.Context, createChatRequest *domain.CreateChatRequest) (*domain.Chat, error) {
	return api.Send[domain.CreateChatRequest, domain.Chat](ctx, c.client, "POST", "/chat/create", createChatRequest)
}
