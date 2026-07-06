package client

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"

	"github.com/slipe-fun/bloom-kit/crypto"
	"github.com/slipe-fun/bloom-kit/domain"
)

func (c *BloomClient) saveCredentials(creds *domain.SavedCredentials) error {
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

func (c *BloomClient) loadCredentials() (*domain.SavedCredentials, error) {
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

	var creds domain.SavedCredentials
	err = json.Unmarshal(plainText, &creds)
	if err != nil {
		return nil, err
	}

	c.credentials = &domain.SavedCredentials{
		UserID: creds.UserID,
		PublicKeys: domain.PublicKeys{
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
