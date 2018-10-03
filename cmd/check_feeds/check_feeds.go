package main

import (
	"fmt"

	"github.com/jasonpenny/tubetoplex/internal/feedstorage"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/mmcdole/gofeed"
)

func main() {
	var db *sqlx.DB
	var err error

	db, err = sqlx.Connect("sqlite3", "__videos.db")
	if err != nil {
		panic("Could not open sqlite file")
	}

	feedstorage.SetupFeedTable(db)

	feeds, err := feedstorage.GetAllFeeds(db)
	if err != nil {
		fmt.Println(err)
		panic("Could not get list of feeds from DB")
	}
	for _, feed := range feeds {
		getFeedItemsSince(feed.URL, feed.LastItemDate)
	}
}

func getFeedItemsSince(URL, lastDate string) {
	fp := gofeed.NewParser()
	feed, _ := fp.ParseURL(URL)
	fmt.Println(feed.Title)
	for _, item := range feed.Items {
		if item.Published <= lastDate {
			break
		}
		fmt.Println(" ", item.Published, item.Link)
	}
	fmt.Println()
}
