package chat

import (
	"github.com/slipe-fun/skid-v4/pkg/identity"
)

func (c *ChatManager) FinalizeHandshake(
	handshakePayload *identity.HandshakePayload,
	userIdentity *identity.User,
	secretKeys *identity.SecretKeys,
	recipientIdentity *identity.User,
) ([]byte, []byte, error) {
	chatKey, syncKey, err := identity.FinalizeKeyExchange(handshakePayload, userIdentity, secretKeys, recipientIdentity, nil, true)
	if err != nil {
		chatKey, syncKey, err = identity.FinalizeKeyExchange(handshakePayload, userIdentity, secretKeys, recipientIdentity, nil, false)
		if err != nil {
			return nil, nil, err
		}
	}

	return chatKey, syncKey, nil
}
