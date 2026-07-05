package domain

type EncryptedSyncKey struct {
	CipherText string `json:"ciphertext"`
	Nonce      string `json:"nonce"`
}

type Handshake struct {
	ReceiverCipherText  string           `json:"receiver_cipher_text"`
	SenderEphemeralX448 string           `json:"sender_ephemeral_x448"`
	EncryptedSyncKey    EncryptedSyncKey `json:"encrypted_sync_key"`
}

type RawChat struct {
	ID        int        `json:"id"`
	Members   []User     `json:"members"`
	Handshake *Handshake `json:"handshake"`
}

type Chat struct {
	RawChat
	LastMessage     *Message `json:"last_message,omitempty"`
	LastReadMessage *Message `json:"last_read_message,omitempty"`
}

type ChatWithKeys struct {
	RawChat
	ChatKey         []byte
	SyncKey         []byte
	LastMessage     *Message `json:"last_message,omitempty"`
	LastReadMessage *Message `json:"last_read_message,omitempty"`
}

type CreateChatRequest struct {
	Recipient string    `json:"recipient"`
	Handshake Handshake `json:"handshake"`
}
