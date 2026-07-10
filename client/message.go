package client

import (
	"context"
	"encoding/json"

	"github.com/slipe-fun/bloom-kit/domain"
	"github.com/slipe-fun/bloom-kit/managers/message"
	"github.com/slipe-fun/bloom-kit/mappers"
)

func (c *BloomClient) loadMessages(chatID, beforeID int) ([]domain.DecryptedMessageWithReply, error) {
	messagesFromStorage, err := c.database.GetMessages(chatID, beforeID, 20)
	if err != nil {
		return nil, err
	}

	var sourceMessages []domain.Message

	if len(messagesFromStorage) > 0 {
		sourceMessages = messagesFromStorage
	} else {
		chat, err := c.database.GetChat(chatID)
		if err != nil {
			return nil, err
		}

		recipient := getChatOtherMember(&domain.Chat{
			RawChat: chat.RawChat,
		}, c.credentials.UserID)
		recipientIdentity := mappers.ConvertUserToIdentity(recipient)

		user := getChatOtherMember(&domain.Chat{
			RawChat: chat.RawChat,
		}, recipient.ID)
		userIdentity := mappers.ConvertUserToIdentity(user)

		direction := message.DirectionBefore
		if beforeID == 0 {
			direction = message.DirectionAfter
		}

		messagesFromServer, err := c.messageManager.GetMessages(
			context.Background(),
			chatID,
			beforeID,
			direction,
			chat.ChatKey,
			chat.SyncKey,
			userIdentity,
			recipientIdentity,
		)
		if err != nil {
			return nil, err
		}

		if len(messagesFromServer) > 0 {
			err = c.database.SaveMessages(messagesFromServer)
			if err != nil {
				return nil, err
			}
			sourceMessages = messagesFromServer
		}
	}

	result := make([]domain.DecryptedMessageWithReply, len(sourceMessages))
	for i := range sourceMessages {
		result[i] = mappers.MapDomainMessageToDecrypted(&sourceMessages[i])
	}

	return result, nil
}

func (c *BloomClient) LoadMessages(chatID, beforeID int) ([]byte, error) {
	messages, err := c.loadMessages(chatID, beforeID)
	if err != nil {
		return nil, err
	}

	return json.Marshal(messages)
}
