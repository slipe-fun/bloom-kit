package user

import "github.com/slipe-fun/bloom-kit/api/user"

type UserManager struct {
	userClient *user.UserClient
}

func NewUserManager(userClient *user.UserClient) *UserManager {
	return &UserManager{
		userClient: userClient,
	}
}
