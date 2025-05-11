package types

type ChannelInfo struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Image       string `json:"image"`
	Description string `json:"description"`
	URL         string `json:"url"`
}

type ChatMessage struct {
	ID        string `json:"id"`
	Message   string `json:"message"`
	UserId    string `json:"userId"`
	UserName  string `json:"username"`
	UserImage string `json:"image"`
	Timestamp int64  `json:"timestamp"`
}
