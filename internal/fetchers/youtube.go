package fetchers

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/adityaayatusy/scrap-chat/internal/utils"
	"github.com/adityaayatusy/scrap-chat/types"
	"github.com/gocolly/colly"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	ErrStreamNotLive = errors.New("stream not live")
	jsonBufferPool   = sync.Pool{
		New: func() interface{} {
			return new(bytes.Buffer)
		},
	}
	builderPool = sync.Pool{
		New: func() interface{} {
			return new(strings.Builder)
		},
	}
)

type Youtube struct {
	cookies      []*http.Cookie
	config       *types.YTCgf
	continuation string
	videoId      string
	gsessionID   string
	sid          string
	httpClient   *http.Client
	header       http.Header
	cookieString string
}

func NewYoutube() *Youtube {
	y := &Youtube{
		httpClient: &http.Client{
			Transport: &http.Transport{
				MaxResponseHeaderBytes: 1 << 20,
			},
		},
	}
	y.header = make(http.Header)
	defaultHeaders(y.header)
	return y
}

func defaultHeaders(h http.Header) {
	headers := map[string]string{
		"accept":                      "*/*",
		"accept-language":             "en-US,en;q=0.9",
		"cache-control":               "no-cache",
		"origin":                      "https://www.youtube.com ",
		"priority":                    "u=1, i",
		"pragma":                      "no-cache",
		"referer":                     "https://www.youtube.com/ ",
		"sec-ch-ua":                   "\"Chromium\";v=\"136\", \"Brave\";v=\"136\", \"Not.A/Brand\";v=\"99\"",
		"sec-ch-ua-arch":              "\"arm\"",
		"sec-ch-ua-bitness":           "\"64\"",
		"sec-ch-ua-full-version-list": "\"Chromium\";v=\"136.0.0.0\", \"Brave\";v=\"136.0.0.0\", \"Not.A/Brand\";v=\"99.0.0.0\"",
		"sec-ch-ua-mobile":            "?0",
		"sec-ch-ua-model":             "\"\"",
		"sec-ch-ua-platform":          "\"macOS\"",
		"sec-ch-ua-platform-version":  "\"15.4.0\"",
		"sec-ch-ua-wow64":             "?0",
		"sec-fetch-dest":              "empty",
		"sec-fetch-mode":              "cors",
		"sec-fetch-site":              "same-site",
		"sec-gpc":                     "1",
		"user-agent":                  "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/136.0.0.0 Safari/537.36",
	}
	for k, v := range headers {
		h.Set(k, v)
	}
}

func (y *Youtube) AddCookies(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open cookie file: %w", err)
	}
	defer file.Close()
	var cookies []*http.Cookie
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "#") || strings.TrimSpace(line) == "" {
			continue
		}
		parts := strings.Split(line, "\t")
		if len(parts) != 7 {
			continue
		}
		timestamp, err := strconv.ParseInt(parts[4], 10, 64)
		if err != nil {
			log.Printf("Error parsing cookie time: %v", err)
			continue
		}
		cookies = append(cookies, &http.Cookie{
			Domain:  parts[0],
			Path:    parts[2],
			Name:    parts[5],
			Value:   parts[6],
			Expires: time.Unix(timestamp, 0),
			Secure:  parts[3] == "TRUE",
		})
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading cookie file: %w", err)
	}
	y.cookies = cookies
	y.cookieString = createCookieString(cookies)
	return nil
}

func createCookieString(cookies []*http.Cookie) string {
	parts := make([]string, 0, len(cookies))
	for _, c := range cookies {
		parts = append(parts, fmt.Sprintf("%s=%s", c.Name, c.Value))
	}
	return strings.Join(parts, "; ")
}

// FetchLiveChat returns a channel streaming live chat messages for a given YouTube URL or channel.
func (y *Youtube) FetchLiveChat(path string) (<-chan *types.ChatMessage, error) {
	url := path
	if strings.Contains(url, "@") {
		info, err := y.FetchChannelInfo(path)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch channel info: %w", err)
		}
		url = info.URL + "/live"
	}

	if err := y.getConfig(url); err != nil {
		return nil, fmt.Errorf("failed to get config: %w", err)
	}
	y.chooseServer()
	y.getSID()

	msg := make(chan *types.ChatMessage)

	go func() {
		defer close(msg)
		y.streamChat(func(params []types.YTChatMessage) {
			for _, param := range params {
				userImage := ""
				if len(param.Author.AuthorImages) > 0 {
					userImage = param.Author.AuthorImages[0].URL
				}
				msg <- &types.ChatMessage{
					ID:        param.ID,
					Message:   param.Message,
					UserId:    param.Author.AuthorID,
					UserName:  param.Author.AuthorName,
					UserImage: userImage,
					Timestamp: param.Timestamp.Unix(),
				}
			}
		})
	}()

	return msg, nil
}

// FetchChannelInfo scrapes channel info from a YouTube channel page.
func (y *Youtube) FetchChannelInfo(path string) (types.ChannelInfo, error) {
	if !strings.HasPrefix(path, "http") && strings.Contains(path, "@") {
		path = "https://www.youtube.com/" + path
	}

	info := &types.ChannelInfo{}
	c := colly.NewCollector(
		colly.MaxDepth(1),
		colly.Async(true),
		colly.UserAgent("Mozilla/5.0"),
	)

	c.OnHTML(`meta[property="og:title"]`, func(e *colly.HTMLElement) {
		info.Name = e.Attr("content")
	})
	c.OnHTML(`meta[property="og:image"]`, func(e *colly.HTMLElement) {
		info.Image = e.Attr("content")
	})
	c.OnHTML(`meta[property="og:description"]`, func(e *colly.HTMLElement) {
		info.Description = e.Attr("content")
	})
	c.OnHTML(`meta[property="og:url"]`, func(e *colly.HTMLElement) {
		url := e.Attr("content")
		info.URL = url
		if strings.Contains(url, "/channel/") {
			info.ID = strings.SplitN(url, "/channel/", 2)[1]
		}
	})

	c.SetRequestTimeout(5 * time.Second)

	err := c.Visit(path)
	if err != nil {
		return types.ChannelInfo{}, err
	}

	c.Wait()

	return *info, nil
}

// getConfig scrapes the YouTube page for config and initial data.
func (y *Youtube) getConfig(url string) error {
	c := colly.NewCollector()

	var ytConfig types.YTCgf
	var ytInitialData types.YTInitialData
	ytCfgRegex := regexp.MustCompile(`ytcfg\.set\((\{.*?\})\);`)
	initialDataRegex := regexp.MustCompile(`(?s)(?:window\s*\[\s*["']ytInitialData["']\s*\]|ytInitialData)\s*=\s*({.+?})\s*;`)

	c.OnHTML("script", func(e *colly.HTMLElement) {
		scriptContent := e.Text
		if match := ytCfgRegex.FindStringSubmatch(scriptContent); len(match) > 1 {
			_ = json.Unmarshal([]byte(match[1]), &ytConfig)
		}
		if match := initialDataRegex.FindStringSubmatch(scriptContent); len(match) > 1 {
			_ = json.Unmarshal([]byte(match[1]), &ytInitialData)
		}
	})

	c.SetRequestTimeout(5 * time.Second)
	if err := c.Visit(url); err != nil {
		return err
	}

	subMenuItems := ytInitialData.Contents.TwoColumnWatchNextResults.ConversationBar.LiveChatRenderer.Header.LiveChatHeaderRenderer.ViewSelector.SortFilterSubMenuRenderer.SubMenuItems
	if len(subMenuItems) == 0 {
		return ErrStreamNotLive
	}
	initialContinuationInfo := subMenuItems[1].Continuation.ReloadContinuationData.Continuation

	y.continuation = initialContinuationInfo
	y.config = &ytConfig
	y.videoId = ytInitialData.CurrentVideoEndpoint.WatchEndpoint.VideoId

	return nil
}

const (
	REG_FIRST_CHAT = `\[\[\d*,\[\[null,null,\["([^"]+)"\]\]\]\]`
	REG_NO_CHAT    = `\[\[\d*,\[\[\[\[.*\[null,null,\["\d*`
	REG_CHAT       = `\d{16,}`
)

var (
	regFirstChat = regexp.MustCompile(REG_FIRST_CHAT)
	regNoChat    = regexp.MustCompile(REG_NO_CHAT)
)

// streamChat handles the polling and parsing of live chat messages.
func (y *Youtube) streamChat(param func([]types.YTChatMessage)) {
	lastTime := time.Now().Unix()
	y.longPooling(func(res string) {
		tempTime := time.Now()
		diff := tempTime.Sub(time.Unix(lastTime, 0))

		switch {
		case IsRegexTrue(regFirstChat, res):
			res, _ := y.sendMessage(&MessageOptions{
				Timestamp: "",
				IsTimeout: false,
				IsFirst:   true,
			})
			param(res)
		case diff >= 10*time.Second:
			res, _ := y.sendMessage(&MessageOptions{
				Timestamp: "",
				IsTimeout: true,
				IsFirst:   false,
			})
			param(res)
		case IsRegexTrue(regNoChat, res):
			// No chat, do nothing
		case func() bool {
			ok, match := RegexGetValue(REG_CHAT, res)
			if ok {
				res, _ := y.sendMessage(&MessageOptions{
					Timestamp: match[0],
					IsTimeout: false,
					IsFirst:   false,
				})
				param(res)
			}
			return ok
		}():
			// handled in the inline func
		default:
			log.Printf("[%dms] Undefined: %s", diff/time.Millisecond, res)
		}

		lastTime = tempTime.Unix()
	})
}

// IsRegexTrue returns true if the regex matches the data.
func IsRegexTrue(r *regexp.Regexp, str string) bool {
	return r.MatchString(str)
}

// RegexGetValue returns true and the matches if the regex finds more than one match.
func RegexGetValue(regex, data string) (bool, []string) {
	re := regexp.MustCompile(regex)
	match := re.FindAllString(data, -1)
	if len(match) > 0 {
		return true, match
	}
	return false, nil
}

func (y *Youtube) copyHeaders(req *http.Request, h http.Header) {
	for k, vv := range h {
		for _, v := range vv {
			req.Header.Add(k, v)
		}
	}
	req.Header.Set("Cookie", y.cookieString)
}

// longPooling performs long polling to receive live chat updates.
func (y *Youtube) longPooling(param func(string)) {
	log.Println("Long pool...")
	for {
		url := fmt.Sprintf("https://signaler-pa.youtube.com/punctual/multi-watch/channel?VER=8&gsessionid=%s&key=%s&RID=rpc&SID=%s&AID=0&CI=0&TYPE=xmlhttp&zx=%s&t=1",
			y.gsessionID, y.config.LIVE_CHAT_BASE_TANGO_CONFIG.API_KEY, y.sid, utils.GenerateZX())

		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			log.Printf("Request error: %v", err)
			return
		}

		y.copyHeaders(req, y.header)

		resp, err := y.httpClient.Do(req)
		if err != nil {
			log.Printf("HTTP error: %v", err)
			return
		}
		defer resp.Body.Close()

		log.Println("Connected, streaming...")
		reader := bufio.NewReader(resp.Body)
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				if err == io.EOF {
					log.Println("Stream closed by server.")
				} else {
					log.Printf("Error reading stream: %v", err)
				}
				break
			}
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			param(line)
		}
		log.Println("Reconnecting...")
		time.Sleep(1 * time.Second)
	}
}

func (y *Youtube) getSID() {
	url := fmt.Sprintf("https://signaler-pa.youtube.com/punctual/multi-watch/channel?VER=8&gsessionid=%s&key=%s&RID=6167&CVER=22&zx=%s&t=1",
		y.gsessionID, y.config.LIVE_CHAT_BASE_TANGO_CONFIG.API_KEY, utils.GenerateZX())
	payloadRaw := fmt.Sprintf("count=1&ofs=0&req0___data__=[[[\"1\",[null,null,null,[9,5],null,[[\"youtube_live_chat_web\"],[1],[[[\"chat~%s\"]]]],null,null,1],null,3]]]", y.videoId)
	payload := strings.NewReader(payloadRaw)

	req, err := http.NewRequest("POST", url, payload)
	if err != nil {
		log.Printf("getSID request error: %v", err)
		return
	}

	y.copyHeaders(req, y.header)
	req.Header.Set("content-type", "application/x-www-form-urlencoded")
	req.Header.Set("x-webchannel-content-type", "application/json+protobuf")

	resp, err := y.httpClient.Do(req)
	if err != nil {
		log.Printf("getSID HTTP error: %v", err)
		return
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("getSID read error: %v", err)
		return
	}
	s := string(body)

	idx := strings.Index(s, "[[")
	if idx == -1 {
		log.Println("getSID: JSON array not found")
		return
	}

	jsonPart := s[idx:]
	var parsed [][]interface{}
	if err := json.Unmarshal([]byte(jsonPart), &parsed); err != nil {
		log.Printf("getSID: error parsing JSON: %v", err)
		return
	}

	innerArray, ok := parsed[0][1].([]interface{})
	if !ok || len(innerArray) < 2 {
		log.Println("getSID: unexpected format")
		return
	}

	sid, ok := innerArray[1].(string)
	if !ok {
		log.Println("getSID: SID is not a string")
		return
	}

	y.sid = sid
}

// chooseServer selects the server for the live chat session.
func (y *Youtube) chooseServer() {
	url := fmt.Sprintf("https://signaler-pa.youtube.com/punctual/v1/chooseServer?key=%s", y.config.LIVE_CHAT_BASE_TANGO_CONFIG.API_KEY) //rawPayload := fmt.Sprintf("[[null,null,null,[9,5],null,[[\"youtube_live_chat_web\"],[1],[[[\"chat~%s\"]]]]],null,null,0]", y.videoId)
	rawPayload := fmt.Sprintf("[[null,null,null,[9,5],null,[[\"youtube_live_chat_web\"],[1],[[[\"chat~%s\"]]]]],null,null,0]", y.videoId)
	payload := strings.NewReader(rawPayload)

	req, err := http.NewRequest("POST", url, payload)
	if err != nil {
		log.Printf("chooseServer request error: %v", err)
		return
	}

	y.copyHeaders(req, y.header)
	req.Header.Set("content-type", "application/json+protobuf")

	resp, err := y.httpClient.Do(req)
	if err != nil {
		log.Printf("chooseServer HTTP error: %v", err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("chooseServer read error: %v", err)
		return
	}

	var result []interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		log.Printf("chooseServer JSON unmarshal error: %v", err)
		return
	}

	if len(result) > 0 {
		gsessionID, ok := result[0].(string)
		if ok {
			y.gsessionID = gsessionID
		} else {
			log.Println("chooseServer: gsessionid is not a string")
		}
	}
}

type MessageOptions struct {
	Timestamp string
	IsTimeout bool
	IsFirst   bool
}

// sendMessage sends a request to get live chat messages.
func (y *Youtube) sendMessage(opts *MessageOptions) ([]types.YTChatMessage, error) {
	url := "https://www.youtube.com/youtubei/v1/live_chat/get_live_chat?prettyPrint=false"

	// 1. Reuse JSON buffer
	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)

	context := types.YTPayloadMessageLive{
		Context:      y.config.INNERTUBE_CONTEXT,
		Continuation: y.continuation,
		WebClientInfo: types.YTWebClientInfo{
			IsDocumentHidden: false,
		},
	}

	if opts.IsTimeout {
		context.IsInvalidationTimeoutRequest = &opts.IsTimeout
	}
	if !opts.IsFirst && !opts.IsTimeout {
		context.InvalidationPayloadLastPublishAtUsec = &opts.Timestamp
	}

	if err := encoder.Encode(context); err != nil {
		return nil, fmt.Errorf("sendMessage: marshal error: %w", err)
	}

	// 2. Stream request body
	req, err := http.NewRequest("POST", url, &buf)
	if err != nil {
		return nil, fmt.Errorf("sendMessage: request error: %w", err)
	}

	res, err := y.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("sendMessage: HTTP error: %w", err)
	}
	defer res.Body.Close()

	// 3. Stream JSON parsing
	var chatMsgResp types.YTChatMessagesResponse
	decoder := json.NewDecoder(res.Body)
	if err := decoder.Decode(&chatMsgResp); err != nil {
		return nil, fmt.Errorf("sendMessage: unmarshal error: %w", err)
	}

	// 4. Pre-allocate messages slice
	actions := chatMsgResp.ContinuationContents.LiveChatContinuation.Actions
	chatMessages := make([]types.YTChatMessage, 0, len(actions))

	// 5. Reuse builder and buffers
	var textBuilder strings.Builder
	const avgMessageSize = 128               // Estimated average message size
	thumbnailsBuffer := make([]string, 0, 2) // Reusable buffer for thumbnails

	for _, action := range actions {
		renderer := action.AddChatItemAction.Item.LiveChatTextMessageRenderer
		if len(renderer.Message.Runs) == 0 {
			continue
		}

		// Reset reusable buffers
		textBuilder.Reset()
		thumbnailsBuffer = thumbnailsBuffer[:0]

		// Pre-allocate message text buffer
		textBuilder.Grow(avgMessageSize)

		for _, run := range renderer.Message.Runs {
			switch {
			case run.Text != "":
				textBuilder.WriteString(run.Text)
			case run.Emoji.IsCustomEmoji:
				// Process thumbnails
				if images := run.Emoji.Image.Thumbnails; len(images) > 0 {
					thumbnailsBuffer = append(thumbnailsBuffer, images[len(images)-1].Url)
				}
				for _, url := range thumbnailsBuffer {
					textBuilder.WriteString(" ")
					textBuilder.WriteString(url)
					textBuilder.WriteString(" ")
				}
			default:
				textBuilder.WriteString(run.Emoji.EmojiId)
			}
		}

		// 6. Avoid nested struct copies
		chatMessages = append(chatMessages, types.YTChatMessage{
			ID: renderer.ID,
			Author: types.YTAuthor{
				AuthorName:   renderer.AuthorName.SimpleText,
				AuthorID:     renderer.AuthorExternalChannelID,
				AuthorImages: renderer.AuthorPhoto.Thumbnails,
			},
			Timestamp: parseMicroSeconds(renderer.TimestampUsec),
			Message:   textBuilder.String(),
		})
	}

	// 7. Update continuation (if exists)
	if conts := chatMsgResp.ContinuationContents.LiveChatContinuation.Continuations; len(conts) > 0 {
		y.continuation = conts[0].InvalidationContinuationData.Continuation
	}

	return chatMessages, nil
}

// parseMicroSeconds parses a microsecond timestamp string to time.Time.
func parseMicroSeconds(timeStampStr string) time.Time {
	tm, err := strconv.ParseInt(timeStampStr, 10, 64)
	if err != nil {
		return time.Time{}
	}
	tm = tm / 1000
	sec := tm / 1000
	msec := tm % 1000
	return time.Unix(sec, msec*int64(time.Millisecond))
}
