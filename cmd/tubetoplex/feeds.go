package main

import (
	"log"

	"github.com/jasonpenny/tubetoplex/internal/feedstorage"
	"github.com/jasonpenny/tubetoplex/internal/videostorage"
	"github.com/jmoiron/sqlx"
	"github.com/mmcdole/gofeed"
)

func pullNewFeeds(db *sqlx.DB) {
	feedstorage.SetupFeedTable(db)

	feeds, err := feedstorage.GetAllFeeds(db)
	if err != nil {
		log.Printf("feedstorage.GetAllFeeds() failedr: %v\n", err)
		return
	}

	for _, feed := range feeds {
		URLs, newestUpdate, err := getFeedItemsSince(db, feed.URL, feed.LastItemDate)
		if err != nil {
			log.Printf("Failed loading %s: %v\n", feed.URL, err)
			continue
		}
		if len(URLs) == 0 {
			continue
		}

		stmt, err := videostorage.PrepareLookupByURL(db)
		if err != nil {
			panic(err)
		}

		for _, URL := range URLs {
			video := &videostorage.Video{Show: feed.Show, URL: URL}

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
		}

		feed.LastItemDate = newestUpdate
		feedstorage.Update(db, feed)
	}
}

func getFeedItemsSince(db *sqlx.DB, URL, lastDate string) ([]string, string, error) {
	fp := gofeed.NewParser()
	feed, err := fp.ParseURL(URL)
	if err != nil {
		return nil, "", err
	}

	newestUpdate := ""
	result := []string{}
	for _, item := range feed.Items {
		if item.Published <= lastDate {
			break
		}
		if item.Published > newestUpdate {
			newestUpdate = item.Published
		}
		result = append(result, item.Link)
	}
	return result, newestUpdate, err
}
