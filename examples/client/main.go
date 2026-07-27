package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/slipe-fun/bloom-kit/client"
	"github.com/slipe-fun/bloom-kit/domain"
	"github.com/tink-crypto/tink-go/v2/subtle/random"
)

type MessagesHandler struct {
	client *client.BloomClient
}

type UserHandler struct{}

func (h *MessagesHandler) OnNewMessage(messageJSON []byte) {
	fmt.Println("New message:", string(messageJSON))
}

func (h *UserHandler) OnUserUpdated(userJSON []byte) {
	var user *domain.User
	if err := json.Unmarshal(userJSON, &user); err != nil {
		return
	}

	fmt.Println("User updated")
	fmt.Println("User id:", user.ID)
	fmt.Println("User username:", user.Username)
	fmt.Println("User display name:", user.DisplayName)
	fmt.Println("User description:", user.Description)
}

func main() {
	key := random.GetRandomBytes(32)
	// key, err := hex.DecodeString("")
	// if err != nil {
	// 	panic(err)
	// }

	bloomClient := client.NewClient("https://api.bloomapp.pw/", "wss://api.bloomapp.pw/ws", "./storage", key)
	if err := bloomClient.Initialize(); err != nil {
		panic(err)
	}

	bloomClient.RegisterMessagesListener(&MessagesHandler{
		client: bloomClient,
	})

	bloomClient.RegisterUserListener(&UserHandler{})

	registerResponse, err := bloomClient.Register()
	// recoveryKey := ""
	// a, err := bloomClient.Login(recoveryKey)

	// userJSON, err := bloomClient.GetMe()
	// if err != nil {
	// 	panic(err)
	// }

	var user *domain.User
	if err := json.Unmarshal(registerResponse.UserJSON, &user); err != nil {
		panic(err)
	}

	fmt.Println(user.ID)

	editUserRequest := &client.EditRequest{}
	editUserRequest.Username = "coolusername"
	editUserRequest.HasUsername = true
	editUserRequest.DisplayName = "cooldisplay"
	editUserRequest.HasDisplayName = true
	editUserRequest.Description = "description description coool"
	editUserRequest.HasDescription = true

	_, err = bloomClient.EditUser(editUserRequest)
	if err != nil {
		panic(err)
	}

	// roomID, fingerprint, err := client.StartExchangeSession("push")
	// if err != nil {
	// 	panic(err)
	// }

	// fmt.Println(roomID, fingerprint)

	// recoveryKey, err := client.Exchange("push", roomID, fingerprint)
	// if err != nil {
	// 	panic(err)
	// }

	// fmt.Println(recoveryKey)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	<-stop
}
