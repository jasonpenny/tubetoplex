package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/BrianAllred/goydl"
)

func download_url(url string, season, episode int) {
	tmp := os.TempDir()
	fmt.Printf("Temp dir %s\n", tmp)

	youtubeDl := goydl.NewYoutubeDl()
	youtubeDl.Options.Output.Value = fmt.Sprintf("%sS%02dE%02d.%%(title)s-%%(id)s.%%(ext)s", tmp, season, episode)
	cmd, err := youtubeDl.Download(url)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Title: %s\n", youtubeDl.Info.Title)
	fmt.Printf("Description: %s\n", youtubeDl.Info.Description)
	fmt.Printf("AverageRating: %f\n", youtubeDl.Info.AverageRating)
	fmt.Printf("UploadDate: %s\n", youtubeDl.Info.UploadDate)

	cmd.Wait()

	filepathglob := fmt.Sprintf("%sS%02dE%02d.%s-%s*", tmp, season, episode, youtubeDl.Info.Title, youtubeDl.Info.ID)
	matches, err := filepath.Glob(filepathglob)
	filename := matches[0]

	fmt.Printf("%v", filename)
}

func main() {
	download_url("https://www.youtube.com/watch?v=C0DPdy98e4c", 3, 2)
}
