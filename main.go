package main

import (
	"context"
	"encoding/hex"
	"fmt"

	"github.com/slipe-fun/bloom-kit/api"
	authClient "github.com/slipe-fun/bloom-kit/api/auth"
	authManager "github.com/slipe-fun/bloom-kit/auth"
	"github.com/slipe-fun/skid-v4/pkg/identity"
	"github.com/tink-crypto/tink-go/v2/subtle/random"
)

func main() {
	client := api.NewClient("https://api.bloomapp.pw")

	authClient := authClient.NewUserClient(client)

	authManager := authManager.NewAuthManager(authClient)

	ctx := context.Background()

	user, secret, err := identity.GenerateIdentity()
	if err != nil {
		panic(err)
	}
	defer secret.Wipe()

	masterKey := random.GetRandomBytes(32)
	recoveryKey := random.GetRandomBytes(32)

	registerResponse, err := authManager.Register(ctx, user, secret, masterKey, recoveryKey)
	if err != nil {
		panic(err)
	}

	fmt.Println("User ID:", registerResponse.User.ID)

	fmt.Println()

	beginLoginResponse, err := authManager.BeginLogin(ctx, registerResponse.User.ID)
	if err != nil {
		panic(err)
	}

	fmt.Println("Challenge:", beginLoginResponse.Challenge)

	fmt.Println()

	finishLoginResponse, decryptedMasterKey, decryptedSecret, err := authManager.FinishLogin(ctx, beginLoginResponse, user, secret, recoveryKey)
	if err != nil {
		panic(err)
	}
	defer decryptedSecret.Wipe()

	fmt.Println("Login user ID:", finishLoginResponse.User.ID)
	fmt.Println("Login token:", finishLoginResponse.Token)

	fmt.Println()

	fmt.Println("Decrypted master key:", hex.EncodeToString(decryptedMasterKey))
}
