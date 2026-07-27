package user

import (
	"context"
	"fmt"
	"strings"

	"github.com/slipe-fun/bloom-kit/api"
	"github.com/slipe-fun/bloom-kit/domain"
)

func (u *UserClient) GetByIDs(ctx context.Context, ids []string) (*[]domain.User, error) {
	return api.Send[struct{}, []domain.User](
		ctx,
		u.client,
		"GET",
		fmt.Sprintf("/users?ids=%s", strings.Join(ids, ",")),
		nil,
	)
}
