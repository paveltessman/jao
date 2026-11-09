package db

import (
	"database/sql"

	_ "modernc.org/sqlite"
)

const dsnString = "db.sqlite?_journal_mode=WAL&_busy_timeout=5000&_foreign_keys=ON"

func New() (*sql.DB, error) {

	db, err := sql.Open("sqlite", dsnString)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(1)

	return db, nil
}
