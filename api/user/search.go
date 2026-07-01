package user

import (
	"context"
	"fmt"

	"github.com/slipe-fun/bloom-kit/api"
	"github.com/slipe-fun/bloom-kit/domain"
)

func (u *UserClient) Search(ctx context.Context, query string) (*[]domain.User, error) {
	return api.Send[struct{}, []domain.User](
		ctx,
		u.client,
		"GET",
		fmt.Sprintf("/user/search?q=%s", query),
		nil,
	)
}
