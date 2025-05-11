package types

import "time"

type YTInnerTubeContext struct {
	Client struct {
		Hl               string `json:"hl"`
		Gl               string `json:"gl"`
		RemoteHost       string `json:"remoteHost"`
		DeviceMake       string `json:"deviceMake"`
		DeviceModel      string `json:"deviceModel"`
		VisitorData      string `json:"visitorData"`
		UserAgent        string `json:"userAgent"`
		ClientName       string `json:"clientName"`
		ClientVersion    string `json:"clientVersion"`
		OsName           string `json:"osName"`
		OsVersion        string `json:"osVersion"`
		OriginalUrl      string `json:"originalUrl"`
		Platform         string `json:"platform"`
		ClientFormFactor string `json:"clientFormFactor"`
		ConfigInfo       struct {
			AppInstallData string `json:"appInstallData"`
		} `json:"configInfo"`
	} `json:"client"`
}

type LiveChatBaseTangoConfig struct {
	API_KEY string `json:"apiKey"`
}

type YTPayloadMessageLive struct {
	Context                              YTInnerTubeContext `json:"context"`
	Continuation                         string             `json:"continuation"`
	WebClientInfo                        YTWebClientInfo    `json:"webClientInfo"`
	InvalidationPayloadLastPublishAtUsec *string            `json:"invalidationPayloadLastPublishAtUsec,omitempty"`
	IsInvalidationTimeoutRequest         *bool              `json:"isInvalidationTimeoutRequest,omitempty"`
}

type YTWebClientInfo struct {
	IsDocumentHidden bool `json:"IsDocumentHidden"`
}

type YTCgf struct {
	INNERTUBE_API_KEY             string
	LIVE_CHAT_BASE_TANGO_CONFIG   LiveChatBaseTangoConfig
	INNERTUBE_CONTEXT             YTInnerTubeContext
	INNERTUBE_CONTEXT_CLIENT_NAME string
	INNERTUBE_CLIENT_VERSION      string
	ID_TOKEN                      string
}

type YTSubMenuItems struct {
	Title        string
	Continuation struct {
		ReloadContinuationData struct {
			Continuation string
		}
	}
}

type YTInitialData struct {
	Contents struct {
		TwoColumnWatchNextResults struct {
			ConversationBar struct {
				LiveChatRenderer struct {
					Header struct {
						LiveChatHeaderRenderer struct {
							ViewSelector struct {
								SortFilterSubMenuRenderer struct {
									SubMenuItems []YTSubMenuItems `json:"subMenuItems"`
								}
							}
						}
					}
				}
			}
		}
	}
	CurrentVideoEndpoint struct {
		WatchEndpoint struct {
			VideoId string `json:"videoId"`
		}
	}
}

type YTChatMessagesResponse struct {
	ContinuationContents struct {
		LiveChatContinuation struct {
			Actions       []YTActions          `json:"actions"`
			Continuations []YTContinuationChat `json:"continuations"`
		} `json:"liveChatContinuation"`
	} `json:"continuationContents"`
}

type YTActions struct {
	AddChatItemAction struct {
		Item struct {
			LiveChatTextMessageRenderer struct {
				ID      string `json:"id"`
				Message struct {
					Runs []YTRuns `json:"runs"`
				} `json:"message"`
				AuthorName struct {
					SimpleText string `json:"simpleText"`
				}
				AuthorPhoto struct {
					Thumbnails []YTThumbnails `json:"thumbnails"`
				} `json:"authorPhoto"`
				AuthorExternalChannelID string `json:"authorExternalChannelId"`
				TimestampUsec           string `json:"timestampUsec"`
			} `json:"liveChatTextMessageRenderer"`
		} `json:"item"`
	} `json:"addChatItemAction"`
}

type YTThumbnails struct {
	URL    string `json:"url"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

type YTRuns struct {
	Text  string `json:"text,omitempty"`
	Emoji struct {
		EmojiId       string `json:"emojiId"`
		IsCustomEmoji bool   `json:"isCustomEmoji,omitempty"`
		Image         struct {
			Thumbnails []struct {
				Url string `json:"url,omitempty"`
			}
		}
	} `json:"emoji,omitempty"`
}

type YTContinuationChat struct {
	TimedContinuationData struct {
		Continuation string `json:"continuation"`
		TimeoutMs    int    `json:"timeoutMs"`
	} `json:"timedContinuationData"`
	InvalidationContinuationData struct {
		Continuation string `json:"continuation"`
		TimeoutMs    int    `json:"timeoutMs"`
	} `json:"invalidationContinuationData"`
}

type YTChatMessage struct {
	ID        string
	Message   string
	Author    YTAuthor
	Timestamp time.Time
}

type YTAuthor struct {
	AuthorID     string
	AuthorName   string
	AuthorImages []YTThumbnails
}
