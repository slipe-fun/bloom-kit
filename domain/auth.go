package domain

type KeysRequest struct {
	AuthLookupID       string              `json:"auth_lookup_id"`
	IdentityKeys       IdentityKeysRequest `json:"identity_keys"`
	EncryptedMasterKey EncryptedKey        `json:"encrypted_master_key"`
}

type RegisterResponse struct {
	Token   string  `json:"token"`
	User    User    `json:"user"`
	Session Session `json:"session"`
}

type BeginLoginResponse struct {
	UserID    string      `json:"user_id"`
	Keys      KeysRequest `json:"keys"`
	Challenge string      `json:"challenge"`
}

type LoginChallenge struct {
	Challenge string `json:"challenge"`
	UserID    string `json:"user_id"`
}

type FinishLoginRequest struct {
	UserID    string `json:"user_id"`
	Signature string `json:"signature"`
}
