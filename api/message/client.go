package message

import (
	"github.com/slipe-fun/bloom-kit/api"
)

type MessageClient struct {
	client *api.Client
}

func NewMessageClient(client *api.Client) *MessageClient {
	return &MessageClient{
		client: client,
	}
}
