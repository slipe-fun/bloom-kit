package mappers

import (
	"encoding/base64"

	"github.com/slipe-fun/skid-v4/pkg/messages"
)

func ConvertRawMessageToEncryptedMessage(ciphertext, nonce, salt string) (*messages.EncryptedMessage, error) {
	ciphertextBytes, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return nil, err
	}

	nonceBytes, err := base64.StdEncoding.DecodeString(nonce)
	if err != nil {
		return nil, err
	}

	saltBytes, err := base64.StdEncoding.DecodeString(salt)
	if err != nil {
		return nil, err
	}

	return &messages.EncryptedMessage{
		Ciphertext: ciphertextBytes,
		Nonce:      nonceBytes,
		Salt:       saltBytes,
	}, nil
}
