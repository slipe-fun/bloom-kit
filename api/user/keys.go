package user

import (
	"context"
	"fmt"

	"github.com/slipe-fun/bloom-kit/api"
	"github.com/slipe-fun/bloom-kit/domain"
)

func (u *UserClient) GetKeys(ctx context.Context, keysType string) (*domain.GetKeysResponse, error) {
	return api.Send[struct{}, domain.GetKeysResponse](
		ctx,
		u.client,
		"GET",
		fmt.Sprintf("/user/keys/%s", keysType),
		nil,
	)
}
