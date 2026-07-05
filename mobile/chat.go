package mobile

import (
	"context"
	"encoding/json"

	"github.com/slipe-fun/bloom-kit/domain"
	"github.com/slipe-fun/skid-v4/pkg/identity"
)

type CreateChatRequest struct {
	UserID            string
	MlKem768PublicKey string
	X448PublicKey     string
	Ed448PublicKey    string
}

type ChatResponse struct {
	domain.Chat
	Me        json.RawMessage `json:"me"`
	Recipient json.RawMessage `json:"recipient"`
}

func (c *BloomClient) CreateChat(receiverUser *CreateChatRequest) ([]byte, error) {
	creds, err := c.loadCredentials()
	if err != nil {
		return nil, err
	}

	publicKeys, err := unmapPublicKeys(&creds.PublicKeys)
	if err != nil {
		return nil, err
	}

	secretKeys, err := unmapSecretKeys(creds.SecretKeys)
	if err != nil {
		return nil, err
	}
	defer secretKeys.Wipe()

	receiverPublicKeys, err := unmapPublicKeys(&PublicKeys{
		MlKem768: receiverUser.MlKem768PublicKey,
		X448:     receiverUser.X448PublicKey,
		Ed448:    receiverUser.Ed448PublicKey,
	})
	if err != nil {
		return nil, err
	}

	sender := identity.User{
		ID:         creds.UserID,
		PublicKeys: *publicKeys,
	}

	receiver := identity.User{
		ID: receiverUser.UserID,
		PublicKeys: identity.PublicKeys{
			MlKem768: receiverPublicKeys.MlKem768,
			X448:     receiverPublicKeys.X448,
			Ed448:    receiverPublicKeys.Ed448,
		},
	}

	createdChat, chatKey, syncKey, err := c.chatManager.Create(context.Background(), &sender, &receiver, secretKeys)
	if err != nil {
		return nil, err
	}

	err = c.database.SaveChat(*createdChat, chatKey, syncKey)
	if err != nil {
		return nil, err
	}

	me := creds.UserJSON

	var recipient []byte
	for _, member := range createdChat.Members {
		if member.ID != creds.UserID {
			userBytes, err := json.Marshal(member)
			if err != nil {
				return nil, err
			}
			recipient = userBytes
			break
		}
	}

	response := ChatResponse{
		Chat:      *createdChat,
		Me:        me,
		Recipient: recipient,
	}

	return json.Marshal(response)
}
