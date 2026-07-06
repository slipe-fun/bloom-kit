package client

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"

	"github.com/slipe-fun/bloom-kit/crypto"
	"github.com/slipe-fun/skid-v4/pkg/identity"
)

type PublicKeys struct {
	MlKem768 string `json:"ml_kem768_public_key"`
	X448     string `json:"x448_public_key"`
	Ed448    string `json:"ed448_public_key"`
}

type SavedCredentials struct {
	UserID      string `json:"user_id"`
	RecoveryKey []byte `json:"recovery_key"`
	MasterKey   []byte `json:"master_key"`
	PublicKeys
	SecretKeys []byte `json:"secret_keys"`
	UserJSON   []byte `json:"user_json"`
	Token      string `json:"token"`
}

func mapPublicKeys(mlKem768, x448, ed448 []byte) *PublicKeys {
	return &PublicKeys{
		MlKem768: base64.StdEncoding.EncodeToString(mlKem768),
		X448:     base64.StdEncoding.EncodeToString(x448),
		Ed448:    base64.StdEncoding.EncodeToString(ed448),
	}
}

func mapSecretKeys(secretKeys *identity.SecretKeys) (*[]byte, error) {
	packSecretKeys, err := secretKeys.Pack()
	if err != nil {
		return nil, err
	}

	return &packSecretKeys, nil
}

func unmapSecretKeys(mappedSecretKeys []byte) (*identity.SecretKeys, error) {
	return identity.Unpack(mappedSecretKeys)
}

func unmapPublicKeys(publicKeys *PublicKeys) (*identity.PublicKeys, error) {
	mlKem768Bytes, err := base64.StdEncoding.DecodeString(publicKeys.MlKem768)
	if err != nil {
		return nil, err
	}

	x448Bytes, err := base64.StdEncoding.DecodeString(publicKeys.X448)
	if err != nil {
		return nil, err
	}

	ed448Bytes, err := base64.StdEncoding.DecodeString(publicKeys.Ed448)
	if err != nil {
		return nil, err
	}

	return &identity.PublicKeys{
		MlKem768: mlKem768Bytes,
		X448:     x448Bytes,
		Ed448:    ed448Bytes,
	}, nil
}

func (c *BloomClient) saveCredentials(creds *SavedCredentials) error {
	plainText, err := json.Marshal(creds)
	if err != nil {
		return err
	}
	defer crypto.Zero(plainText)

	if len(c.encryptionKey) != 32 {
		return errors.New("invalid encryption key")
	}

	ciphertext, err := crypto.Encrypt(c.encryptionKey, plainText, nil)
	if err != nil {
		return err
	}

	filePath := filepath.Join(c.storagePath, "session.dat")
	err = os.WriteFile(filePath, ciphertext, 0600)
	if err != nil {
		return err
	}

	crypto.Zero(creds.RecoveryKey)
	crypto.Zero(creds.MasterKey)
	crypto.Zero(creds.SecretKeys)

	c.credentials = creds

	return nil
}

func (c *BloomClient) loadCredentials() (*SavedCredentials, error) {
	filePath := filepath.Join(c.storagePath, "session.dat")

	ciphertext, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	if len(c.encryptionKey) != 32 {
		return nil, errors.New("invalid encryption key")
	}

	plainText, err := crypto.Decrypt(c.encryptionKey, ciphertext, nil)
	if err != nil {
		return nil, err
	}
	defer crypto.Zero(plainText)

	var creds SavedCredentials
	err = json.Unmarshal(plainText, &creds)
	if err != nil {
		return nil, err
	}

	c.credentials = &SavedCredentials{
		UserID: creds.UserID,
		PublicKeys: PublicKeys{
			MlKem768: creds.PublicKeys.MlKem768,
			X448:     creds.PublicKeys.X448,
			Ed448:    creds.PublicKeys.Ed448,
		},
		UserJSON: creds.UserJSON,
		Token:    creds.Token,
	}

	return &creds, nil
}

func (c *BloomClient) ClearCredentials() {
	filePath := filepath.Join(c.storagePath, "session.dat")

	_ = os.Remove(filePath)

	c.credentials = nil

	c.apiClient.SetToken("")
}
