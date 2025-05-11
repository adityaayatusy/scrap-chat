package types

type ChannelAbout struct {
	Description         string `json:"description"`
	SubscriberCountText string `json:"subscriberCountText"`
	ViewCountText       string `json:"viewCountText"`
	JoinedDateText      string `json:"content"`
	CanonicalChannelURL string `json:"canonicalChannelUrl"`
	VideoCountText      string `json:"videoCountText"`
	Country             string `json:"country"`
	BusinessEmailButton struct {
		TitleFormatted struct {
			Content string `json:"content"`
		} `json:"titleFormatted"`
	} `json:"businessEmailRevealButton"`
	Links []struct {
		Title struct {
			Content string `json:"content"`
		} `json:"title"`
		Link struct {
			Content     string `json:"content"`
			CommandRuns []struct {
				OnTap struct {
					URL string `json:"url"`
				} `json:"onTap"`
			} `json:"commandRuns"`
		} `json:"link"`
	} `json:"links"`
}

type AboutRenderer struct {
	AboutChannelRenderer struct {
		Metadata struct {
			ViewModel ChannelAbout `json:"aboutChannelViewModel"`
		} `json:"metadata"`
	} `json:"aboutChannelRenderer"`
}

type ContinuationItem struct {
	AboutRenderer AboutRenderer `json:"aboutChannelRenderer"`
}

type AppendContinuationItemsAction struct {
	ContinuationItems []ContinuationItem `json:"continuationItems"`
}

type ResponseAbout struct {
	OnResponseReceivedEndpoints []struct {
		AppendContinuationItemsAction AppendContinuationItemsAction `json:"appendContinuationItemsAction"`
	} `json:"onResponseReceivedEndpoints"`
}

type Context struct {
	Client       ClientInfo `json:"client"`
	Continuation string     `json:"continuation"`
}

type ClientInfo struct {
	ClientName    string `json:"clientName"`
	ClientVersion string `json:"clientVersion"`
	Platform      string `json:"platform,omitempty"`
	DeviceModel   string `json:"deviceModel,omitempty"`
	UserAgent     string `json:"userAgent,omitempty"`
}
