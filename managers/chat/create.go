package chat

import (
	"context"
	"strconv"

	"github.com/slipe-fun/bloom-kit/crypto"
	"github.com/slipe-fun/bloom-kit/domain"
	"github.com/slipe-fun/bloom-kit/mappers"
	"github.com/slipe-fun/skid-v4/pkg/identity"
	"github.com/slipe-fun/skid-v4/pkg/messages"
	"github.com/tink-crypto/tink-go/v2/subtle/random"
)

func (c *ChatManager) Create(
	ctx context.Context,
	sender *identity.User,
	receiver *identity.User,
	secretKeys *identity.SecretKeys,
) (*domain.Chat, []byte, []byte, error) {
	handshake, chatKey, syncKey, err := identity.InitiateKeyExchange(sender, secretKeys, receiver)
	if err != nil {
		return nil, nil, nil, err
	}

	createChatResponse, err := c.chatClient.CreatePrivateChat(ctx, receiver.ID, handshake)
	if err != nil {
		return nil, nil, nil, err
	}

	return createChatResponse, chatKey, syncKey, nil
}

func (c *ChatManager) CreateGroup(
	ctx context.Context,
	sender *identity.User,
	secretKeys *identity.SecretKeys,
	title string,
	users *[]domain.User,
) (*domain.Chat, []byte, error) {
	groupKey := random.GetRandomBytes(32)

	var members []domain.GroupMemberRequest

	for i := range *users {
		users := *users
		user := users[i]

		recipientIdentity := mappers.ConvertUserToIdentity(&user)
		if recipientIdentity == nil {
			continue
		}

		handshake, chatKey, syncKey, err := identity.InitiateKeyExchange(sender, secretKeys, recipientIdentity)
		if err != nil {
			continue
		}

		mappedHandshake := mappers.MapHandshake(handshake)

		_ = mappedHandshake
		member := domain.GroupMemberRequest{
			MemberID:  user.ID,
			Handshake: *mappedHandshake,
		}

		encryptedGroupKey, err := messages.Encrypt(chatKey, groupKey, syncKey, sender, recipientIdentity)
		if err != nil {
			continue
		}

		member.EncryptedGroupKey = *mappers.MapEncryptedGroupKey(encryptedGroupKey)

		members = append(members, member)
	}

	createGroupRequest, err := c.chatClient.CreateGroupChat(ctx, title, members)
	if err != nil {
		return nil, nil, err
	}

	groupEncryptionKey, err := crypto.HKDF(groupKey, []byte(strconv.Itoa(createGroupRequest.ID)), "skid:v4:group_encryption_key", 32)
	if err != nil {
		return nil, nil, err
	}

	return createGroupRequest, groupEncryptionKey, nil
}
