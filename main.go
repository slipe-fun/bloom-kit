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

	registerRequestBody := domain.NewKeysRequest(user, encryptedSecretKeys, encryptedMasterKey)

	registerResponse, err := userClient.Register(ctx, registerRequestBody)
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

	encryptedMasterKeyBytes, encryptedSecretKeysBytes, challenge, err := domain.DecodeBeginLoginResponse(beginLoginResponse, user)
	if err != nil {
		panic(err)
	}

	decryptedMasterKey, err := identity.DecryptMasterKey(encryptedMasterKeyBytes, recoveryKey, user)
	if err != nil {
		panic(err)
	}

	decryptedSecretKeys, err := identity.DecryptSecretKeys(encryptedSecretKeysBytes, user, decryptedMasterKey)
	if err != nil {
		panic(err)
	}
	defer decryptedSecretKeys.Wipe()

	message, err := json.Marshal(challenge)
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
