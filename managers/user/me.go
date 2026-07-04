package user

import (
	"context"

	"github.com/slipe-fun/bloom-kit/domain"
)

func (u *UserManager) GetMe(ctx context.Context) (*domain.User, error) {
	return u.userClient.GetMe(ctx)
}
