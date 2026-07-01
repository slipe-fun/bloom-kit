package auth

import "github.com/slipe-fun/bloom-kit/api/auth"

type AuthManager struct {
	authClient *auth.AuthClient
}

func NewAuthManager(authClient *auth.AuthClient) *AuthManager {
	return &AuthManager{
		authClient: authClient,
	}
}
