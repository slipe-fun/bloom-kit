package client

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"

	"github.com/slipe-fun/skid-v4/pkg/identity"
	"github.com/tink-crypto/tink-go/v2/subtle/random"
)

type RegisterResult struct {
	Token       string
	UserJSON    []byte
	RecoveryKey string
}

type LoginResult struct {
	Token    string
	UserJSON []byte
}

func (c *BloomClient) Register() (*RegisterResult, error) {
	userIdentity, secret, err := identity.GenerateIdentity()
	if err != nil {
		return nil, err
	}
	defer secret.Wipe()

	masterKey := random.GetRandomBytes(32)
	recoveryKey := random.GetRandomBytes(32)

	registerResponse, err := c.authManager.Register(context.Background(), userIdentity, secret, masterKey, recoveryKey)
	if err != nil {
		panic(err)
	}

	mappedSecretKeys, err := mapSecretKeys(secret)
	if err != nil {
		return nil, err
	}

	userBytes, err := json.Marshal(registerResponse.User)
	if err != nil {
		return nil, err
	}

	err = c.saveCredentials(&SavedCredentials{
		UserID:      registerResponse.User.ID,
		RecoveryKey: recoveryKey,
		MasterKey:   masterKey,
		PublicKeys: *mapPublicKeys(
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

	c.apiClient.SetToken(registerResponse.Token)

	return &RegisterResult{
		Token:       registerResponse.Token,
		UserJSON:    userBytes,
		RecoveryKey: hex.EncodeToString(recoveryKey),
	}, nil
}

func (c *BloomClient) Login(userID, recoveryKey string) (*LoginResult, error) {
	beginLoginResponse, err := c.authManager.BeginLogin(context.Background(), userID)
	if err != nil {
		return nil, err
	}

	recoveryKeyBytes, err := hex.DecodeString(recoveryKey)
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
		ID: userID,
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

	mappedSecretKeys, err := mapSecretKeys(secretKeys)
	if err != nil {
		return nil, err
	}

	err = c.saveCredentials(&SavedCredentials{
		UserID:      finishLoginResult.User.ID,
		RecoveryKey: recoveryKeyBytes,
		MasterKey:   masterKey,
		PublicKeys: *mapPublicKeys(
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

	c.apiClient.SetToken(finishLoginResult.Token)

	return &LoginResult{
		Token:    finishLoginResult.Token,
		UserJSON: userBytes,
	}, nil
}

func (c *BloomClient) RestoreSession() (*LoginResult, error) {
	creds, err := c.loadCredentials()
	if err != nil {
		return nil, err
	}

	c.apiClient.SetToken(creds.Token)

	return &LoginResult{
		Token:    creds.Token,
		UserJSON: creds.UserJSON,
	}, nil
}
