package types

type ChannelInfo struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Image       string `json:"image"`
	Description string `json:"description"`
	URL         string `json:"url"`
}

type Author struct {
	ID         string
	Name       string
	Thumbnail  string
	URL        string
	IsUploader bool
	IsVerified bool
}

type LiveChatMessage struct {
	ID        string
	Message   string
	Author    Author
	Timestamp int64
}

type ChatMessage struct {
	ID          string
	Parent      string
	Message     string
	Author      Author
	IsPinned    bool
	IsFavorited bool
	ReplyCount  int
	LikeCount   int
	Timestamp   int64
}
