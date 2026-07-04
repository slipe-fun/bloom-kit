package mobile

import (
	"context"
	"encoding/json"
)

func (c *BloomClient) GetMe() ([]byte, error) {
	user, err := c.userManager.GetMe(context.Background())
	if err != nil {
		return nil, err
	}

	userBytes, err := json.Marshal(user)
	if err != nil {
		return nil, err
	}

	return userBytes, nil
}
