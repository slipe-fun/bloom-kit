package client

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/slipe-fun/bloom-kit/domain"
	"github.com/slipe-fun/bloom-kit/mappers"
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
	createdChat, err := c.createChat(receiverUser)
	if err != nil {
		return nil, err
	}

	return json.Marshal(createdChat)
}

func (c *BloomClient) GetChats() ([]byte, error) {
	chats, err := c.getChats()
	if err != nil {
		return nil, err
	}

	return json.Marshal(chats)
}

func (c *BloomClient) GetLocalChats() ([]byte, error) {
	chats, err := c.getLocalChats()
	if err != nil {
		return nil, err
	}

	return json.Marshal(chats)
}

func (c *BloomClient) RegisterChatsListener(listener ChatsListener) {
	c.listenerMu.Lock()
	c.chatsListener = listener
	c.listenerMu.Unlock()

	c.notifyChatsUpdated()
}

func (c *BloomClient) UnregisterChatsListener() {
	c.listenerMu.Lock()
	c.chatsListener = nil
	c.listenerMu.Unlock()
}

func (c *BloomClient) StartChatsSync() {
	c.syncMu.Lock()
	defer c.syncMu.Unlock()

	if c.syncCancel != nil {
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	c.syncCancel = cancel

	go func() {
		c.syncRemoteChats()

		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				c.syncRemoteChats()
			}
		}
	}()
}

func (c *BloomClient) StopChatsSync() {
	c.syncMu.Lock()
	defer c.syncMu.Unlock()

	if c.syncCancel != nil {
		c.syncCancel()
		c.syncCancel = nil
	}
}

func getChatOtherMember(chat *domain.Chat, memberID string) *domain.User {
	for i, member := range chat.Members {
		if member.ID != memberID {
			return &chat.Members[i]
		}
	}
	return nil
}

func (c *BloomClient) createChat(receiverUser *CreateChatRequest) (*ChatResponse, error) {
	creds, err := c.loadCredentials()
	if err != nil {
		return nil, err
	}

	publicKeys, err := mappers.UnmapPublicKeys(&creds.PublicKeys)
	if err != nil {
		return nil, err
	}

	secretKeys, err := mappers.UnmapSecretKeys(creds.SecretKeys)
	if err != nil {
		return nil, err
	}
	defer secretKeys.Wipe()

	receiverPublicKeys, err := mappers.UnmapPublicKeys(&domain.PublicKeys{
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

	return &response, nil
}

func (c *BloomClient) getChats() (*[]ChatResponse, error) {
	creds, err := c.loadCredentials()
	if err != nil {
		return nil, err
	}

	publicKeys, err := mappers.UnmapPublicKeys(&c.credentials.PublicKeys)
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

	chats, err := c.chatManager.GetChats(context.Background())
	if err != nil {
		return nil, err
	}

	localChats, err := c.database.GetChats()
	if err != nil {
		return nil, err
	}
	localChatsMap := make(map[int]domain.ChatWithKeys, len(localChats))
	for _, lc := range localChats {
		localChatsMap[lc.ID] = lc
	}

	chatsSlice := *chats
	var result []ChatResponse
	var ids []int
	var members []domain.User
	var messages []domain.Message

	chatsMap := make(map[int]*domain.Chat, len(chatsSlice))
	resultMap := make(map[int]int)

	for i := range chatsSlice {
		chat := &chatsSlice[i]
		chatsMap[chat.ID] = chat

		recipient := getChatOtherMember(chat, c.credentials.UserID)
		if recipient == nil {
			return nil, errors.New("no chat recipient")
		}
		recipientIdentity := mappers.ConvertUserToIdentity(recipient)

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

		if lc, exists := localChatsMap[chat.ID]; exists {
			if chat.EncryptedLastMessage != nil {
				if lc.LastMessage != nil && lc.LastMessage.ID == chat.EncryptedLastMessage.ID {
					newChatObject.LastMessage = lc.LastMessage
				} else {
					decryptedLastMessage, err := c.messageManager.DecryptMessage(chat.EncryptedLastMessage, lc.ChatKey, lc.SyncKey, userIdentity, recipientIdentity)
					if err == nil {
						newChatObject.LastMessage = &domain.Message{
							MessageWithDecryptedData: domain.MessageWithDecryptedData{
								RawMessageWithReply: *chat.EncryptedLastMessage,
								Message:             *decryptedLastMessage,
							},
						}
						messages = append(messages, *newChatObject.LastMessage)
					} else if lc.LastMessage != nil {
						newChatObject.LastMessage = lc.LastMessage
					}
				}
			} else if lc.LastMessage != nil {
				newChatObject.LastMessage = lc.LastMessage
			}
		}

		result = append(result, newChatObject)
		resultMap[chat.ID] = len(result) - 1
	}

	missing, err := c.database.GetMissingChatIDs(ids)
	if err != nil {
		return nil, err
	}

	secretKeys, err := mappers.UnmapSecretKeys(creds.SecretKeys)
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
			continue
		}

		recipientIdentity := mappers.ConvertUserToIdentity(recipient)

		handshakePayload, err := mappers.DecodeHandshake(chat.Handshake)
		if err != nil {
			continue
		}

		chatKey, syncKey, err := c.chatManager.FinalizeHandshake(handshakePayload, userIdentity, secretKeys, recipientIdentity)
		if err != nil {
			continue
		}

		var decryptedLastMessage *domain.Message
		if chat.EncryptedLastMessage != nil {
			decLastMessage, err := c.messageManager.DecryptMessage(chat.EncryptedLastMessage, chatKey, syncKey, userIdentity, recipientIdentity)
			if err == nil {
				decryptedLastMessage = &domain.Message{
					MessageWithDecryptedData: domain.MessageWithDecryptedData{
						RawMessageWithReply: *chat.EncryptedLastMessage,
						Message:             *decLastMessage,
					},
				}
				messages = append(messages, *decryptedLastMessage)

				if idx, found := resultMap[id]; found {
					result[idx].LastMessage = decryptedLastMessage
				}
			}
		}

		newChat := domain.ChatWithKeys{
			RawChat: chat.RawChat,
			ChatKey: chatKey,
			SyncKey: syncKey,
		}

		missingChats = append(missingChats, newChat)
	}

	if len(missingChats) > 0 {
		err = c.database.SaveChats(missingChats)
		if err != nil {
			return nil, err
		}
	}

	if len(members) > 0 {
		err = c.database.SaveUsers(members)
		if err != nil {
			return nil, err
		}
	}

	if len(messages) > 0 {
		err = c.database.SaveMessages(messages)
		if err != nil {
			return nil, err
		}
	}

	return &result, nil
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
				LastMessage: chat.LastMessage,
			},
		}

		newChatObject.Me = c.credentials.UserJSON

		recipient := getChatOtherMember(&domain.Chat{
			RawChat: chat.RawChat,
		}, c.credentials.UserID)
		if recipient == nil {
			continue
		}
		recipientJSON, err := json.Marshal(recipient)
		if err != nil {
			continue
		}
		newChatObject.Recipient = recipientJSON

		result = append(result, newChatObject)
	}

	return result, nil
}

func (c *BloomClient) notifyChatsUpdated() {
	c.listenerMu.Lock()
	listener := c.chatsListener
	c.listenerMu.Unlock()

	if listener == nil {
		return
	}

	localChatsBytes, err := c.GetLocalChats()
	if err != nil {
		return
	}

	listener.OnChatsUpdated(localChatsBytes)
}

func (c *BloomClient) syncRemoteChats() {
	_, err := c.getChats()
	if err != nil {
		return
	}

	c.notifyChatsUpdated()
}
