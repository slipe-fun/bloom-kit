package main

import (
	"context"
	"fmt"

	"github.com/slipe-fun/bloom-kit/api"
	"github.com/slipe-fun/bloom-kit/api/user"
	"github.com/slipe-fun/skid-v4/pkg/identity"
	"github.com/tink-crypto/tink-go/v2/subtle/random"
)

func main() {
	client := api.NewClient("https://api.bloomapp.pw")

	userClient := user.NewUserClient(client)

	ctx := context.Background()

	userIdentity, secret, err := identity.GenerateIdentity()
	if err != nil {
		panic(err)
	}
	defer secret.Wipe()

	masterKey := random.GetRandomBytes(32)
	recoveryKey := random.GetRandomBytes(32)

	encryptedSecretKeys, err := identity.EncryptSecretKeys(userIdentity, secret, masterKey)
	if err != nil {
		panic(err)
	}

	encryptedMasterKey, err := identity.EncryptMasterKey(masterKey, recoveryKey, userIdentity, secret)
	if err != nil {
		panic(err)
	}

	registerRequestBody := user.NewKeysRequest(userIdentity, encryptedSecretKeys, encryptedMasterKey)

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

	encryptedMasterKeyBytes, encryptedSecretKeysBytes, challenge, err := user.DecodeBeginLoginResponse(beginLoginResponse, userIdentity)
	if err != nil {
		panic(err)
	}

	decryptedMasterKey, err := identity.DecryptMasterKey(encryptedMasterKeyBytes, recoveryKey, userIdentity)
	if err != nil {
		panic(err)
	}

	decryptedSecretKeys, err := identity.DecryptSecretKeys(encryptedSecretKeysBytes, userIdentity, decryptedMasterKey)
	if err != nil {
		panic(err)
	}
	defer decryptedSecretKeys.Wipe()

	finishLoginRequest, err := user.NewFinishLoginRequest(userIdentity, secret, challenge)
	if err != nil {
		panic(err)
	}

	finishLoginResponse, err := userClient.FinishLogin(ctx, finishLoginRequest)
	if err != nil {
		panic(err)
	}

	fmt.Println(finishLoginResponse)
}
