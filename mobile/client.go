package mobile

import (
	"github.com/slipe-fun/bloom-kit/api"
	authClient "github.com/slipe-fun/bloom-kit/api/auth"
	chatClient "github.com/slipe-fun/bloom-kit/api/chat"
	userClient "github.com/slipe-fun/bloom-kit/api/user"
	authManager "github.com/slipe-fun/bloom-kit/managers/auth"
	chatManager "github.com/slipe-fun/bloom-kit/managers/chat"
	userManager "github.com/slipe-fun/bloom-kit/managers/user"
)

type BloomClient struct {
	apiClient     *api.Client
	authManager   *authManager.AuthManager
	userManager   *userManager.UserManager
	chatManager   *chatManager.ChatManager
	storagePath   string
	encryptionKey []byte
}

func NewClient(baseURL, storagePath string, encryptionKey []byte) *BloomClient {
	c := api.NewClient(baseURL)

	ac := authClient.NewAuthClient(c)
	uc := userClient.NewUserClient(c)
	cc := chatClient.NewChatClient(c)

	localKey := make([]byte, len(encryptionKey))
	copy(localKey, encryptionKey)

	return &BloomClient{
		apiClient:     c,
		authManager:   authManager.NewAuthManager(ac),
		userManager:   userManager.NewUserManager(uc),
		chatManager:   chatManager.NewChatManager(cc),
		storagePath:   storagePath,
		encryptionKey: localKey,
	}
}

func (c *BloomClient) SetToken(token string) {
	c.apiClient.SetToken(token)
}
