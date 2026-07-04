package user

import (
	"context"

	"github.com/slipe-fun/bloom-kit/domain"
)

func (u *UserManager) Search(ctx context.Context, query string) (*[]domain.User, error) {
	return u.userClient.Search(ctx, query)
}
