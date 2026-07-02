package mobile

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
	Token     string
	UserJSON  []byte
	MasterKey string
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

	userBytes, err := json.Marshal(registerResponse.User)
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

	c.apiClient.SetToken(finishLoginResult.Token)

	return &LoginResult{
		Token:     finishLoginResult.Token,
		UserJSON:  userBytes,
		MasterKey: hex.EncodeToString(masterKey),
	}, nil
}
