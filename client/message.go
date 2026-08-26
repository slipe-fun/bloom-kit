package client

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/slipe-fun/bloom-kit/domain"
	"github.com/slipe-fun/bloom-kit/mappers"
)

func (c *BloomClient) loadMessagesFromStorage(chatID, beforeID int) ([]domain.DecryptedMessageWithReply, error) {
	if c.credentials == nil {
		return nil, errors.New("unauthorized: client is not logged in")
	}

	messagesFromStorage, err := c.database.GetMessages(chatID, beforeID, 20)
	if err != nil {
		return nil, err
	}

	var sourceMessages []domain.Message

	if len(messagesFromStorage) > 0 {
		sourceMessages = messagesFromStorage
	} else {
		return []domain.DecryptedMessageWithReply{}, nil
	}

	result := make([]domain.DecryptedMessageWithReply, len(sourceMessages))
	for i := range sourceMessages {
		result[i] = mappers.MapDomainMessageToDecrypted(&sourceMessages[i])
	}

	return result, nil
}

func (c *BloomClient) getLastMessagesFromServer(chatID int) ([]domain.DecryptedMessageWithReply, error) {
	if c.credentials == nil {
		return nil, errors.New("unauthorized: client is not logged in")
	}

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

	var sourceMessages []domain.Message

	messages, err := c.messageManager.GetMessages(context.Background(), chatID, 0, "after", chat.ChatKey, chat.SyncKey, userIdentity, recipientIdentity)
	if err != nil {
		return nil, err
	}

	if len(messages) > 0 {
		sourceMessages = messages
	} else {
		return []domain.DecryptedMessageWithReply{}, nil
	}

	result := make([]domain.DecryptedMessageWithReply, len(sourceMessages))
	for i := range sourceMessages {
		result[i] = mappers.MapDomainMessageToDecrypted(&sourceMessages[i])
	}

	return result, nil
}

func (c *BloomClient) sendMessage(chatID int, replyToID *int, content string) (*domain.DecryptedMessageWithReply, error) {
	if c.credentials == nil {
		return nil, errors.New("unauthorized: client is not logged in")
	}

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

	if err := c.database.SaveMessage(message); err != nil {
		return nil, err
	}

	c.notifyChatsUpdated()

	var replyTo *domain.DecryptedMessage
	if message.ReplyToMessage != nil {
		replyTo = &domain.DecryptedMessage{
			ID:        message.ReplyToMessage.ID,
			Content:   string(message.ReplyToMessage.Content),
			AuthorID:  message.ReplyToMessage.AuthorID,
			Timestamp: message.ReplyToMessage.Timestamp,
			CreatedAt: message.ReplyToMessage.CreatedAt,
			Seen:      message.ReplyToMessage.Seen,
		}
	}

	return &domain.DecryptedMessageWithReply{
		DecryptedMessage: domain.DecryptedMessage{
			ID:        message.ID,
			Content:   content,
			AuthorID:  userIdentity.ID,
			Timestamp: message.Timestamp,
			CreatedAt: message.CreatedAt,
			Seen:      message.Seen,
		},
		ReplyTo: replyTo,
	}, nil
}

func (c *BloomClient) SendMessage(chatID int, replyToID int, content string) ([]byte, error) {
	if c.credentials == nil {
		return nil, errors.New("unauthorized: client is not logged in")
	}

	var replyToPtr *int
	if replyToID >= 0 {
		replyToPtr = &replyToID
	}

	message, err := c.sendMessage(chatID, replyToPtr, content)
	if err != nil {
		return nil, err
	}

	return json.Marshal(message)
}

func (c *BloomClient) LoadMessages(sourceType string, chatID, beforeID int) ([]byte, error) {
	if c.credentials == nil {
		return nil, errors.New("unauthorized: client is not logged in")
	}

	var result []domain.DecryptedMessageWithReply

	switch sourceType {
	case "server":
		messages, err := c.getLastMessagesFromServer(chatID)
		if err != nil {
			return nil, err
		}

		result = messages
	case "storage":
		messages, err := c.loadMessagesFromStorage(chatID, beforeID)
		if err != nil {
			return nil, err
		}

		result = messages
	}

	return json.Marshal(result)
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
