package main

import (
	"log"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/jasonpenny/tubetoplex/internal/videostorage"
	"github.com/jmoiron/sqlx"
	tumblr "github.com/tumblr/tumblr.go"
	tumblr_go "github.com/tumblr/tumblrclient.go"
)

func pullNewTumblrPosts(db *sqlx.DB) {
	consumerKey := os.Getenv("TUMBLR_CONSUMER_KEY")
	consumerSecret := os.Getenv("TUMBLR_CONSUMER_SECRET")
	token := os.Getenv("TUMBLR_TOKEN")
	tokenSecret := os.Getenv("TUMBLR_TOKEN_SECRET")

	offset := 0
	for {
		params := url.Values{}
		params.Set("limit", "10")
		if offset > 0 {
			params.Set("offset", strconv.Itoa(offset))
		}

		client := tumblr_go.NewClientWithToken(consumerKey, consumerSecret, token, tokenSecret)

		resp, err := tumblr.GetPosts(client, "softwaredevvideos", params)
		if err != nil {
			panic(err)
		}

		offset += 10

		allPosts, err := resp.All()
		if err != nil {
			panic(err)
		}

		if len(allPosts) == 0 {
			log.Printf("PULL: No posts returned from Tumblr for offset %v\n", offset)
			return
		}

		stmt, err := videostorage.PrepareLookupByURL(db)
		if err != nil {
			panic(err)
		}
		pageHadNoNewVideos := true
		for _, post := range allPosts {
			video := &videostorage.Video{}

			switch pt := post.(type) {
			case *tumblr.LinkPost:
				video.URL = pt.Url
				for _, tag := range pt.Tags {
					if tag != "unprocessed" {
						video.Show = strings.ToLower(tag)
						break
					}
				}
			case *tumblr.VideoPost:
				video.URL = pt.PermalinkUrl
				for _, tag := range pt.Tags {
					if tag != "unprocessed" {
						video.Show = strings.ToLower(tag)
						break
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
			default:
				continue
			}

			videos, err := videostorage.Find(stmt, video)
			if err != nil {
				panic(err)
			}

			if len(videos) > 0 {
				// this video has already been stored, stop paging through posts
				continue
			}

			_, err = videostorage.Add(db, video, "new")
			if err != nil {
				panic(err)
			}
			pageHadNoNewVideos = false
			log.Printf("PULL: Added video %s\n", video.URL)
		}

		if pageHadNoNewVideos {
			log.Printf("PULL: No new videos")
			return
		}
	}
}
