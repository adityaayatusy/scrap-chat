package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/xorvus/scrap-chat/pkg/platform"
	"github.com/xorvus/scrap-chat/pkg/scrapchat"
	"github.com/xorvus/scrap-chat/types"
	"io"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
)

var version = "dev"

func main() {
	var showVersion bool
	flag.BoolVar(&showVersion, "version", false, "Display program version")
	flag.BoolVar(&showVersion, "v", false, "Display program version (short form)")

	var msgType string
	flag.StringVar(&msgType, "type", "", "Type of scrap [live, video, info]")
	flag.StringVar(&msgType, "t", "", "Type of scrap [live, video, info] (short form)")

	var output string
	flag.StringVar(&output, "output", "log", "Output result destination [log, file]")
	flag.StringVar(&output, "o", "log", "Output result destination [log, file] (short form)")

	var format string
	flag.StringVar(&format, "format", "default", "Format of result [default, json, custom]")
	flag.StringVar(&format, "f", "default", "Format of result [default, json, custom] (short form)")

	var customOutput string
	flag.StringVar(&customOutput, "custom-output", "", "Custom output template (e.g., \"TITLE: TITLE, ID: ID\")")
	flag.StringVar(&customOutput, "co", "", "Custom output template (short form)")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] <url>\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Options:\n")
		fmt.Fprintf(os.Stderr, "  -v, --version           Display program version\n")
		fmt.Fprintf(os.Stderr, "  -t, --type              Type of scrap [live, video, info]\n")
		fmt.Fprintf(os.Stderr, "  -o, --output            Output destination [log, file]\n")
		fmt.Fprintf(os.Stderr, "  -f, --format            Format of result [default, json, custom]\n")
		fmt.Fprintf(os.Stderr, "  -co, --custom-output     Custom output template (for format=custom)\n")
	}

	flag.Parse()

	if showVersion {
		fmt.Println("Version:", version)
		os.Exit(0)
	}

	if flag.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "Error: Missing URL")
		flag.Usage()
		os.Exit(1)
	}
	url := flag.Arg(0)

	var chat platform.ChatFetcher = scrapchat.New("youtube")

	switch strings.ToLower(msgType) {
	case "live":
		liveChat, err := chat.FetchLiveChat(url)
		if err != nil {
			log.Fatalf("Error fetching live chat: %v", err)
		}

		handleLiveOutput(liveChat, output, format, customOutput)
	case "video":
		//chat.FetchVideoComments(url, nil)
	case "info":
		result, err := chat.FetchChannelInfo(url)
		if err != nil {
			log.Fatalf("Error fetching info: %v", err)
		}
		handleInfoOutput(result, output, format, customOutput)
	default:
		fmt.Fprintln(os.Stderr, "Error: Unknown type. Use -h for help.")
		os.Exit(1)
	}
}

func handleLiveOutput(chats <-chan *types.LiveChatMessage, output, format, customOutput string) {
	var writer *os.File
	var err error
	isFirst := true
	needCloseArray := false

	if output == "file" && format == "json" {
		writer, err = os.OpenFile("live_output.json", os.O_CREATE|os.O_RDWR, 0644)
		if err != nil {
			log.Fatalf("Failed to open file: %v", err)
		}

		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		go func() {
			<-c
			if needCloseArray {
				_, _ = writer.WriteString("\n]\n")
				writer.Sync()
			}
			writer.Close()
			fmt.Println("\nProgram interrupted. Closed JSON array in file.")
			os.Exit(0)
		}()

		info, err := writer.Stat()
		if err != nil {
			log.Fatalf("Failed to stat file: %v", err)
		}

		if info.Size() == 0 {
			_, err = writer.WriteString("[\n")
			if err != nil {
				log.Fatalf("Failed to write array start: %v", err)
			}
			isFirst = true
		} else {
			data, err := os.ReadFile("live_output.json")
			if err != nil {
				log.Fatalf("Failed to read existing file: %v", err)
			}

			trimmed := bytes.TrimRight(data, "\n\r ")
			if len(trimmed) < 2 || string(trimmed[len(trimmed)-2:]) != "]}" {
				if string(trimmed[len(trimmed)-2:]) == "]\n" || string(trimmed[len(trimmed)-1:]) == "]" {
					trimmed = bytes.TrimRight(trimmed, "]\n\r ")
				}
			}

			err = os.WriteFile("live_output.json", trimmed, 0644)
			if err != nil {
				log.Fatalf("Failed to truncate file for append: %v", err)
			}

			writer.Seek(0, io.SeekEnd)
			isFirst = false
		}

		needCloseArray = true
		defer func() {
			if needCloseArray {
				_, _ = writer.WriteString("\n]\n")
				writer.Close()
			}
		}()
	} else {
		writer = os.Stdout
	}

	for chat := range chats {
		var line string
		switch format {
		case "json":
			jsonOutput, err := json.MarshalIndent(chat, "  ", "  ")
			if err != nil {
				log.Fatalf("Failed to marshal JSON: %v", err)
			}
			line = string(jsonOutput)

		case "custom":
			if strings.TrimSpace(customOutput) == "" {
				log.Fatal("Custom format selected but no custom-output template provided")
			}
			line = applyLiveCustomTemplate(customOutput, chat)

		default:
			line = fmt.Sprintf("%+v", chat)
		}

		if format == "json" && output == "file" {
			if !isFirst {
				_, _ = writer.WriteString(",\n")
			}
			_, err := writer.WriteString(line)
			if err != nil {
				log.Fatalf("Failed to write to file: %v", err)
			}
			fmt.Printf("%s :[%s] %s\n", time.Unix(chat.Timestamp, 0).Format("2006/01/02 15:04:05"), chat.Author.Name, chat.Message)
		} else {
			if !isFirst {
				fmt.Fprintln(writer)
			}
			fmt.Fprint(writer, line)
		}

		isFirst = false
	}
}

func handleInfoOutput(result *types.ChannelInfo, output, format, customOutput string) {
	var formatted string

	switch format {
	case "json":
		jsonOutput, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			log.Fatalf("Failed to marshal JSON: %v", err)
		}
		formatted = string(jsonOutput)
	case "custom":
		if customOutput == "" {
			log.Fatal("Custom format selected but no custom-output template provided")
		}
		formatted = applyCustomTemplate(customOutput, result)
	default:
		formatted = fmt.Sprintf("%+v", result)
	}

	if output == "file" {
		ext := "txt"
		if format == "json" {
			ext = "json"
		}
		err := os.WriteFile("info_output."+ext, []byte(formatted), 0644)
		if err != nil {
			log.Fatalf("Failed to write file: %v", err)
		}
		fmt.Println("Result written to info_output.txt")
	} else {
		fmt.Println(formatted)
	}
}

func applyCustomTemplate(template string, info *types.ChannelInfo) string {
	replacer := strings.NewReplacer(
		"ID", info.ID,
		"NAME", info.Name,
		"DESC", info.Description,
		"IMAGE", info.Image,
		"URL", info.URL,
	)
	return replacer.Replace(template)
}

func applyLiveCustomTemplate(template string, info *types.LiveChatMessage) string {
	replacer := strings.NewReplacer(
		"ID", info.ID,
		"MESSAGE", info.Message,
		"AUTHOR_ID", info.Author.ID,
		"AUTHOR_NAME", info.Author.Name,
		"AUTHOR_URL", info.Author.URL,
		"AUTHOR_THUMBNAIL", info.Author.Thumbnail,
		"TIME", strconv.FormatInt(info.Timestamp, 10),
	)
	return replacer.Replace(template)
}
