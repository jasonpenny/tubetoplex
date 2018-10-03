package main

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/mmcdole/gofeed"
)

type feed struct {
	ID           int    `db:"id"`
	Show         string `db:"show"`
	URL          string `db:"url"`
	LastItemDate string `db:"last_update"`
}

func setupFeedTable(db *sqlx.DB) {
	db.MustExec(`
	CREATE TABLE IF NOT EXISTS feeds (
		id INTEGER PRIMARY KEY,
		show VARCHAR,
		url TEXT,
		last_update VARCHAR
	);
	`)
}

func getFeeds(db *sqlx.DB) ([]feed, error) {
	feeds := []feed{}
	err := db.Select(&feeds, "SELECT * FROM feeds ORDER BY id")
	return feeds, err
}

func main() {
	var db *sqlx.DB
	var err error

	db, err = sqlx.Connect("sqlite3", "__videos.db")
	if err != nil {
		panic("Could not open sqlite file")
	}

	setupFeedTable(db)

	feeds, err := getFeeds(db)
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
		if item.Published < lastDate {
			break
		}
		fmt.Println(" ", item.Published, item.Link)
	}
	fmt.Println()
}
