package main

import (
	"crypto/rand"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/jasonpenny/tubetoplex/internal/filecopier"
	"github.com/jasonpenny/tubetoplex/internal/showstorage"
	"github.com/jasonpenny/tubetoplex/internal/videostorage"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Printf("Usage: %s <tv show> <filename>\n", os.Args[0])
		os.Exit(1)
	}
	var db *sqlx.DB
	var err error

	db, err = sqlx.Connect("sqlite3", "__videos.db")
	if err != nil {
		panic("Could not open sqlite file")
	}

	videostorage.SetupTable(db)
	showstorage.SetupTable(db)

	show, err := showstorage.Find(db, os.Args[1])
	if err != nil {
		fmt.Printf("Show not found %v\n", err)
		os.Exit(2)
	}

	inputFilename := filepath.Base(os.Args[2])
	video := &videostorage.Video{
		Show:  show.Name,
		URL:   inputFilename,
		Title: inputFilename[:strings.LastIndex(inputFilename, ".")],
	}

	stmt, err := videostorage.PrepareLookupByURL(db)
	if err != nil {
		panic(err)
	}
	videos, err := videostorage.Find(stmt, video)
	if err != nil {
		panic(err)
	}
	if len(videos) == 0 {
		_, err := videostorage.Add(db, video, "new")
		if err != nil {
			panic(err)
		}

		// reload video to get the id
		videos, err = videostorage.Find(stmt, video)
		if err != nil {
			panic(err)
		}
		video = &videos[0]
	}

	// number the video
	show.NextEpisode++
	if _, err := showstorage.Update(db, show); err != nil {
		log.Printf("Show could not be updated %v\n", err)
		os.Exit(3)
	}

	video.SeasonNum = show.LatestSeason
	video.EpisodeNum = show.NextEpisode
	//videostorage.Update(db, video, "numbered")

	// create the dir to hold the video
	dir := randomDir()

	video.Filename = filepath.Join(
		dir,
		fmt.Sprintf(
			"S%02dE%02d.%s",
			video.SeasonNum, video.EpisodeNum,
			filepath.Base(os.Args[2]),
		),
	)

	// copy the file to the dir with the season and episode number
	if err := filecopier.CopyFile(os.Args[2], video.Filename); err != nil {
		log.Printf("Failed to copy video file %v\n", err)
		os.Exit(5)
	}

	// mark as downloaded, further steps will be handled by tubetoplex
	if _, err := videostorage.Update(db, video, "downloaded"); err != nil {
		log.Printf("Failed to update [video] %v\n", err)
		os.Exit(6)
	}
}

func randomDir() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		panic("Could not read 16 random bytes")
	}

	path := filepath.Join(".", "download", fmt.Sprintf("%X", b))
	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		panic("Could not create download directory")
	}

	return path
}
