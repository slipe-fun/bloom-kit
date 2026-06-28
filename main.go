package main

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/slipe-fun/bloom-kit/api"
	"github.com/slipe-fun/bloom-kit/api/user"
	"github.com/slipe-fun/bloom-kit/domain"
	"github.com/slipe-fun/skid-v4/pkg/identity"
	"github.com/tink-crypto/tink-go/v2/subtle/random"
)

func main() {
	client := api.NewClient("https://api.bloomapp.pw")

	userClient := user.NewUserClient(client)

	ctx := context.Background()

	user, secret, err := identity.GenerateIdentity()
	if err != nil {
		panic(err)
	}
	defer secret.Wipe()

	masterKey := random.GetRandomBytes(32)
	recoveryKey := random.GetRandomBytes(32)

	encryptedSecretKeys, err := identity.EncryptSecretKeys(user, secret, masterKey)
	if err != nil {
		panic(err)
	}

	encryptedMasterKey, err := identity.EncryptMasterKey(masterKey, recoveryKey, user, secret)
	if err != nil {
		panic(err)
	}

	registerRequestBody := domain.KeysRequest{
		IdentityKeys: domain.IdentityKeysRequest{
			EncryptedSecretKeys: domain.EncryptedKey{
				Ciphertext: base64.StdEncoding.EncodeToString(encryptedSecretKeys.Ciphertext),
				Nonce:      base64.StdEncoding.EncodeToString(encryptedSecretKeys.Nonce),
				Salt:       base64.StdEncoding.EncodeToString(encryptedSecretKeys.Salt),
				Signature:  base64.StdEncoding.EncodeToString(encryptedSecretKeys.Signature),
			},
			IdentityPublicKeys: domain.IdentityPublicKeys{
				MlKemPublicKey: base64.StdEncoding.EncodeToString(user.PublicKeys.MlKem768),
				EcdhPublicKey:  base64.StdEncoding.EncodeToString(user.PublicKeys.X448),
				EdPublicKey:    base64.StdEncoding.EncodeToString(user.PublicKeys.Ed448),
			},
		},
		EncryptedMasterKey: domain.EncryptedKey{
			Ciphertext: base64.StdEncoding.EncodeToString(encryptedMasterKey.Ciphertext),
			Nonce:      base64.StdEncoding.EncodeToString(encryptedMasterKey.Nonce),
			Salt:       base64.StdEncoding.EncodeToString(encryptedMasterKey.Salt),
			Signature:  base64.StdEncoding.EncodeToString(encryptedMasterKey.Signature),
		},
	}

	registerResponse, err := userClient.Register(ctx, &registerRequestBody)
	if err != nil {
		panic(err.Error())
	}

	fmt.Println(registerResponse)
}
