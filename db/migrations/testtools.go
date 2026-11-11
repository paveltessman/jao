package migrations

import (
	"database/sql"
	"os"
	"testing"
)

const dbFile = "testdb.sqlite"

func removeDBFile(t *testing.T) {
	t.Helper()
	err := os.Remove(dbFile)
	if os.IsNotExist(err) {
		return
	}
	if err != nil {
		t.Fatalf("can't remove database file after test: %v", err)
	}
}

func DBForTests(t *testing.T) *sql.DB {
	t.Helper()

	removeDBFile(t)

	db, err := sql.Open("sqlite", dbFile)
	if err != nil {
		t.Fatalf("unable to create db for tests: %v", err)
	}

	t.Cleanup(func() { removeDBFile(t) })

	return db
}

func DBForTestsWithMigrations(t *testing.T) *sql.DB {
	t.Helper()

	db := DBForTests(t)

	svc, err := NewService(db)
	if err != nil {
		t.Fatalf("unable to create service for migrations: %v", err)
	}

	_, err = svc.Up(t.Context())
	if err != nil {
		t.Fatalf("unable to apply migrations for test db: %v", err)
	}

	return db
}
