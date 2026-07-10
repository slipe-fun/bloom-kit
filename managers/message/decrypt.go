package message

import (
	"github.com/slipe-fun/bloom-kit/domain"
	"github.com/slipe-fun/bloom-kit/mappers"
	"github.com/slipe-fun/skid-v4/pkg/identity"
	message "github.com/slipe-fun/skid-v4/pkg/messages"
)

func (m *MessageManager) DecryptMessage(
	msg *domain.RawMessageWithReply,
	encryptionKey []byte,
	syncKey []byte,
	userIdentity *identity.User,
	recipientIdentity *identity.User,
) (*message.Message, error) {
	convertedMessage, err := mappers.ConvertRawMessageToEncryptedMessage(msg.Ciphertext, msg.Nonce, msg.Salt)
	if err != nil {
		return nil, err
	}

	return message.Decrypt(encryptionKey, *convertedMessage, syncKey, userIdentity, recipientIdentity)
}
