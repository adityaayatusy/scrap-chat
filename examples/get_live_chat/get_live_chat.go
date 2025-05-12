package main

import (
	"fmt"
	"github.com/adityaayatusy/scrap-chat/pkg/platform"
	"github.com/adityaayatusy/scrap-chat/pkg/scrapchat"
	"log"
	"math/rand"
	_ "net/http/pprof"
	"os"
	"time"
)

func getRandomPort() int {
	rand.Seed(time.Now().UnixNano())
	return rand.Intn(65535-1024) + 1024 // Hindari port < 1024
}

func main() {
	var chat platform.ChatFetcher = scrapchat.New("youtube")
	if len(os.Args) < 2 {
		fmt.Println("Usage: program <arg>")
		return
	}
	arg := os.Args[1]

	data, err := chat.FetchLiveChat(arg)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	for msg := range data {
		log.Printf("(%s) %s\n", msg.UserName, msg.Message)
	}
}
