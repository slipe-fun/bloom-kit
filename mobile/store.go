package mobile

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
	RecoveryKey string `json:"recovery_key"`
	MasterKey   string `json:"master_key"`
	PublicKeys
	SecretKeys string `json:"secret_keys"`
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

func mapSecretKeys(secretKeys *identity.SecretKeys) (*string, error) {
	packSecretKeys, err := secretKeys.Pack()
	if err != nil {
		return nil, err
	}

	encodedSecretKeys := base64.StdEncoding.EncodeToString(packSecretKeys)

	return &encodedSecretKeys, nil
}

func unmapSecretKeys(mappedSecretKeys string) (*identity.SecretKeys, error) {
	mappedSecretKeysBytes, err := base64.StdEncoding.DecodeString(mappedSecretKeys)
	if err != nil {
		return nil, err
	}

	return identity.Unpack(mappedSecretKeysBytes)
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

	if len(c.encryptionKey) != 32 {
		return errors.New("invalid encryption key")
	}

	ciphertext, err := crypto.Encrypt(c.encryptionKey, plainText, nil)
	if err != nil {
		return err
	}

	filePath := filepath.Join(c.storagePath, "session.dat")
	return os.WriteFile(filePath, ciphertext, 0600)
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

	var creds SavedCredentials
	err = json.Unmarshal(plainText, &creds)
	if err != nil {
		return nil, err
	}

	return &creds, nil
}

func (c *BloomClient) ClearCredentials() {
	filePath := filepath.Join(c.storagePath, "session.dat")

	_ = os.Remove(filePath)

	c.apiClient.SetToken("")
}
