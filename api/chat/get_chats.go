package chat

import (
	"context"

	"github.com/slipe-fun/bloom-kit/api"
	"github.com/slipe-fun/bloom-kit/domain"
)

func (c *ChatClient) GetChats(ctx context.Context) (*[]domain.Chat, error) {
	return api.Send[struct{}, []domain.Chat](
		ctx,
		c.client,
		"GET",
		"/chats",
		nil,
	)
}
