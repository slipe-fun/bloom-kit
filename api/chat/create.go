package chat

import (
	"context"

	"github.com/slipe-fun/bloom-kit/api"
	"github.com/slipe-fun/bloom-kit/domain"
	"github.com/slipe-fun/bloom-kit/mappers"
	"github.com/slipe-fun/skid-v4/pkg/identity"
)

func (c *ChatClient) CreatePrivateChat(ctx context.Context, recipient string, handshake *identity.HandshakePayload) (*domain.Chat, error) {
	req := &domain.CreateChatRequest{
		Type:      "private",
		Recipient: recipient,
		Handshake: mappers.MapHandshake(handshake),
	}
	return api.Send[domain.CreateChatRequest, domain.Chat](ctx, c.client, "POST", "/chat/create", req)
}

func (c *ChatClient) CreateGroupChat(ctx context.Context, title string, members []domain.GroupMemberRequest) (*domain.Chat, error) {
	req := &domain.CreateChatRequest{
		Type:    "group",
		Title:   title,
		Members: members,
	}
	return api.Send[domain.CreateChatRequest, domain.Chat](ctx, c.client, "POST", "/chat/create", req)
}
