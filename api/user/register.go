package user

import (
	"context"

	"github.com/slipe-fun/bloom-kit/api"
	"github.com/slipe-fun/bloom-kit/domain"
)

func (u *UserClient) Register(ctx context.Context, keys *domain.KeysRequest) (*domain.RegisterResponse, error) {
	return api.Send[domain.KeysRequest, domain.RegisterResponse](ctx, u.client, "POST", "/auth/register", keys)
}
