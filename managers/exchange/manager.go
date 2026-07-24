package exchange

import (
	"github.com/slipe-fun/bloom-kit/api/exchange"
)

type ExchangeManager struct {
	exchangeClient *exchange.ExchangeClient
}

func NewExchangeManager(exchangeClient *exchange.ExchangeClient) *ExchangeManager {
	return &ExchangeManager{
		exchangeClient: exchangeClient,
	}
}
