package migrations

import (
	"database/sql"
	"io/fs"
	"os"
	"testing"

	_ "modernc.org/sqlite"
)

const dbFile = "testdb.sqlite"

func countMigrations(t *testing.T) int {
	t.Helper()

	migrationDir, err := fs.Sub(sqlFiles, "sql")
	if err != nil {
		t.Fatalf("error loading migration files: %v", err)
	}
	names, err := fs.Glob(migrationDir, "*.sql")
	if err != nil {
		t.Fatalf("error counting migration files: %v", err)
	}
	return len(names)

}

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

func db(t *testing.T) (*sql.DB, error) {
	t.Helper()

	removeDBFile(t)

	db, err := sql.Open("sqlite", dbFile)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func TestUpAndDown(t *testing.T) {

	db, err := db(t)
	if err != nil {
		t.Fatalf("can't create database: %v", err)
	}

	t.Cleanup(func() {
		removeDBFile(t)
	})

	migrationsCount := countMigrations(t)

	svc, err := NewService(db)
	if err != nil {
		t.Fatalf("can't create migration service: %v", err)
	}

	for range 2 {
		_, err := svc.Up(t.Context())
		if err != nil {
			t.Fatalf("error when applying pending migrations: %v", err)
		}

		for range migrationsCount {
			_, err := svc.Down(t.Context())
			if err != nil {
				t.Fatalf("error when rolling back migrations: %v", err)
			}
		}

	}
}
