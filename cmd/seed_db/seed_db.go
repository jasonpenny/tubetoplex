package main

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	tumblr "github.com/tumblr/tumblr.go"
	tumblr_go "github.com/tumblr/tumblrclient.go"
)

func main() {
	var db *sqlx.DB
	var err error

	db, err = sqlx.Connect("sqlite3", "__videos.db")
	if err != nil {
		panic("Could not open sqlite file")
	}

	db.MustExec(`
	CREATE TABLE IF NOT EXISTS videos (
		id INTEGER PRIMARY KEY,
		url TEXT,
		show VARCHAR(255),
		filename TEXT,
		title VARCHAR(255),
		description TEXT,
		average_rating NUMERIC,
		upload_date VARCHAR(8),
		step VARCHAR
	);
	`)

	type Video struct {
		Id            int     `db:"id"`
		Url           string  `db:"url"`
		Show          string  `db:"show"`
		Filename      string  `db:"filename"`
		Title         string  `db:"title"`
		Description   string  `db:"description"`
		AverageRating float64 `db:"average_rating"`
		UploadDate    string  `db:"upload_date"`
		Step          string  `db:"step"`
	}

	client := tumblr_go.NewClientWithToken(
		os.Getenv("TUMBLR_CONSUMER_KEY"),
		os.Getenv("TUMBLR_CONSUMER_SECRET"),
		os.Getenv("TUMBLR_TOKEN"),
		os.Getenv("TUMBLR_TOKEN_SECRET"),
	)

	offset := 0
	for {
		params := url.Values{}
		params.Set("limit", "10")
		params.Set("offset", strconv.Itoa(offset))
		offset += 10

		resp, err := tumblr.GetPosts(client, "softwaredevvideos", params)
		if err != nil {
			panic(err)
		}

		allPosts, err := resp.All()
		if err != nil {
			panic(err)
		}

		if len(allPosts) == 0 {
			break
		}

		stmt, err := db.PrepareNamed(`SELECT * FROM videos WHERE url = :url`)
		for _, post := range allPosts {
			video := &Video{}

			should_store := true
			switch pt := post.(type) {
			case *tumblr.LinkPost:
				fmt.Printf("link   %d %v %v\n", pt.Id, pt.Url, pt.Tags)
				video.Url = pt.Url
				for _, tag := range pt.Tags {
					if tag != "unprocessed" {
						video.Show = strings.ToLower(tag)
					}
					if tag == "unprocessed" {
						should_store = false
					}
				}
				if video.Url == "" {
					fmt.Printf("%v\n", pt)
				}
			case *tumblr.VideoPost:
				fmt.Printf("video  %d %v %v\n", pt.Id, pt.PermalinkUrl, pt.Tags)
				video.Url = pt.PermalinkUrl
				for _, tag := range pt.Tags {
					if tag != "unprocessed" {
						video.Show = strings.ToLower(tag)
					}
					if tag == "unprocessed" {
						should_store = false
					}
				}
				if video.Url == "" {
					// fallback to parsing out of source_url
					if u, err := url.Parse(pt.SourceUrl); err == nil {
						if m, err := url.ParseQuery(u.RawQuery); err == nil {
							video.Url = m["z"][0]
						}
					}
				}

				if video.Url == "" {
					fmt.Printf("%v\n", pt)
				}
			default:
				continue
			}

			if !should_store || (video.Url == "") {
				fmt.Printf("  Skipping\n")
				continue
			}

			videos := []Video{}
			err = stmt.Select(&videos, video)
			if err != nil {
				panic(err)
			}

			if len(videos) == 0 {
				video.Step = "downloaded"

				fmt.Println("  Inserting video")
				_, err = db.NamedExec(
					`INSERT INTO videos (url, show, filename, title, description, average_rating, upload_date, step)
					VALUES (:url, :show, :filename, :title, :description, :average_rating, :upload_date, :step)`,
					&video,
				)
				if err != nil {
					panic(err)
				}
			}
		}
	}
}
