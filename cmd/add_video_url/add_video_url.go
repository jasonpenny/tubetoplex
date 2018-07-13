package main

import (
	"fmt"
	"os"

	"github.com/jasonpenny/tubetoplex/internal/showstorage"
	"github.com/jasonpenny/tubetoplex/internal/videostorage"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Printf("Usage: %s <tv show> <url>\n", os.Args[0])
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

	video := &videostorage.Video{
		Show: show.Name,
		URL:  os.Args[2],
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
		_, err = videostorage.Add(db, video, "new")
		if err != nil {
			panic(err)
		}
	}
}
