package exchange

import "context"

func (m *ExchangeManager) StartSession(ctx context.Context) (string, error) {
	session, err := m.exchangeClient.StartSession(ctx)
	if err != nil {
		return "", err
	}

	return session.RoomID, nil
}
