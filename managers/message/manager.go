package message

import (
	"github.com/slipe-fun/bloom-kit/api/message"
)

type MessageManager struct {
	messageClient *message.MessageClient
}

func NewMessageManager(messageClient *message.MessageClient) *MessageManager {
	return &MessageManager{
		messageClient: messageClient,
	}
}
