package client

import (
	"context"
	"encoding/json"
	"errors"

	chatApi "github.com/slipe-fun/bloom-kit/api/chat"
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

func getChatOtherMember(chat *domain.Chat, memberID string) *domain.User {
	for i, member := range chat.Members {
		if member.ID != memberID {
			return &chat.Members[i]
		}
	}
	return nil
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

	recipient := getChatOtherMember(createdChat, c.credentials.UserID)
	if recipient == nil {
		return nil, errors.New("no chat recipient")
	}
	recipientJSON, err := json.Marshal(recipient)
	if err != nil {
		return nil, err
	}

	response := ChatResponse{
		Chat:      *createdChat,
		Me:        me,
		Recipient: recipientJSON,
	}

	err = c.database.SaveUsers(createdChat.Members)
	if err != nil {
		return nil, err
	}

	return json.Marshal(response)
}

func (c *BloomClient) getChats() (*[]ChatResponse, error) {
	creds, err := c.loadCredentials()
	if err != nil {
		return nil, err
	}

	chats, err := c.chatManager.GetChats(context.Background())
	if err != nil {
		return nil, err
	}

	chatsSlice := *chats
	var result []ChatResponse
	var ids []int
	var members []domain.User

	chatsMap := make(map[int]*domain.Chat, len(chatsSlice))

	for i := range chatsSlice {
		chat := &chatsSlice[i]
		chatsMap[chat.ID] = chat

		recipient := getChatOtherMember(chat, c.credentials.UserID)
		if recipient == nil {
			return nil, errors.New("no chat recipient")
		}
		members = append(members, *recipient)
		ids = append(ids, chat.ID)

		newChatObject := ChatResponse{
			Chat: domain.Chat{
				RawChat: domain.RawChat{
					ID:        chat.ID,
					Members:   chat.Members,
					Handshake: chat.Handshake,
				},
			},
		}

		newChatObject.Me = c.credentials.UserJSON
		recipientJSON, err := json.Marshal(recipient)
		if err != nil {
			return nil, err
		}
		newChatObject.Recipient = recipientJSON

		result = append(result, newChatObject)
	}

	missing, err := c.database.GetMissingChatIDs(ids)
	if err != nil {
		return nil, err
	}

	publicKeys, err := unmapPublicKeys(&c.credentials.PublicKeys)
	if err != nil {
		return nil, err
	}

	userIdentity := &identity.User{
		ID: c.credentials.UserID,
		PublicKeys: identity.PublicKeys{
			MlKem768: publicKeys.MlKem768,
			X448:     publicKeys.X448,
			Ed448:    publicKeys.Ed448,
		},
	}

	secretKeys, err := unmapSecretKeys(creds.SecretKeys)
	if err != nil {
		return nil, err
	}
	defer secretKeys.Wipe()

	var missingChats []domain.ChatWithKeys
	for _, id := range missing {
		chat, exists := chatsMap[id]
		if !exists {
			continue
		}

		recipient := getChatOtherMember(chat, c.credentials.UserID)
		if recipient == nil {
			return nil, errors.New("no chat recipient")
		}

		recipientPublicKeys, err := unmapPublicKeys(&PublicKeys{
			MlKem768: recipient.MlKemPublicKey,
			X448:     recipient.EcdhPublicKey,
			Ed448:    recipient.EdPublicKey,
		})
		if err != nil {
			return nil, err
		}

		recipientIdentity := &identity.User{
			ID: recipient.ID,
			PublicKeys: identity.PublicKeys{
				MlKem768: recipientPublicKeys.MlKem768,
				X448:     recipientPublicKeys.X448,
				Ed448:    recipientPublicKeys.Ed448,
			},
		}

		handshakePayload, err := chatApi.DecodeHandshake(chat.Handshake)
		if err != nil {
			return nil, err
		}

		chatKey, syncKey, err := identity.FinalizeKeyExchange(handshakePayload, userIdentity, secretKeys, recipientIdentity, nil, true)
		if err != nil {
			chatKey, syncKey, err = identity.FinalizeKeyExchange(handshakePayload, userIdentity, secretKeys, recipientIdentity, nil, false)
			if err != nil {
				return nil, err
			}
		}

		missingChats = append(missingChats, domain.ChatWithKeys{
			RawChat: chat.RawChat,
			ChatKey: chatKey,
			SyncKey: syncKey,
		})
	}

	err = c.database.SaveChats(missingChats)
	if err != nil {
		return nil, err
	}

	err = c.database.SaveUsers(members)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *BloomClient) GetChats() ([]byte, error) {
	chats, err := c.getChats()
	if err != nil {
		return nil, err
	}

	return json.Marshal(chats)
}

func (c *BloomClient) getLocalChats() ([]ChatResponse, error) {
	chats, err := c.database.GetChats()
	if err != nil {
		return nil, err
	}

	result := make([]ChatResponse, 0, len(chats))
	for i := range chats {
		chat := &chats[i]

		newChatObject := ChatResponse{
			Chat: domain.Chat{
				RawChat: domain.RawChat{
					ID:        chat.ID,
					Members:   chat.Members,
					Handshake: chat.Handshake,
				},
			},
		}

		newChatObject.Me = c.credentials.UserJSON

		recipient := getChatOtherMember(&domain.Chat{
			RawChat: chat.RawChat,
		}, c.credentials.UserID)
		if recipient == nil {
			return nil, errors.New("no chat recipient")
		}
		recipientJSON, err := json.Marshal(recipient)
		if err != nil {
			return nil, err
		}
		newChatObject.Recipient = recipientJSON

		result = append(result, newChatObject)
	}

	return result, nil
}

func (c *BloomClient) GetLocalChats() ([]byte, error) {
	chats, err := c.getLocalChats()
	if err != nil {
		return nil, err
	}

	return json.Marshal(chats)
}
