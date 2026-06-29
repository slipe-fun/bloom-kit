package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/cloudflare/circl/sign/ed448"
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
		panic(err)
	}

	fmt.Println(registerResponse)

	fmt.Println(registerResponse.User.ID)

	beginLoginResponse, err := userClient.BeginLogin(ctx, registerResponse.User.ID)
	if err != nil {
		panic(err)
	}

	fmt.Println(beginLoginResponse)

	masterKeyCiphertext, err := base64.StdEncoding.DecodeString(beginLoginResponse.Keys.EncryptedMasterKey.Ciphertext)
	if err != nil {
		panic(err)
	}

	masterKeyNonce, err := base64.StdEncoding.DecodeString(beginLoginResponse.Keys.EncryptedMasterKey.Nonce)
	if err != nil {
		panic(err)
	}

	masterKeySalt, err := base64.StdEncoding.DecodeString(beginLoginResponse.Keys.EncryptedMasterKey.Salt)
	if err != nil {
		panic(err)
	}

	masterKeySignature, err := base64.StdEncoding.DecodeString(beginLoginResponse.Keys.EncryptedMasterKey.Signature)
	if err != nil {
		panic(err)
	}

	decryptedMasterKey, err := identity.DecryptMasterKey(&identity.EncryptedMasterKey{
		Ciphertext: masterKeyCiphertext,
		Nonce:      masterKeyNonce,
		Salt:       masterKeySalt,
		Signature:  masterKeySignature,
	}, recoveryKey, user)
	if err != nil {
		panic(err)
	}

	secretKeysCiphertext, err := base64.StdEncoding.DecodeString(beginLoginResponse.Keys.IdentityKeys.EncryptedSecretKeys.Ciphertext)
	if err != nil {
		panic(err)
	}

	secretKeysNonce, err := base64.StdEncoding.DecodeString(beginLoginResponse.Keys.IdentityKeys.EncryptedSecretKeys.Nonce)
	if err != nil {
		panic(err)
	}

	secretKeysSalt, err := base64.StdEncoding.DecodeString(beginLoginResponse.Keys.IdentityKeys.EncryptedSecretKeys.Salt)
	if err != nil {
		panic(err)
	}

	secretKeysSignature, err := base64.StdEncoding.DecodeString(beginLoginResponse.Keys.IdentityKeys.EncryptedSecretKeys.Signature)
	if err != nil {
		panic(err)
	}

	decryptedSecretKeys, err := identity.DecryptSecretKeys(&identity.EncryptedSecretKeys{
		Ciphertext: secretKeysCiphertext,
		Nonce:      secretKeysNonce,
		Salt:       secretKeysSalt,
		Signature:  secretKeysSignature,
	}, user, decryptedMasterKey)
	if err != nil {
		panic(err)
	}
	defer decryptedSecretKeys.Wipe()

	data := domain.LoginChallenge{
		Challenge: beginLoginResponse.Challenge,
		UserID:    registerResponse.User.ID,
	}

	message, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}

	signature := ed448.Sign(decryptedSecretKeys.Ed448, message, "")

	finishLoginRequest := &domain.FinishLoginRequest{
		UserID:    registerResponse.User.ID,
		Signature: base64.StdEncoding.EncodeToString(signature),
	}

	finishLoginResponse, err := userClient.FinishLogin(ctx, finishLoginRequest)
	if err != nil {
		panic(err)
	}

	fmt.Println(finishLoginResponse)
}
