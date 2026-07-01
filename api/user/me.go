package user

import (
	"context"

	"github.com/slipe-fun/bloom-kit/api"
	"github.com/slipe-fun/bloom-kit/domain"
)

func (u *UserClient) GetMe(ctx context.Context) (*domain.User, error) {
	return api.Send[struct{}, domain.User](ctx, u.client, "GET", "/user/me", nil)
}
