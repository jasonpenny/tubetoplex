package youtubedl

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"path/filepath"

	"github.com/BrianAllred/goydl"
)

type VideoInfo struct {
	Season      int
	Episode     int
	Title       string
	Description string
	Rating      float64
	UploadDate  string
	Filename    string
}

func DownloadURL(url string, season, episode int) *VideoInfo {
	dir := randomDir()

	youtubeDl := goydl.NewYoutubeDl()
	youtubeDl.Options.Output.Value = filepath.Join(
		dir,
		fmt.Sprintf(
			"S%02dE%02d.%%(title)s-%%(id)s.%%(ext)s",
			season, episode,
		),
	)

	cmd, err := youtubeDl.Download(url)
	if err != nil {
		log.Fatal(err)
	}

	// without this, the 2nd time it runs it stalls
	go io.Copy(ioutil.Discard, youtubeDl.Stdout)
	go io.Copy(ioutil.Discard, youtubeDl.Stderr)

	result := &VideoInfo{
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

	filepathglob := filepath.Join(
		dir,
		fmt.Sprintf(
			"S%02dE%02d.*",
			season, episode,
		),
	)

	matches, err := filepath.Glob(filepathglob)

	result.Filename = matches[0]

	return result
}
