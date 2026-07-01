package auth

import (
	"context"
	"fmt"

	"github.com/slipe-fun/bloom-kit/api"
	"github.com/slipe-fun/bloom-kit/domain"
)

func (a *AuthClient) BeginLogin(ctx context.Context, userID string) (*domain.BeginLoginResponse, error) {
	return api.Send[struct{}, domain.BeginLoginResponse](
		ctx,
		a.client,
		"GET",
		fmt.Sprintf("/auth/login/begin/%s", userID),
		nil,
	)
}

func (a *AuthClient) FinishLogin(ctx context.Context, finishLoginRequest *domain.FinishLoginRequest) (*domain.RegisterResponse, error) {
	return api.Send[domain.FinishLoginRequest, domain.RegisterResponse](
		ctx,
		a.client,
		"POST",
		"/auth/login/finish",
		finishLoginRequest,
	)
}
