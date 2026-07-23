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
			for i, j := 0, len(messagesFromServer)-1; i < j; i, j = i+1, j-1 {
				messagesFromServer[i], messagesFromServer[j] = messagesFromServer[j], messagesFromServer[i]
			}

			err = c.database.SaveMessages(messagesFromServer)
			if err != nil {
				return nil, err
			}
			sourceMessages = messagesFromServer

			if beforeID == 0 {
				c.notifyChatsUpdated()
			}
		}
	}

	result := make([]domain.DecryptedMessageWithReply, len(sourceMessages))
	for i := range sourceMessages {
		result[i] = mappers.MapDomainMessageToDecrypted(&sourceMessages[i])
	}

	return result, nil
}

func (c *BloomClient) sendMessage(chatID int, replyToID *int, content string) (*domain.DecryptedMessageWithReply, error) {
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

	message, err := c.messageManager.Send(context.Background(), content, chatID, replyToID, chat.ChatKey, chat.SyncKey, userIdentity, recipientIdentity)
	if err != nil {
		return nil, err
	}

	var replyTo *domain.DecryptedMessage
	if message.ReplyToMessage != nil {
		replyTo = &domain.DecryptedMessage{
			ID:        message.ReplyToMessage.ID,
			Content:   string(message.ReplyToMessage.Content),
			AuthorID:  message.ReplyToMessage.AuthorID,
			Timestamp: message.ReplyToMessage.Timestamp,
			Seen:      message.ReplyToMessage.Seen,
		}
	}

	return &domain.DecryptedMessageWithReply{
		DecryptedMessage: domain.DecryptedMessage{
			ID:        message.ID,
			Content:   content,
			AuthorID:  userIdentity.ID,
			Timestamp: message.Timestamp,
			Seen:      message.Seen,
		},
		ReplyTo: replyTo,
	}, nil
}

func (c *BloomClient) SendMessage(chatID int, replyToID *int, content string) ([]byte, error) {
	message, err := c.sendMessage(chatID, replyToID, content)
	if err != nil {
		return nil, err
	}

	return json.Marshal(message)
}

func (c *BloomClient) LoadMessages(chatID, beforeID int) ([]byte, error) {
	messages, err := c.loadMessages(chatID, beforeID)
	if err != nil {
		return nil, err
	}

	return json.Marshal(messages)
}

func (c *BloomClient) RegisterMessagesListener(listener MessagesListener) {
	c.listenerMu.Lock()
	c.messagesListener = listener
	c.listenerMu.Unlock()
}

func (c *BloomClient) UnregisterMessagesListener() {
	c.listenerMu.Lock()
	c.messagesListener = nil
	c.listenerMu.Unlock()
}
