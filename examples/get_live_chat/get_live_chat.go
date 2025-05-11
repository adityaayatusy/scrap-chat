package main

import (
	"fmt"
	"github.com/adityaayatusy/scrap-chat/pkg/platform"
	"github.com/adityaayatusy/scrap-chat/pkg/scrapchat"
)

func main() {
	var chat platform.ChatFetcher = scrapchat.New("youtube")
	data, err := chat.FetchLiveChat("https://www.youtube.com/@taraartsgameindonesia")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	for msg := range data {
		fmt.Printf("(%s) %s\n", msg.UserName, msg.Message)
	}
}
