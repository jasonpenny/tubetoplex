package feedstorage

import (
	"database/sql"

	"github.com/jmoiron/sqlx"
)

// Feed stores an RSS/Atom feed to automatically grab videos.
type Feed struct {
	ID           int    `db:"id"`
	Show         string `db:"show"`
	URL          string `db:"url"`
	LastItemDate string `db:"last_update"`
}

// SetupFeedTable creates the feeds table if it does not exist.
func SetupFeedTable(db *sqlx.DB) {
	db.MustExec(`
	CREATE TABLE IF NOT EXISTS feeds (
		id INTEGER PRIMARY KEY,
		show VARCHAR,
		url TEXT,
		last_update VARCHAR
	);
	`)
}

// Update stores new data in the database.
func Update(db *sqlx.DB, feed Feed) (sql.Result, error) {
	return db.NamedExec(
		`
		UPDATE feeds SET
			show = :show,
			url = :url,
			last_update = :last_update
		WHERE id = :id
		`,
		&feed,
	)
}

// GetAllFeeds retrieves all feeds from the database in id order.
func GetAllFeeds(db *sqlx.DB) ([]Feed, error) {
	feeds := []Feed{}
	err := db.Select(&feeds, "SELECT * FROM feeds ORDER BY id")
	return feeds, err
}
