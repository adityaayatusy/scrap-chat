package main

import (
	"fmt"
	"github.com/xorvus/scrap-chat/pkg/platform"
	"github.com/xorvus/scrap-chat/pkg/scrapchat"
)

func main() {
	var chat platform.ChatFetcher = scrapchat.New("youtube")

	channelInfo, err := chat.FetchChannelInfo("https://www.youtube.com/@LofiGirl")
	if err != nil {
		fmt.Println("Error fetching channel info:", err)
		return
	}

	fmt.Printf("ChannelID: %s, ChannelName: %s, URL: %s", channelInfo.ID, channelInfo.Name, channelInfo.URL)
}
