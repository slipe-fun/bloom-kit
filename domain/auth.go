package domain

import (
	"encoding/base64"

	"github.com/slipe-fun/skid-v4/pkg/identity"
)

type KeysRequest struct {
	IdentityKeys       IdentityKeysRequest `json:"identity_keys"`
	EncryptedMasterKey EncryptedKey        `json:"encrypted_master_key"`
}

type RegisterResponse struct {
	Token   string  `json:"token"`
	User    User    `json:"user"`
	Session Session `json:"session"`
}

type BeginLoginResponse struct {
	Keys      KeysRequest `json:"keys"`
	Challenge string      `json:"challenge"`
}

type LoginChallenge struct {
	Challenge string `json:"challenge"`
	UserID    string `json:"user_id"`
}

type FinishLoginRequest struct {
	UserID    string `json:"user_id"`
	Signature string `json:"signature"`
}

func NewKeysRequest(
	user *identity.User,
	encryptedSecretKeys *identity.EncryptedSecretKeys,
	encryptedMasterKey *identity.EncryptedMasterKey,
) *KeysRequest {
	return &KeysRequest{
		IdentityKeys: IdentityKeysRequest{
			EncryptedSecretKeys: EncryptedKey{
				Ciphertext: base64.StdEncoding.EncodeToString(encryptedSecretKeys.Ciphertext),
				Nonce:      base64.StdEncoding.EncodeToString(encryptedSecretKeys.Nonce),
				Salt:       base64.StdEncoding.EncodeToString(encryptedSecretKeys.Salt),
				Signature:  base64.StdEncoding.EncodeToString(encryptedSecretKeys.Signature),
			},
			IdentityPublicKeys: IdentityPublicKeys{
				MlKemPublicKey: base64.StdEncoding.EncodeToString(user.PublicKeys.MlKem768),
				EcdhPublicKey:  base64.StdEncoding.EncodeToString(user.PublicKeys.X448),
				EdPublicKey:    base64.StdEncoding.EncodeToString(user.PublicKeys.Ed448),
			},
		},
		EncryptedMasterKey: EncryptedKey{
			Ciphertext: base64.StdEncoding.EncodeToString(encryptedMasterKey.Ciphertext),
			Nonce:      base64.StdEncoding.EncodeToString(encryptedMasterKey.Nonce),
			Salt:       base64.StdEncoding.EncodeToString(encryptedMasterKey.Salt),
			Signature:  base64.StdEncoding.EncodeToString(encryptedMasterKey.Signature),
		},
	}
}

func DecodeBeginLoginResponse(response *BeginLoginResponse, user *identity.User) (*identity.EncryptedMasterKey, *identity.EncryptedSecretKeys, *LoginChallenge, error) {
	masterKeyCiphertext, err := base64.StdEncoding.DecodeString(response.Keys.EncryptedMasterKey.Ciphertext)
	if err != nil {
		return nil, nil, nil, err
	}

	masterKeyNonce, err := base64.StdEncoding.DecodeString(response.Keys.EncryptedMasterKey.Nonce)
	if err != nil {
		return nil, nil, nil, err
	}

	masterKeySalt, err := base64.StdEncoding.DecodeString(response.Keys.EncryptedMasterKey.Salt)
	if err != nil {
		return nil, nil, nil, err
	}

	masterKeySignature, err := base64.StdEncoding.DecodeString(response.Keys.EncryptedMasterKey.Signature)
	if err != nil {
		return nil, nil, nil, err
	}

	secretKeysCiphertext, err := base64.StdEncoding.DecodeString(response.Keys.IdentityKeys.EncryptedSecretKeys.Ciphertext)
	if err != nil {
		return nil, nil, nil, err
	}

	secretKeysNonce, err := base64.StdEncoding.DecodeString(response.Keys.IdentityKeys.EncryptedSecretKeys.Nonce)
	if err != nil {
		return nil, nil, nil, err
	}

	secretKeysSalt, err := base64.StdEncoding.DecodeString(response.Keys.IdentityKeys.EncryptedSecretKeys.Salt)
	if err != nil {
		return nil, nil, nil, err
	}

	secretKeysSignature, err := base64.StdEncoding.DecodeString(response.Keys.IdentityKeys.EncryptedSecretKeys.Signature)
	if err != nil {
		return nil, nil, nil, err
	}

	encryptedMasterKeyBytes := identity.EncryptedMasterKey{
		Ciphertext: masterKeyCiphertext,
		Nonce:      masterKeyNonce,
		Salt:       masterKeySalt,
		Signature:  masterKeySignature,
	}

	encryptedSecretKeysBytes := identity.EncryptedSecretKeys{
		Ciphertext: secretKeysCiphertext,
		Nonce:      secretKeysNonce,
		Salt:       secretKeysSalt,
		Signature:  secretKeysSignature,
	}

	challenge := LoginChallenge{
		Challenge: response.Challenge,
		UserID:    user.ID,
	}

	return &encryptedMasterKeyBytes, &encryptedSecretKeysBytes, &challenge, nil
}
