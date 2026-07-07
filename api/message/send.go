package message

import (
	"context"

	"github.com/slipe-fun/bloom-kit/api"
	"github.com/slipe-fun/bloom-kit/domain"
)

func (m *MessageClient) Send(ctx context.Context, sendMessageRequest *domain.SendMessageRequest) (*domain.RawMessageWithReply, error) {
	return api.Send[domain.SendMessageRequest, domain.RawMessageWithReply](ctx, m.client, "POST", "/message/send", sendMessageRequest)
}
