package user

import (
	"github.com/slipe-fun/bloom-kit/api"
)

type UserClient struct {
	client *api.Client
}

func NewUserClient(client *api.Client) *UserClient {
	return &UserClient{
		client: client,
	}
}
