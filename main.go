package main

import (
	"fmt"
	"log"
	"os"

	"github.com/BrianAllred/goydl"
)

func main() {
	tmp := os.TempDir()
	fmt.Printf("Temp dir %s\n", tmp)

	youtubeDl := goydl.NewYoutubeDl()
	youtubeDl.Options.Output.Value = tmp + "%(title)s-%(id)s.%(ext)s"
	cmd, err := youtubeDl.Download("https://www.youtube.com/watch?v=C0DPdy98e4c")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Title: %s\n", youtubeDl.Info.Title)
	fmt.Printf("Description: %s\n", youtubeDl.Info.Description)
	fmt.Printf("AverageRating: %f\n", youtubeDl.Info.AverageRating)
	fmt.Printf("UploadDate: %s\n", youtubeDl.Info.UploadDate)
	// can't trust this, doesn't get set by the lib and wrong extension with the fix
	// fmt.Printf("Filename: %v\n", youtubeDl.Info.Filename)

	cmd.Wait()
}
