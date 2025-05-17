package scrapchat

import (
	"context"
	"github.com/xorvus/scrap-chat/internal/fetchers"
	plf "github.com/xorvus/scrap-chat/pkg/platform"
	"github.com/xorvus/scrap-chat/types"
	"log"
	"time"
)

type ScrapChat struct {
	platform string
	scrapper plf.ChatFetcher
}

func New(platform string, opts ...any) *ScrapChat {
	var scrapper plf.ChatFetcher

	ctx := context.Background()
	verbose := false

	for _, opt := range opts {
		switch v := opt.(type) {
		case bool:
			verbose = v
		case context.Context:
			ctx = v
		}
	}

	switch platform {
	case "youtube":
		scrapper = fetchers.NewYoutube(&ctx, verbose)
	default:
		log.Fatalf("Platform not support")
		return nil
	}

	return &ScrapChat{
		platform: platform,
		scrapper: scrapper,
	}
}

func (s *ScrapChat) AddCookies(path string) error {
	return s.scrapper.AddCookies(path)
}

func (s *ScrapChat) FetchLiveChat(streamID string) (<-chan *types.LiveChatMessage, error) {
	return s.scrapper.FetchLiveChat(streamID)
}

func (s *ScrapChat) FetchVideoComments(streamID string, date *time.Time) (<-chan *types.ChatMessage, error) {
	return s.scrapper.FetchVideoComments(streamID, date)
}

func (s *ScrapChat) FetchChannelInfo(path string) (*types.ChannelInfo, error) {
	return s.scrapper.FetchChannelInfo(path)
}
