package exchange

import (
	"github.com/slipe-fun/bloom-kit/api"
)

type ExchangeClient struct {
	client *api.Client
}

func NewExchangeClient(client *api.Client) *ExchangeClient {
	return &ExchangeClient{
		client: client,
	}
}
