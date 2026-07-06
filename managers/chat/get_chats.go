package chat

import (
	"context"

	"github.com/slipe-fun/bloom-kit/domain"
)

func (c *ChatManager) GetChats(ctx context.Context) (*[]domain.Chat, error) {
	return c.chatClient.GetChats(ctx)
}
