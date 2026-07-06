package mappers

import (
	"encoding/base64"

	"github.com/slipe-fun/bloom-kit/domain"
	"github.com/slipe-fun/skid-v4/pkg/identity"
)

func MapPublicKeys(mlKem768, x448, ed448 []byte) *domain.PublicKeys {
	return &domain.PublicKeys{
		MlKem768: base64.StdEncoding.EncodeToString(mlKem768),
		X448:     base64.StdEncoding.EncodeToString(x448),
		Ed448:    base64.StdEncoding.EncodeToString(ed448),
	}
}

func MapSecretKeys(secretKeys *identity.SecretKeys) (*[]byte, error) {
	packSecretKeys, err := secretKeys.Pack()
	if err != nil {
		return nil, err
	}

	return &packSecretKeys, nil
}

func UnmapSecretKeys(mappedSecretKeys []byte) (*identity.SecretKeys, error) {
	return identity.Unpack(mappedSecretKeys)
}

func UnmapPublicKeys(publicKeys *domain.PublicKeys) (*identity.PublicKeys, error) {
	mlKem768Bytes, err := base64.StdEncoding.DecodeString(publicKeys.MlKem768)
	if err != nil {
		return nil, err
	}

	x448Bytes, err := base64.StdEncoding.DecodeString(publicKeys.X448)
	if err != nil {
		return nil, err
	}

	ed448Bytes, err := base64.StdEncoding.DecodeString(publicKeys.Ed448)
	if err != nil {
		return nil, err
	}

	return &identity.PublicKeys{
		MlKem768: mlKem768Bytes,
		X448:     x448Bytes,
		Ed448:    ed448Bytes,
	}, nil
}

func ConvertUserToIdentity(user *domain.User) *identity.User {
	publicKeys, err := UnmapPublicKeys(&domain.PublicKeys{
		MlKem768: user.MlKemPublicKey,
		X448:     user.EcdhPublicKey,
		Ed448:    user.EdPublicKey,
	})
	if err != nil {
		return nil
	}

	return &identity.User{
		ID: user.ID,
		PublicKeys: identity.PublicKeys{
			MlKem768: publicKeys.MlKem768,
			X448:     publicKeys.X448,
			Ed448:    publicKeys.Ed448,
		},
	}
}
