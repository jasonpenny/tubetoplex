package main

import (
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/jasonpenny/tubetoplex/internal/filecopier"
	"github.com/jasonpenny/tubetoplex/internal/plexshowupdater"
	"github.com/jasonpenny/tubetoplex/internal/showstorage"
	"github.com/jasonpenny/tubetoplex/internal/videostorage"
	"github.com/jasonpenny/tubetoplex/internal/youtubedl"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	run()
	for _ = range time.NewTicker(time.Hour).C {
		run()
	}
}

func run() {
	log.Printf("RUN: %s\n", time.Now().Format(time.RFC850))

	var db *sqlx.DB
	var err error

	db, err = sqlx.Connect("sqlite3", "__videos.db")
	if err != nil {
		panic("Could not open sqlite file")
	}

	videostorage.SetupTable(db)
	showstorage.SetupTable(db)

	pullNewFeeds(db)
	pullNewTumblrPosts(db)
	applyShowNumbersToNewPosts(db)
	downloadNumberedVideos(db)
	createNFOs(db)
	copyFiles(db)
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
		vi, err := youtubedl.DownloadURL(video.URL, video.SeasonNum, video.EpisodeNum)
		if err != nil {
			log.Printf("DOWNLOAD: Failed to start download: %v\n", err)
			if _, err := videostorage.Update(db, &video, "failed-download"); err != nil {
				log.Printf("DOWNLOAD: Additionally, updating db failed to set as failed-download  %v\n", err)
			}
			continue
		}

		// store more details
		video.Filename = vi.Filename
		video.Title = vi.Title
		video.Description = vi.Description
		video.AverageRating = vi.Rating
		video.UploadDate = vi.UploadDate

		// transition to downloaded
		if _, err := videostorage.Update(db, &video, "downloaded"); err != nil {
			log.Printf("DOWNLOAD: Updating db failed to set as downloaded  %v\n", err)
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
