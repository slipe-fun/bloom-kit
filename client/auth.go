package client

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"

	"github.com/slipe-fun/bloom-kit/domain"
	"github.com/slipe-fun/bloom-kit/mappers"
	"github.com/slipe-fun/skid-v4/pkg/identity"
)

type RegisterResult struct {
	UserJSON       json.RawMessage `json:"user_json"`
	RawRecoveryKey string          `json:"raw_recovery_key"`
}

type LoginResult struct {
	UserJSON json.RawMessage `json:"user_json"`
}

func (c *BloomClient) Register() (*RegisterResult, error) {
	userIdentity, secret, rawRecoveryKey, recoveryKey, masterKey, lookupID, err := identity.GenerateIdentity()
	if err != nil {
		return nil, err
	}
	defer secret.Wipe()

	registerResponse, err := c.authManager.Register(context.Background(), userIdentity, secret, masterKey, recoveryKey, lookupID)
	if err != nil {
		return nil, err
	}

	mappedSecretKeys, err := mappers.MapSecretKeys(secret)
	if err != nil {
		return nil, err
	}

	userBytes, err := json.Marshal(registerResponse.User)
	if err != nil {
		return nil, err
	}

	err = c.saveCredentials(&domain.SavedCredentials{
		UserID:      registerResponse.User.ID,
		RecoveryKey: recoveryKey,
		MasterKey:   masterKey,
		PublicKeys: *mappers.MapPublicKeys(
			userIdentity.PublicKeys.MlKem768,
			userIdentity.PublicKeys.X448,
			userIdentity.PublicKeys.Ed448,
		),
		SecretKeys: *mappedSecretKeys,
		UserJSON:   userBytes,
		Token:      registerResponse.Token,
	})
	if err != nil {
		return nil, err
	}

	c.database.SaveUser(&registerResponse.User)

	c.apiClient.SetToken(registerResponse.Token)
	c.startWebSocket(context.Background(), c.wsURL)

	c.updateUser(userBytes)

	return &RegisterResult{
		UserJSON:       userBytes,
		RawRecoveryKey: hex.EncodeToString(rawRecoveryKey),
	}, nil
}

func (c *BloomClient) Login(rawRecoveryKey string) (*LoginResult, error) {
	rawRecoveryKeyBytes, err := hex.DecodeString(rawRecoveryKey)
	if err != nil {
		return nil, err
	}

	recoveryKeyBytes, lookupID, err := identity.NewRecoveryKey(rawRecoveryKeyBytes)
	if err != nil {
		return nil, err
	}

	beginLoginResponse, err := c.authManager.BeginLogin(context.Background(), lookupID)
	if err != nil {
		return nil, err
	}

	decodedMlKem768, err := base64.StdEncoding.DecodeString(beginLoginResponse.Keys.IdentityKeys.IdentityPublicKeys.MlKemPublicKey)
	if err != nil {
		return nil, err
	}

	decodedX448, err := base64.StdEncoding.DecodeString(beginLoginResponse.Keys.IdentityKeys.IdentityPublicKeys.EcdhPublicKey)
	if err != nil {
		return nil, err
	}

	decodedEd448, err := base64.StdEncoding.DecodeString(beginLoginResponse.Keys.IdentityKeys.IdentityPublicKeys.EdPublicKey)
	if err != nil {
		return nil, err
	}

	userIdentity := &identity.User{
		ID: beginLoginResponse.UserID,
		PublicKeys: identity.PublicKeys{
			MlKem768: decodedMlKem768,
			X448:     decodedX448,
			Ed448:    decodedEd448,
		},
	}

	finishLoginResult, masterKey, secretKeys, err := c.authManager.FinishLogin(context.Background(), beginLoginResponse, userIdentity, recoveryKeyBytes)
	if err != nil {
		return nil, err
	}
	defer secretKeys.Wipe()

	userBytes, err := json.Marshal(finishLoginResult.User)
	if err != nil {
		return nil, err
	}

	mappedSecretKeys, err := mappers.MapSecretKeys(secretKeys)
	if err != nil {
		return nil, err
	}

	err = c.saveCredentials(&domain.SavedCredentials{
		UserID:      finishLoginResult.User.ID,
		RecoveryKey: recoveryKeyBytes,
		MasterKey:   masterKey,
		PublicKeys: *mappers.MapPublicKeys(
			userIdentity.PublicKeys.MlKem768,
			userIdentity.PublicKeys.X448,
			userIdentity.PublicKeys.Ed448,
		),
		SecretKeys: *mappedSecretKeys,
		UserJSON:   userBytes,
		Token:      finishLoginResult.Token,
	})
	if err != nil {
		return nil, err
	}

	c.database.SaveUser(&finishLoginResult.User)

	c.apiClient.SetToken(finishLoginResult.Token)
	c.startWebSocket(context.Background(), c.wsURL)

	c.updateUser(userBytes)

	return &LoginResult{
		UserJSON: userBytes,
	}, nil
}

func (c *BloomClient) RestoreSession() (*LoginResult, error) {
	creds, err := c.loadCredentials()
	if err != nil {
		return nil, err
	}

	c.apiClient.SetToken(creds.Token)

	c.updateUser(creds.UserJSON)

	return &LoginResult{
		UserJSON: creds.UserJSON,
	}, nil
}
