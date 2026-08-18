package domain

import "time"

type EncryptedSyncKey struct {
	CipherText string `json:"ciphertext"`
	Nonce      string `json:"nonce"`
}

type Handshake struct {
	ReceiverCipherText  string           `json:"receiver_cipher_text"`
	SenderEphemeralX448 string           `json:"sender_ephemeral_x448"`
	EncryptedSyncKey    EncryptedSyncKey `json:"encrypted_sync_key"`
}

type EncryptedGroupKey struct {
	Ciphertext string `json:"ciphertext"`
	Nonce      string `json:"nonce"`
	Salt       string `json:"salt"`
}

type GroupMember struct {
	MemberID          string `json:"recipient_id"`
	InvitedByID       string `json:"invited_by_id"`
	Handshake         `json:"handshake"`
	EncryptedGroupKey `json:"encrypted_group_key"`
}

type RawChat struct {
	ID        int        `json:"id"`
	Members   *[]User    `json:"members,omitempty"`
	Handshake *Handshake `json:"handshake,omitempty"`
	Title     string     `json:"title,omitempty"`
	Type      string     `json:"type"`
	CreatedAt time.Time  `json:"created_at"`
}

type Chat struct {
	RawChat
	EncryptedLastMessage     *RawMessageWithReply `json:"last_message,omitempty"`
	LastMessage              *Message
	EncryptedLastReadMessage *RawMessageWithReply `json:"last_read_message,omitempty"`
	LastReadMessage          *Message
}

type ChatWithKeys struct {
	RawChat
	ChatKey         []byte
	SyncKey         []byte
	LastMessage     *Message `json:"last_message,omitempty"`
	LastReadMessage *Message `json:"last_read_message,omitempty"`
}

type GroupMemberRequest struct {
	MemberID          string            `json:"member_id"`
	Handshake         Handshake         `json:"handshake"`
	EncryptedGroupKey EncryptedGroupKey `json:"encrypted_group_key"`
}

type CreateChatRequest struct {
	Type string `json:"type"`

	Recipient string     `json:"recipient,omitempty"`
	Handshake *Handshake `json:"handshake,omitempty"`

	Title   string               `json:"title,omitempty"`
	Members []GroupMemberRequest `json:"members,omitempty"`
}
