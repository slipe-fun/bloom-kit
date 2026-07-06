package client

import (
	"context"
	"encoding/json"

	"github.com/slipe-fun/bloom-kit/domain"
)

type EditRequest struct {
	Username    string
	HasUsername bool

	DisplayName    string
	HasDisplayName bool

	Description    string
	HasDescription bool
}

func (c *BloomClient) GetMe() ([]byte, error) {
	user, err := c.userManager.GetMe(context.Background())
	if err != nil {
		return nil, err
	}

	err = c.database.SaveUser(user)
	if err != nil {
		return nil, err
	}

	userBytes, err := json.Marshal(user)
	if err != nil {
		return nil, err
	}

	return userBytes, nil
}

func (c *BloomClient) SearchUsers(query string) ([]byte, error) {
	users, err := c.userManager.Search(context.Background(), query)
	if err != nil {
		return nil, err
	}

	err = c.database.SaveUsers(*users)
	if err != nil {
		return nil, err
	}

	usersBytes, err := json.Marshal(users)
	if err != nil {
		return nil, err
	}

	return usersBytes, nil
}

func (c *BloomClient) EditUser(req *EditRequest) ([]byte, error) {
	var username *string
	if req.HasUsername {
		username = &req.Username
	}

	var displayName *string
	if req.HasDisplayName {
		displayName = &req.DisplayName
	}

	var description *string
	if req.HasDescription {
		description = &req.Description
	}

	editedUser, err := c.userManager.Edit(
		context.Background(),
		username,
		displayName,
		description,
	)
	if err != nil {
		return nil, err
	}

	err = c.database.SaveUser(&editedUser.User)
	if err != nil {
		return nil, err
	}

	return json.Marshal(editedUser.User)
}

func (c *BloomClient) GetUser(userID string) ([]byte, error) {
	user, err := c.userManager.Get(context.Background(), userID)
	if err != nil {
		return nil, err
	}

	err = c.database.SaveUser(user)
	if err != nil {
		return nil, err
	}

	userBytes, err := json.Marshal(user)
	if err != nil {
		return nil, err
	}

	return userBytes, nil
}

func (c *BloomClient) getOrFetchUser(userID string) (*domain.User, error) {
	user, err := c.database.GetUser(userID)
	if err != nil {
		return nil, err
	}

	if user == nil {
		user, err = c.userManager.Get(context.Background(), userID)
		if err != nil {
			return nil, err
		}
		err = c.database.SaveUser(user)
		if err != nil {
			return nil, err
		}
	}

	return user, nil
}
