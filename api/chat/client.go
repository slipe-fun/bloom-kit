package chat

import (
	"github.com/slipe-fun/bloom-kit/api"
)

type ChatClient struct {
	client *api.Client
}

func NewChatClient(client *api.Client) *ChatClient {
	return &ChatClient{
		client: client,
	}
}
