package domain

type PublicKeys struct {
	MlKem768 string `json:"ml_kem768_public_key"`
	X448     string `json:"x448_public_key"`
	Ed448    string `json:"ed448_public_key"`
}

type SavedCredentials struct {
	UserID      string `json:"user_id"`
	RecoveryKey []byte `json:"recovery_key"`
	MasterKey   []byte `json:"master_key"`
	PublicKeys
	SecretKeys []byte `json:"secret_keys"`
	UserJSON   []byte `json:"user_json"`
	Token      string `json:"token"`
}
