package crypto

import (
	"errors"
	"runtime"

	"github.com/tink-crypto/tink-go/v2/aead/subtle"
)

func NewAes(key []byte) (*subtle.AESGCMSIV, error) {
	if len(key) != 16 && len(key) != 32 {
		return nil, errors.New("invalid AES key size: must be exactly 16 or 32 bytes")
	}

	localKey := make([]byte, len(key))
	copy(localKey, key)

	defer func() {
		Zero(localKey)
	}()

	return subtle.NewAESGCMSIV(localKey)
}

func Encrypt(key, plaintext, aad []byte) ([]byte, error) {
	aes, err := NewAes(key)
	if err != nil {
		return nil, err
	}

	defer func() {
		if aes != nil {
			var zeroAES subtle.AESGCMSIV
			*aes = zeroAES
			runtime.KeepAlive(aes)
		}
	}()

	fullResult, err := aes.Encrypt(plaintext, aad)
	if err != nil {
		return nil, err
	}

	if len(fullResult) < 12 {
		return nil, errors.New("failed to encrypt: output too short")
	}

	defer func() {
		Zero(fullResult)
	}()

	return fullResult, nil
}

func Decrypt(key, fullResult, aad []byte) ([]byte, error) {
	ciphertext := make([]byte, len(fullResult)-12)
	iv := make([]byte, 12)
	copy(iv, fullResult[:12])
	copy(ciphertext, fullResult[12:])

	if len(iv) != 12 {
		return nil, errors.New("invalid IV length: must be exactly 12 bytes")
	}

	if len(ciphertext) < 16 {
		return nil, errors.New("ciphertext too short: must be at least 16 bytes (tag size)")
	}

	aes, err := NewAes(key)
	if err != nil {
		return nil, err
	}

	defer func() {
		if aes != nil {
			var zeroAES subtle.AESGCMSIV
			*aes = zeroAES
			runtime.KeepAlive(aes)
		}
	}()

	fullCiphertext := make([]byte, 0, len(iv)+len(ciphertext))
	fullCiphertext = append(fullCiphertext, iv...)
	fullCiphertext = append(fullCiphertext, ciphertext...)

	defer func() {
		Zero(fullCiphertext)
	}()

	return aes.Decrypt(fullCiphertext, aad)
}
