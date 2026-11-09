package migrations

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"strconv"
	"strings"

	"github.com/pressly/goose/v3"
)

//go:embed sql/*.sql
var sqlFiles embed.FS

type Service struct {
	provider *goose.Provider
}

func NewService(db *sql.DB) (*Service, error) {

	migrationDir, err := fs.Sub(sqlFiles, "sql")
	if err != nil {
		return nil, fmt.Errorf("can't root migration files: %w", err)
	}

	provider, err := goose.NewProvider(goose.DialectSQLite3, db, migrationDir)
	if err != nil {
		return nil, err
	}

	svc := &Service{provider}

	return svc, nil
}

func (s *Service) Up(ctx context.Context) (string, error) {

	result, err := s.provider.Up(ctx)
	if err != nil {
		return "", err
	}

	if len(result) == 0 {
		return "migrations: schema is up to date", nil
	}

	var buf strings.Builder

	for i, r := range result {
		fmt.Fprintf(&buf, "migrations: applied %s", r.String())
		if i != len(result)-1 {
			buf.WriteByte('\n')
		}
	}

	return buf.String(), nil
}

func (s *Service) Down(ctx context.Context) (string, error) {

	result, err := s.provider.Down(ctx)

	if errors.Is(err, goose.ErrNoNextVersion) {
		return "migrations: no next version to roll back", nil
	}

	if err != nil {
		return "", err
	}

	out := fmt.Sprintf("migrations: rolled back %s", result.String())

	return out, nil
}

func (s *Service) Version(ctx context.Context) (string, error) {
	version, err := s.provider.GetDBVersion(ctx)
	if err != nil {
		return "", err
	}
	return strconv.Itoa(int(version)), nil
}
