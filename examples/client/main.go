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

func (h *MessagesHandler) OnNewMessage(messageJSON []byte) {
	fmt.Println("New message:", string(messageJSON))
	h.client.SendMessage(84, nil, "hii")
}

func main() {
	key := random.GetRandomBytes(32)
	// key, err := hex.DecodeString("")
	// if err != nil {
	// 	panic(err)
	// }

	client := client.NewClient("https://api.bloomapp.pw/", "wss://api.bloomapp.pw/ws", "./storage", key)
	if err := client.Initialize(); err != nil {
		panic(err)
	}

	userJSON, err := client.GetMe()
	if err != nil {
		panic(err)
	}

	var user *domain.User
	if err = json.Unmarshal(userJSON, &user); err != nil {
		panic(err)
	}

	fmt.Println(user.ID)

	client.RegisterMessagesListener(&MessagesHandler{
		client: client,
	})

	roomID, fingerprint, err := client.StartExchangeSession("push")
	if err != nil {
		panic(err)
	}

	fmt.Println(roomID, fingerprint)

	recoveryKey, err := client.Exchange("push", roomID, fingerprint)
	if err != nil {
		panic(err)
	}

	fmt.Println(recoveryKey)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	<-stop
}
