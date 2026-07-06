package client

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"

	"github.com/slipe-fun/bloom-kit/api"
	authClient "github.com/slipe-fun/bloom-kit/api/auth"
	chatClient "github.com/slipe-fun/bloom-kit/api/chat"
	userClient "github.com/slipe-fun/bloom-kit/api/user"
	"github.com/slipe-fun/bloom-kit/database"
	authManager "github.com/slipe-fun/bloom-kit/managers/auth"
	chatManager "github.com/slipe-fun/bloom-kit/managers/chat"
	userManager "github.com/slipe-fun/bloom-kit/managers/user"
)

type ChatsListener interface {
	OnChatsUpdated(chatsJSON []byte)
}

type BloomClient struct {
	apiClient     *api.Client
	authManager   *authManager.AuthManager
	userManager   *userManager.UserManager
	chatManager   *chatManager.ChatManager
	credentials   *SavedCredentials
	database      *database.Database
	storagePath   string
	encryptionKey []byte

	chatsListener ChatsListener
	listenerMu    sync.RWMutex
	syncCancel    context.CancelFunc
	syncMu        sync.Mutex
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

func (c *BloomClient) Initialize() error {
	dbKey := make([]byte, len(c.encryptionKey))
	copy(dbKey, c.encryptionKey)
	db, err := database.NewDatabase(dbKey, c.storagePath)
	if err != nil {
		return err
	}
	c.database = db
	_, err = c.loadCredentials()
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return fmt.Errorf("failed to load credentials: %w", err)
	}
	return nil
}

func (c *BloomClient) SetToken(token string) {
	c.apiClient.SetToken(token)
}
