package main

import (
	"context"
	"encoding/hex"
	"fmt"

	"github.com/slipe-fun/bloom-kit/api"
	authClient "github.com/slipe-fun/bloom-kit/api/auth"
	userClient "github.com/slipe-fun/bloom-kit/api/user"
	authManager "github.com/slipe-fun/bloom-kit/managers/auth"
	userManager "github.com/slipe-fun/bloom-kit/managers/user"
	"github.com/slipe-fun/skid-v4/pkg/identity"
	"github.com/tink-crypto/tink-go/v2/subtle/random"
)

func main() {
	client := api.NewClient("https://api.bloomapp.pw")

	authClient := authClient.NewAuthClient(client)
	userClient := userClient.NewUserClient(client)

	authManager := authManager.NewAuthManager(authClient)
	userManager := userManager.NewUserManager(userClient)

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

	client.SetToken(registerResponse.Token)

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

	client.SetToken(finishLoginResponse.Token)

	fmt.Println("Login user ID:", finishLoginResponse.User.ID)
	fmt.Println("Login token:", finishLoginResponse.Token)

	fmt.Println()

	fmt.Println("Decrypted master key:", hex.EncodeToString(decryptedMasterKey))

	fmt.Println()

	me, err := userClient.GetMe(ctx)
	if err != nil {
		panic(err)
	}

	fmt.Println("Current username:", me.Username)

	searchResults, err := userClient.Search(ctx, "a")
	if err != nil {
		panic(err)
	}

	fmt.Println("Search results:", len(*searchResults))

	fmt.Println()

	newDisplayName := "hi"
	editedUser, err := userManager.Edit(ctx, nil, &newDisplayName, nil)
	if err != nil {
		panic(err)
	}

	fmt.Println("Edited user username:", editedUser.User.Username)
	fmt.Println("Edited user display name:", editedUser.User.DisplayName)
	fmt.Println("Edited user description:", editedUser.User.Description)

	fmt.Println()

	getUserResponse, err := userClient.Get(ctx, "5FAwMKQUEYzAE4")
	if err != nil {
		panic(err)
	}

	fmt.Println("User username:", getUserResponse.Username)
	fmt.Println("User display name:", getUserResponse.DisplayName)
	fmt.Println("User description:", getUserResponse.Description)

	fmt.Println()

	userKeysResponse, err := userClient.GetKeys(ctx, "master")
	if err != nil {
		panic(err)
	}

	fmt.Println(userKeysResponse)
}
