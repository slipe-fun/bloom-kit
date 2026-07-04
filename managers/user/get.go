package user

import (
	"context"

	"github.com/slipe-fun/bloom-kit/domain"
)

func (u *UserManager) Get(ctx context.Context, userID string) (*domain.User, error) {
	return u.userClient.Get(ctx, userID)
}
