package exchange

import (
	"context"

	"github.com/slipe-fun/bloom-kit/api"
	"github.com/slipe-fun/bloom-kit/domain"
)

func (c *ExchangeClient) StartSession(ctx context.Context) (*domain.StartExchangeSessionResponse, error) {
	return api.Send[struct{}, domain.StartExchangeSessionResponse](ctx, c.client, "POST", "/exchange/session", nil)
}
