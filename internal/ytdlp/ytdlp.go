package ytdlp

import (
	"fmt"
	"github.com/adityaayatusy/scrap-chat/internal/utils"
	"log"
	"os"
	"os/exec"
	"runtime"
)

const version = "2025.04.30"

type YtDlp struct {
}

func NewYtDlp() *YtDlp {
	return &YtDlp{}
}

func (d *YtDlp) Check() error {
	path := "yt-dlp"

	if !utils.FileExists(path) {
		os := runtime.GOOS
		url := ""
		switch os {
		case "windows":
			url = fmt.Sprintf("https://github.com/yt-dlp/yt-dlp/releases/download/%s/yt-dlp.exe", version)
		case "linux", "darwin":
			url = fmt.Sprintf("https://github.com/yt-dlp/yt-dlp/releases/download/%s/yt-dlp", version)
		default:
			return fmt.Errorf("Running on unknown OS: %s\n", os)
		}

		if url != "" {
			utils.DownloadFile(url, path)
		}
	}

	return nil
}

func (d *YtDlp) DownloadComments(url string) {
	err := d.Check()
	if err != nil {
		log.Fatalln(err)
	}

	outputFile := "comments.json"

	// Create or open the output file
	file, err := os.Create(outputFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating file: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	// Set up the yt-dlp command
	cmd := exec.Command("yt-dlp", "--get-comments", "--no-download", "--print", "%(comments)j", url)

	// Redirect command output to the file
	cmd.Stdout = file
	// Capture stderr for error handling
	cmd.Stderr = os.Stderr

	// Run the command
	err = cmd.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error running yt-dlp: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Comments saved to %s\n", outputFile)
}
