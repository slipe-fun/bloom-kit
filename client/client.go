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
	exchangeClient "github.com/slipe-fun/bloom-kit/api/exchange"
	messageClient "github.com/slipe-fun/bloom-kit/api/message"
	userClient "github.com/slipe-fun/bloom-kit/api/user"
	"github.com/slipe-fun/bloom-kit/database"
	"github.com/slipe-fun/bloom-kit/domain"
	authManager "github.com/slipe-fun/bloom-kit/managers/auth"
	chatManager "github.com/slipe-fun/bloom-kit/managers/chat"
	exchangeManager "github.com/slipe-fun/bloom-kit/managers/exchange"
	messageManager "github.com/slipe-fun/bloom-kit/managers/message"
	userManager "github.com/slipe-fun/bloom-kit/managers/user"
	"github.com/slipe-fun/skid-v4/pkg/identity"
)

type ChatsListener interface {
	OnChatsUpdated(chatsJSON []byte)
}

type MessagesListener interface {
	OnNewMessage(messageJSON []byte)
}

type BloomClient struct {
	wsURL           string
	apiClient       *api.Client
	authManager     *authManager.AuthManager
	exchangeManager *exchangeManager.ExchangeManager
	userManager     *userManager.UserManager
	chatManager     *chatManager.ChatManager
	messageManager  *messageManager.MessageManager
	credentials     *domain.SavedCredentials
	database        *database.Database
	storagePath     string
	encryptionKey   []byte

	exchangeSecretKeys   *identity.SecretKeys
	exchangeUserIdentity *identity.User
	exchangeCancel       context.CancelFunc
	exchangeMu           sync.Mutex

	chatsListener    ChatsListener
	messagesListener MessagesListener
	listenerMu       sync.RWMutex
	syncCancel       context.CancelFunc
	syncMu           sync.Mutex
}

func NewClient(baseURL, wsURL, storagePath string, encryptionKey []byte) *BloomClient {
	c := api.NewClient(baseURL)

	ac := authClient.NewAuthClient(c)
	ec := exchangeClient.NewExchangeClient(c)
	uc := userClient.NewUserClient(c)
	cc := chatClient.NewChatClient(c)
	mc := messageClient.NewMessageClient(c)

	localKey := make([]byte, len(encryptionKey))
	copy(localKey, encryptionKey)

	return &BloomClient{
		wsURL:           wsURL,
		apiClient:       c,
		authManager:     authManager.NewAuthManager(ac),
		exchangeManager: exchangeManager.NewExchangeManager(ec),
		userManager:     userManager.NewUserManager(uc),
		chatManager:     chatManager.NewChatManager(cc),
		messageManager:  messageManager.NewMessageManager(mc),
		storagePath:     storagePath,
		encryptionKey:   localKey,
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
	c.SetToken(c.credentials.Token)
	c.startWebSocket(context.Background(), c.wsURL)
	return nil
}

func (c *BloomClient) SetToken(token string) {
	c.apiClient.SetToken(token)
}
