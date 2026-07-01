package user

import (
	"context"

	"github.com/slipe-fun/bloom-kit/api/user"
	"github.com/slipe-fun/bloom-kit/domain"
)

func (u *UserManager) Edit(ctx context.Context, username, displayName, description *string) (*domain.EditUserResponse, error) {
	editUserRequest := user.NewEditUserRequest(username, displayName, description)

	return u.userClient.Edit(ctx, editUserRequest)
}
