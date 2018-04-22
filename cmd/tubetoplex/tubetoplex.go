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
		for _, post := range allPosts {
			video := &videostorage.Video{}

			switch pt := post.(type) {
			case *tumblr.LinkPost:
				video.Url = pt.Url
				for _, tag := range pt.Tags {
					if tag != "unprocessed" {
						video.Show = strings.ToLower(tag)
						break
					}
				}
			case *tumblr.VideoPost:
				video.Url = pt.PermalinkUrl
				for _, tag := range pt.Tags {
					if tag != "unprocessed" {
						video.Show = strings.ToLower(tag)
						break
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
			default:
				continue
			}

			videos, err := videostorage.Find(stmt, video)
			if err != nil {
				panic(err)
			}

			if len(videos) > 0 {
				// this video has already been stored, stop paging through posts
				break
			}

			_, err = videostorage.Add(db, video, "new")
			if err != nil {
				panic(err)
			}
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
			log.Printf("%+v\n", err)
			continue
		}

		show.NextEpisode++
		showstorage.Update(db, show)

		video.SeasonNum = show.LatestSeason
		video.EpisodeNum = show.NextEpisode
		videostorage.Update(db, &video, "numbered")
	}
}

func downloadNumberedVideos(db *sqlx.DB) {
	videos, err := videostorage.FindForStep(db, "numbered")
	if err != nil {
		panic(err)
	}

	for _, video := range videos {
		vi := youtubedl.DownloadURL(video.Url, video.SeasonNum, video.EpisodeNum)

		// store more details
		video.Filename = vi.Filename
		video.Title = vi.Title
		video.Description = vi.Description
		video.AverageRating = vi.Rating
		video.UploadDate = vi.UploadDate

		// transition to downloaded
		videostorage.Update(db, &video, "downloaded")
	}
}

func createNFOs(db *sqlx.DB) {
	videos, err := videostorage.FindForStep(db, "downloaded")
	if err != nil {
		panic(err)
	}

	for _, video := range videos {
		plexshowupdater.CreateNFOFile(
			video.Title,
			video.SeasonNum,
			video.EpisodeNum,
			video.Description,
			video.AverageRating,
			video.UploadDate,
			video.Filename,
		)

		videostorage.Update(db, &video, "nfoed")
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
			log.Printf("%+v\n", err)
			continue
		}

		nfoFile := plexshowupdater.NFOFilenameForVideo(video.Filename)

		filecopier.CopyFile(
			nfoFile,
			filepath.Join(
				show.Path,
				filepath.Base(nfoFile),
			),
		)

		// copy file
		filecopier.CopyFile(
			video.Filename,
			filepath.Join(
				show.Path,
				filepath.Base(video.Filename),
			),
		)

		videostorage.Update(db, &video, "copied")
	}
}
