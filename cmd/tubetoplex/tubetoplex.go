package main

import (
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/jasonpenny/tubetoplex/internal/filecopier"
	"github.com/jasonpenny/tubetoplex/internal/plexshowupdater"
	"github.com/jasonpenny/tubetoplex/internal/showstorage"
	"github.com/jasonpenny/tubetoplex/internal/videostorage"
	"github.com/jasonpenny/tubetoplex/internal/youtubedl"
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
	showstorage.SetupTable(db)

	pullNewPosts(db)
	applyShowNumbersToNewPosts(db)
	downloadNumberedVideos(db)
	createNFOs(db)
	copyFiles(db)
}

func pullNewPosts(db *sqlx.DB) {
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

func applyShowNumbersToNewPosts(db *sqlx.DB) {
	videos, err := videostorage.FindForStep(db, "new")
	if err != nil {
		panic(err)
	}

	for _, video := range videos {
		show, err := showstorage.Find(db, video.Show)
		if err != nil {
			log.Printf("APPLY_NUMBER: finding show '%s' failed: %+v\n", video.Show, err)
			continue
		}

		show.NextEpisode++
		if _, err := showstorage.Update(db, show); err != nil {
			log.Printf("NUMBERED: failed to update [show] %v\n", err)
			continue
		}

		video.SeasonNum = show.LatestSeason
		video.EpisodeNum = show.NextEpisode
		if _, err := videostorage.Update(db, &video, "numbered"); err != nil {
			log.Printf("NUMBERED: failed to update [video] %v\n", err)
			continue
		}

		log.Printf("NUMBERED: updated video %s", video.URL)
	}
}

func downloadNumberedVideos(db *sqlx.DB) {
	videos, err := videostorage.FindForStep(db, "numbered")
	if err != nil {
		panic(err)
	}

	for _, video := range videos {
		log.Printf("DOWNLOAD: Started downloading video %s", video.URL)
		vi := youtubedl.DownloadURL(video.URL, video.SeasonNum, video.EpisodeNum)

		// store more details
		video.Filename = vi.Filename
		video.Title = vi.Title
		video.Description = vi.Description
		video.AverageRating = vi.Rating
		video.UploadDate = vi.UploadDate

		// transition to downloaded
		if _, err := videostorage.Update(db, &video, "downloaded"); err != nil {
			log.Printf("DOWNLOAD: download failed %v\n", err)
			continue
		}

		log.Printf("DOWNLOAD: Finished downloading video %s", video.URL)
	}
}

func createNFOs(db *sqlx.DB) {
	videos, err := videostorage.FindForStep(db, "downloaded")
	if err != nil {
		panic(err)
	}

	for _, video := range videos {
		err = plexshowupdater.CreateNFOFile(
			video.Title,
			video.SeasonNum,
			video.EpisodeNum,
			video.Description,
			video.AverageRating,
			video.UploadDate,
			video.Filename,
		)
		if err != nil {
			log.Printf("NFOED: Unable to create NFO file %v\n", err)
			continue
		}

		if _, err = videostorage.Update(db, &video, "nfoed"); err != nil {
			log.Printf("NFOED: Unable to update [video] %v\n", err)
		}
	}
}

func copyFiles(db *sqlx.DB) {
	videos, err := videostorage.FindForStep(db, "nfoed")
	if err != nil {
		panic(err)
	}

	for _, video := range videos {
		show, err := showstorage.Find(db, video.Show)
		if err != nil {
			log.Printf("COPY_FILES: finding show '%s' failed: %+v\n", video.Show, err)
			continue
		}

		nfoFile := plexshowupdater.NFOFilenameForVideo(video.Filename)

		err = filecopier.CopyFile(
			nfoFile,
			filepath.Join(
				show.Path,
				filepath.Base(nfoFile),
			),
		)

		if err == nil {
			err = filecopier.CopyFile(
				video.Filename,
				filepath.Join(
					show.Path,
					filepath.Base(video.Filename),
				),
			)
		}

		if err != nil {
			log.Printf("COPY_FILES: Error, could not copy files %v\n", err)
			continue
		}

		if _, err := videostorage.Update(db, &video, "copied"); err != nil {
			log.Printf("COPY_FILES: Unable to update [video] %v\n", err)
			continue
		}

		log.Printf("COPY_FILES: Finished copying video %s\n", video.URL)

		downloadDir := filepath.Dir(video.Filename)
		if err := os.RemoveAll(downloadDir); err != nil {
			log.Printf("COPY_FILES: Error, could not delete video dir: %v\n", err)
		}
	}
}
