package chat

import (
	"encoding/base64"

	"github.com/slipe-fun/bloom-kit/domain"
	"github.com/slipe-fun/skid-v4/pkg/identity"
)

func NewCreateChatRequest(recipientID string, handshake *identity.HandshakePayload) *domain.CreateChatRequest {
	return &domain.CreateChatRequest{
		Recipient: recipientID,
		Handshake: *MapHandshake(handshake),
	}
}

func MapHandshake(handshake *identity.HandshakePayload) *domain.Handshake {
	return &domain.Handshake{
		ReceiverCipherText:  base64.StdEncoding.EncodeToString(handshake.ReceiverCiphertext),
		SenderEphemeralX448: base64.StdEncoding.EncodeToString(handshake.SenderEphemeralX448),
		EncryptedSyncKey: domain.EncryptedSyncKey{
			CipherText: base64.StdEncoding.EncodeToString(handshake.EncryptedSyncKey.Ciphertext),
			Nonce:      base64.StdEncoding.EncodeToString(handshake.EncryptedSyncKey.Nonce),
		},
	}
}

func DecodeHandshake(handshake *domain.Handshake) (*identity.HandshakePayload, error) {
	receiverCiphertextBytes, err := base64.StdEncoding.DecodeString(handshake.ReceiverCipherText)
	if err != nil {
		return nil, err
	}

	senderEphemeralX448Bytes, err := base64.StdEncoding.DecodeString(handshake.SenderEphemeralX448)
	if err != nil {
		return nil, err
	}

	encryptedSyncKeyCiphertextBytes, err := base64.StdEncoding.DecodeString(handshake.EncryptedSyncKey.CipherText)
	if err != nil {
		return nil, err
	}

	encryptedSyncKeyNonceBytes, err := base64.StdEncoding.DecodeString(handshake.EncryptedSyncKey.Nonce)
	if err != nil {
		return nil, err
	}

	return &identity.HandshakePayload{
		ReceiverCiphertext:  receiverCiphertextBytes,
		SenderEphemeralX448: senderEphemeralX448Bytes,
		EncryptedSyncKey: identity.EncryptedSyncKey{
			Ciphertext: encryptedSyncKeyCiphertextBytes,
			Nonce:      encryptedSyncKeyNonceBytes,
		},
	}, nil
}
