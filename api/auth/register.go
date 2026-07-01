package auth

import (
	"context"

	"github.com/slipe-fun/bloom-kit/api"
	"github.com/slipe-fun/bloom-kit/domain"
)

func (a *AuthClient) Register(ctx context.Context, keys *domain.KeysRequest) (*domain.RegisterResponse, error) {
	return api.Send[domain.KeysRequest, domain.RegisterResponse](ctx, a.client, "POST", "/auth/register", keys)
}
