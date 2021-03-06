package videostorage

import (
	"database/sql"

	"github.com/jmoiron/sqlx"
)

// Video represents a video file and metadata.
type Video struct {
	ID            int     `db:"id"`
	URL           string  `db:"url"`
	Show          string  `db:"show"`
	Filename      string  `db:"filename"`
	Title         string  `db:"title"`
	Description   string  `db:"description"`
	AverageRating float64 `db:"average_rating"`
	UploadDate    string  `db:"upload_date"`
	SeasonNum     int     `db:"season_num"`
	EpisodeNum    int     `db:"episode_num"`
	Step          string  `db:"step"`
}

// SetupTable creates the shows table if it does not exist.
func SetupTable(db *sqlx.DB) {
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
		season_num INTEGER,
		episode_num INTEGER,
		step VARCHAR
	);
	`)
}

// Add puts a new video into the database.
func Add(db *sqlx.DB, video *Video, step string) (sql.Result, error) {
	video.Step = step

	return db.NamedExec(
		`INSERT INTO videos (
			url, show, filename, title, description,
			average_rating, upload_date,
			season_num, episode_num, step
		)
		VALUES (
			:url, :show, :filename, :title, :description,
			:average_rating, :upload_date,
			:season_num, :episode_num, :step
		)`,
		&video,
	)
}

// Update stores new data in the database.
func Update(db *sqlx.DB, video *Video, step string) (sql.Result, error) {
	video.Step = step

	return db.NamedExec(
		`
		UPDATE videos SET
			url = :url,
			show = :show,
			filename = :filename,
			title = :title,
			description = :description,
			average_rating = :average_rating,
			upload_date = :upload_date,
			season_num = :season_num,
			episode_num = :episode_num,
			step = :step
		WHERE id = :id
		`,
		&video,
	)
}

// PrepareLookupByURL returns a reusable statement for looking up a video by url.
func PrepareLookupByURL(db *sqlx.DB) (*sqlx.NamedStmt, error) {
	return db.PrepareNamed(`SELECT * FROM videos WHERE url = :url`)
}

// Find looks up a video by url.
func Find(stmt *sqlx.NamedStmt, video *Video) ([]Video, error) {
	result := []Video{}
	err := stmt.Select(&result, video)
	return result, err
}

// FindForStep returns all videos for the step.
func FindForStep(db *sqlx.DB, step string) ([]Video, error) {
	videos := []Video{}
	stmt, err := db.Preparex(`SELECT * FROM videos WHERE step = ? ORDER BY id`)
	if err != nil {
		return videos, err
	}
	err = stmt.Select(&videos, step)
	return videos, err
}
