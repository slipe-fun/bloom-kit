package mobile

import (
	"github.com/slipe-fun/bloom-kit/api"
	authClient "github.com/slipe-fun/bloom-kit/api/auth"
	userClient "github.com/slipe-fun/bloom-kit/api/user"
	authManager "github.com/slipe-fun/bloom-kit/managers/auth"
	userManager "github.com/slipe-fun/bloom-kit/managers/user"
)

type BloomClient struct {
	apiClient   *api.Client
	authManager *authManager.AuthManager
	userManager *userManager.UserManager
}

func NewClient(baseURL string) *BloomClient {
	c := api.NewClient(baseURL)

	ac := authClient.NewAuthClient(c)
	uc := userClient.NewUserClient(c)

	return &BloomClient{
		apiClient:   c,
		authManager: authManager.NewAuthManager(ac),
		userManager: userManager.NewUserManager(uc),
	}
}

func (c *BloomClient) SetToken(token string) {
	c.apiClient.SetToken(token)
}
