package main

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/jasonpenny/tubetoplex/internal/videostorage"
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

	videostorage.SetupTable(db)

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

		stmt, err := videostorage.PrepareLookupByURL(db)
		if err != nil {
			panic(err)
		}

		const unprocessed = "unprocessed"
		for _, post := range allPosts {
			video := &videostorage.Video{}

			shouldStore := true
			switch pt := post.(type) {
			case *tumblr.LinkPost:
				fmt.Printf("link   %d %v %v\n", pt.Id, pt.Url, pt.Tags)
				video.URL = pt.Url
				for _, tag := range pt.Tags {
					if tag != unprocessed {
						video.Show = strings.ToLower(tag)
					}
					if tag == unprocessed {
						shouldStore = false
					}
				}
				if video.URL == "" {
					fmt.Printf("%v\n", pt)
				}
			case *tumblr.VideoPost:
				fmt.Printf("video  %d %v %v\n", pt.Id, pt.PermalinkUrl, pt.Tags)
				video.URL = pt.PermalinkUrl
				for _, tag := range pt.Tags {
					if tag != unprocessed {
						video.Show = strings.ToLower(tag)
					}
					if tag == unprocessed {
						shouldStore = false
					}
				}
				if video.URL == "" {
					// fallback to parsing out of source_url
					if u, err := url.Parse(pt.SourceUrl); err == nil {
						if m, err := url.ParseQuery(u.RawQuery); err == nil {
							video.URL = m["z"][0]
						}
					}
				}

				if video.URL == "" {
					fmt.Printf("%v\n", pt)
				}
			default:
				continue
			}

			if !shouldStore || (video.URL == "") {
				fmt.Printf("  Skipping\n")
				continue
			}

			videos, err := videostorage.Find(stmt, video)
			if err != nil {
				panic(err)
			}

			if len(videos) == 0 {
				fmt.Println("  Inserting video")
				_, err = videostorage.Add(db, video, "downloaded")
				if err != nil {
					panic(err)
				}
			}
		}
	}
}
