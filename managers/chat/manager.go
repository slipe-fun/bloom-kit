package chat

import (
	"github.com/slipe-fun/bloom-kit/api/chat"
)

type ChatManager struct {
	chatClient *chat.ChatClient
}

func NewChatManager(chatClient *chat.ChatClient) *ChatManager {
	return &ChatManager{
		chatClient: chatClient,
	}
}
