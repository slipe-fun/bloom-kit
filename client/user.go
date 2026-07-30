package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"

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
	if c.credentials == nil {
		return nil, errors.New("unauthorized: client is not logged in")
	}

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

	if !bytes.Equal(userBytes, c.credentials.UserJSON) {
		c.UpdateCredentialsUserJSON(userBytes)
	}

	c.updateUser(userBytes)

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
	if c.credentials == nil {
		return nil, errors.New("unauthorized: client is not logged in")
	}

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

	editedUserBytes, err := json.Marshal(editedUser.User)
	if err != nil {
		return nil, err
	}

	c.updateUser(editedUserBytes)

	return editedUserBytes, nil
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

func (c *BloomClient) GetOrFetchMe() ([]byte, error) {
	if c.credentials == nil {
		return nil, errors.New("unauthorized: client is not logged in")
	}

	user, err := c.getOrFetchUser(c.credentials.UserID)
	if err != nil {
		return nil, err
	}

	userJSON, err := json.Marshal(user)
	if err != nil {
		return nil, err
	}

	if !bytes.Equal(userJSON, c.credentials.UserJSON) {
		c.UpdateCredentialsUserJSON(userJSON)
	}

	c.updateUser(userJSON)

	return userJSON, nil
}

func (c *BloomClient) updateUser(userJSON []byte) {
	c.listenerMu.RLock()
	userListener := c.userListener
	c.listenerMu.RUnlock()

	if userListener != nil {
		userListener.OnUserUpdated(userJSON)
	}
}

func (c *BloomClient) RegisterUserListener(listener UserListener) {
	c.listenerMu.Lock()
	c.userListener = listener
	c.listenerMu.Unlock()

	c.GetOrFetchMe()
}

func (c *BloomClient) UnregisterUserListener() {
	c.listenerMu.Lock()
	c.userListener = nil
	c.listenerMu.Unlock()
}
