package crypto

import (
	"crypto/hkdf"
	"crypto/sha256"
)

func HKDF(secret, salt []byte, info string, length int) ([]byte, error) {
	return hkdf.Key(sha256.New, secret, salt, info, length)
}
