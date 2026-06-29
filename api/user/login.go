package user

import (
	"context"
	"fmt"

	"github.com/slipe-fun/bloom-kit/api"
	"github.com/slipe-fun/bloom-kit/domain"
)

func (u *UserClient) BeginLogin(ctx context.Context, userID string) (*domain.BeginLoginResponse, error) {
	return api.Send[struct{}, domain.BeginLoginResponse](
		ctx,
		u.client,
		"GET",
		fmt.Sprintf("/auth/login/begin/%s", userID),
		nil,
	)
}

func (u *UserClient) FinishLogin(ctx context.Context, finishLoginRequest *domain.FinishLoginRequest) (*domain.RegisterResponse, error) {
	return api.Send[domain.FinishLoginRequest, domain.RegisterResponse](
		ctx,
		u.client,
		"POST",
		"/auth/login/finish",
		finishLoginRequest,
	)
}
