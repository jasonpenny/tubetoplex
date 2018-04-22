package showstorage

import (
	"database/sql"

	"github.com/jmoiron/sqlx"
)

type Show struct {
	Name         string `db:"name"`
	LatestSeason int    `db:"latest_season"`
	NextEpisode  int    `db:"next_episode"`
	Path         string `dB:"path"`
}

func SetupTable(db *sqlx.DB) {
	db.MustExec(`
	CREATE TABLE IF NOT EXISTS shows (
		name VARCHAR,
		latest_season INTEGER,
		next_episode INTEGER,
		path VARCHAR
	);
	`)
}

func Find(db *sqlx.DB, name string) (*Show, error) {
	show := Show{}
	err := db.Get(&show, `SELECT * FROM shows WHERE name = ?`, name)
	return &show, err
}

func Update(db *sqlx.DB, show *Show) (sql.Result, error) {
	return db.NamedExec(
		`
			UPDATE shows SET
				latest_season = :latest_season,
				next_episode = :next_episode,
				path = :path
			WHERE name = :name
		`,
		&show,
	)
}
