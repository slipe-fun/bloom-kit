package user

import (
	"context"

	"github.com/slipe-fun/bloom-kit/api"
	"github.com/slipe-fun/bloom-kit/domain"
)

func (u *UserClient) Edit(ctx context.Context, editUserRequest *domain.EditUserRequest) (*domain.EditUserResponse, error) {
	return api.Send[domain.EditUserRequest, domain.EditUserResponse](ctx, u.client, "POST", "/user/edit", editUserRequest)
}
