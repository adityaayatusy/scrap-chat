package platform

import (
	"github.com/adityaayatusy/scrap-chat/types"
	"time"
)

type ChatFetcher interface {
	AddCookies(path string) error
	FetchLiveChat(streamID string) (<-chan *types.LiveChatMessage, error)
	FetchVideoComments(videoID string, date *time.Time) (<-chan *types.ChatMessage, error)
	FetchChannelInfo(path string) (*types.ChannelInfo, error)
}
