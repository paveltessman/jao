package migrations

import (
	"io/fs"
	"testing"
)

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

func TestUpAndDown(t *testing.T) {

	db := DBForTests(t)

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
