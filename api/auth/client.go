package auth

import (
	"github.com/slipe-fun/bloom-kit/api"
)

type AuthClient struct {
	client *api.Client
}

func NewAuthClient(client *api.Client) *AuthClient {
	return &AuthClient{
		client: client,
	}
}
