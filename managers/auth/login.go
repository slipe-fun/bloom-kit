package auth

import (
	"context"

	"github.com/slipe-fun/bloom-kit/api/auth"
	"github.com/slipe-fun/bloom-kit/domain"
	"github.com/slipe-fun/skid-v4/pkg/identity"
)

func (a *AuthManager) BeginLogin(ctx context.Context, userID string) (*domain.BeginLoginResponse, error) {
	return a.authClient.BeginLogin(ctx, userID)
}

func (a *AuthManager) FinishLogin(
	ctx context.Context,
	beginLoginResponse *domain.BeginLoginResponse,
	user *identity.User,
	secretKeys *identity.SecretKeys,
	recoveryKey []byte,
) (*domain.RegisterResponse, []byte, *identity.SecretKeys, error) {
	encryptedMasterKeyBytes, encryptedSecretKeysBytes, challenge, err := auth.DecodeBeginLoginResponse(beginLoginResponse, user)
	if err != nil {
		return nil, nil, nil, err
	}

	decryptedMasterKey, err := identity.DecryptMasterKey(encryptedMasterKeyBytes, recoveryKey, user)
	if err != nil {
		return nil, nil, nil, err
	}

	decryptedSecretKeys, err := identity.DecryptSecretKeys(encryptedSecretKeysBytes, user, decryptedMasterKey)
	if err != nil {
		return nil, nil, nil, err
	}
	defer decryptedSecretKeys.Wipe()

	finishLoginRequest, err := auth.NewFinishLoginRequest(user, secretKeys, challenge)
	if err != nil {
		return nil, nil, nil, err
	}

	finishLoginResponse, err := a.authClient.FinishLogin(ctx, finishLoginRequest)
	if err != nil {
		return nil, nil, nil, err
	}

	return finishLoginResponse, decryptedMasterKey, decryptedSecretKeys, nil
}
