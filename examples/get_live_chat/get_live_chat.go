package main

import (
	"fmt"
	"github.com/xorvus/scrap-chat/pkg/platform"
	"github.com/xorvus/scrap-chat/pkg/scrapchat"
	"log"
	_ "net/http/pprof"
	"os"
)

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
		log.Printf("(%s) %s\n", msg.Author.Name, msg.Message)
	}
}
