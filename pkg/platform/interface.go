package platform

import "github.com/adityaayatusy/scrap-chat/types"

type ChatFetcher interface {
	AddCookies(path string) error
	FetchLiveChat(streamID string) (<-chan *types.ChatMessage, error)
	//FetchVideoComments(videoID string) ([]types.Comment, error)
	FetchChannelInfo(path string) (types.ChannelInfo, error)
}
