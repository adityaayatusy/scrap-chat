package main

import (
	"fmt"
	"github.com/adityaayatusy/scrap-chat/pkg/platform"
	"github.com/adityaayatusy/scrap-chat/pkg/scrapchat"
)

func main() {
	var chat platform.ChatFetcher = scrapchat.New("youtube")

	channelInfo, err := chat.FetchChannelInfo("https://www.youtube.com/@TheJooomers")
	if err != nil {
		fmt.Println("Error fetching channel info:", err)
		return
	}

	fmt.Printf("ChannelID: %s, ChannelName: %s, URL: %s", channelInfo.ID, channelInfo.Name, channelInfo.URL)
}
