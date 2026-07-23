package auth

import (
	"context"

	"github.com/slipe-fun/bloom-kit/api/auth"
	"github.com/slipe-fun/bloom-kit/domain"
	"github.com/slipe-fun/skid-v4/pkg/identity"
)

func (a *AuthManager) Register(
	ctx context.Context,
	user *identity.User,
	secretKeys *identity.SecretKeys,
	masterKey, recoveryKey []byte,
) (*domain.RegisterResponse, error) {
	authLookupID := identity.DeriveAuthLookupID(recoveryKey)

	encryptedSecretKeys, err := identity.EncryptSecretKeys(user, secretKeys, masterKey)
	if err != nil {
		return nil, err
	}

	encryptedMasterKey, err := identity.EncryptMasterKey(masterKey, recoveryKey, user, secretKeys)
	if err != nil {
		return nil, err
	}

	registerRequestBody := auth.NewKeysRequest(user, authLookupID, encryptedSecretKeys, encryptedMasterKey)

	return a.authClient.Register(ctx, registerRequestBody)
}
