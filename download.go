package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/BrianAllred/goydl"
)

type videoInfo struct {
	Season      int
	Episode     int
	Title       string
	Description string
	Rating      float64
	UploadDate  string
	Filename    string
}

func downloadURL(url string, season, episode int) *videoInfo {
	tmp := os.TempDir()

	youtubeDl := goydl.NewYoutubeDl()
	youtubeDl.Options.Output.Value = fmt.Sprintf(
		"%sS%02dE%02d.%%(title)s-%%(id)s.%%(ext)s",
		tmp, season, episode,
	)
	cmd, err := youtubeDl.Download(url)
	if err != nil {
		log.Fatal(err)
	}

	result := &videoInfo{
		Season:      season,
		Episode:     episode,
		Title:       youtubeDl.Info.Title,
		Description: youtubeDl.Info.Description,
		Rating:      youtubeDl.Info.AverageRating,
		UploadDate:  youtubeDl.Info.UploadDate,
	}

	if err = cmd.Wait(); err != nil {
		log.Fatal(err)
	}

	filepathglob := fmt.Sprintf(
		"%sS%02dE%02d.%s-%s*",
		tmp, season, episode, youtubeDl.Info.Title, youtubeDl.Info.ID,
	)
	matches, err := filepath.Glob(filepathglob)

	result.Filename = matches[0]

	return result
}
