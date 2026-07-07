package main

import (
	"context"
	"encoding/hex"
	"fmt"

	"github.com/slipe-fun/bloom-kit/api"
	authClient "github.com/slipe-fun/bloom-kit/api/auth"
	chatClient "github.com/slipe-fun/bloom-kit/api/chat"
	messageClient "github.com/slipe-fun/bloom-kit/api/message"
	userClient "github.com/slipe-fun/bloom-kit/api/user"
	authManager "github.com/slipe-fun/bloom-kit/managers/auth"
	chatManager "github.com/slipe-fun/bloom-kit/managers/chat"
	messageManager "github.com/slipe-fun/bloom-kit/managers/message"
	userManager "github.com/slipe-fun/bloom-kit/managers/user"
	"github.com/slipe-fun/bloom-kit/mappers"
	"github.com/slipe-fun/skid-v4/pkg/identity"
	"github.com/tink-crypto/tink-go/v2/subtle/random"
)

func main() {
	client := api.NewClient("https://api.bloomapp.pw")

	authClient := authClient.NewAuthClient(client)
	userClient := userClient.NewUserClient(client)
	chatClient := chatClient.NewChatClient(client)
	messageClient := messageClient.NewMessageClient(client)

	authManager := authManager.NewAuthManager(authClient)
	userManager := userManager.NewUserManager(userClient)
	chatManager := chatManager.NewChatManager(chatClient)
	messageManager := messageManager.NewMessageManager(messageClient)

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

	fmt.Println("Sender user ID:", registerResponse.User.ID)

	receiver, err := userManager.Get(ctx, "8SmiAxnunJiW7x")
	if err != nil {
		panic(err)
	}

	fmt.Println("Receiver user ID:", receiver.ID)

	receiverIdentity := mappers.ConvertUserToIdentity(receiver)

	createdChat, chatKey, syncKey, err := chatManager.Create(ctx, user, receiverIdentity, secret)
	if err != nil {
		panic(err)
	}

	fmt.Println("Created chat ID:", createdChat.ID)
	fmt.Println("Created chat key:", hex.EncodeToString(chatKey))
	fmt.Println("Created chat sync key:", hex.EncodeToString(syncKey))

	message, err := messageManager.Send(ctx, "hi chat", createdChat.ID, nil, chatKey, syncKey, user, receiverIdentity)
	if err != nil {
		panic(err)
	}

	fmt.Println("Created message ID:", message.ID)
	fmt.Println("Created message content:", string(message.Content))
	fmt.Println("Created message author ID:", message.AuthorID)

	// fmt.Println()

	// beginLoginResponse, err := authManager.BeginLogin(ctx, registerResponse.User.ID)
	// if err != nil {
	// 	panic(err)
	// }

	// fmt.Println("Challenge:", beginLoginResponse.Challenge)

	// fmt.Println()

	// finishLoginResponse, decryptedMasterKey, decryptedSecret, err := authManager.FinishLogin(ctx, beginLoginResponse, user, recoveryKey)
	// if err != nil {
	// 	panic(err)
	// }
	// defer decryptedSecret.Wipe()

	// client.SetToken(finishLoginResponse.Token)

	// fmt.Println("Login user ID:", finishLoginResponse.User.ID)
	// fmt.Println("Login token:", finishLoginResponse.Token)

	// fmt.Println()

	// fmt.Println("Decrypted master key:", hex.EncodeToString(decryptedMasterKey))

	// fmt.Println()

	// me, err := userClient.GetMe(ctx)
	// if err != nil {
	// 	panic(err)
	// }

	// fmt.Println("Current username:", me.Username)

	// searchResults, err := userClient.Search(ctx, "a")
	// if err != nil {
	// 	panic(err)
	// }

	// fmt.Println("Search results:", len(*searchResults))

	// fmt.Println()

	// newDisplayName := "hi"
	// editedUser, err := userManager.Edit(ctx, nil, &newDisplayName, nil)
	// if err != nil {
	// 	panic(err)
	// }

	// fmt.Println("Edited user username:", editedUser.User.Username)
	// fmt.Println("Edited user display name:", editedUser.User.DisplayName)
	// fmt.Println("Edited user description:", editedUser.User.Description)

	// fmt.Println()

	// getUserResponse, err := userClient.Get(ctx, "5FAwMKQUEYzAE4")
	// if err != nil {
	// 	panic(err)
	// }

	// fmt.Println("User username:", getUserResponse.Username)
	// fmt.Println("User display name:", getUserResponse.DisplayName)
	// fmt.Println("User description:", getUserResponse.Description)

	// fmt.Println()

	// userKeysResponse, err := userClient.GetKeys(ctx, "master")
	// if err != nil {
	// 	panic(err)
	// }

	// fmt.Println(userKeysResponse)
}
