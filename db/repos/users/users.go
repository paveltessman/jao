package users

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"
	"uuid"

	"github.com/paveltessman/jao/auth"
	"github.com/paveltessman/jao/db/internal/dbtools"
	"github.com/paveltessman/jao/db/internal/sqlc"
	"github.com/paveltessman/jao/users"
)

var _ auth.UsersRepo = (*Repo)(nil)

type Repo struct {
	queries *sqlc.Queries
}

func New(db *sql.DB) *Repo {
	r := &Repo{queries: sqlc.New(db)}
	return r
}

func (r *Repo) Insert(ctx context.Context, user users.User) error {

	now := time.Now()
	args := sqlc.UserInsertParams{
		ID:             user.ID.String(),
		Name:           user.Name,
		Lastseen:       user.LastSeen,
		Active:         user.Active,
		Passhash:       user.PassHash,
		RefreshTokenID: user.RefreshTokenID,
		UpdatedAt:      now,
		CreatedAt:      now,
	}

	err := r.queries.UserInsert(ctx, args)
	if err != nil {
		return fmt.Errorf("%w: %w", dbtools.ErrDb, err)
	}

	return nil
}

func (r *Repo) Get(ctx context.Context, id uuid.UUID) (users.User, error) {

	u, err := r.queries.UserGetByID(ctx, id.String())
	if err != nil {
		return users.User{}, fmt.Errorf("%w: %w", dbtools.ErrDb, err)
	}

	return sqlcToUser(u)
}

func (r *Repo) Update(ctx context.Context, user users.User) error {

	args := sqlc.UserUpdateParams{
		Lastseen:       user.LastSeen,
		Active:         user.Active,
		Passhash:       user.PassHash,
		RefreshTokenID: user.RefreshTokenID,
		UpdatedAt:      time.Now(),
		ID:             user.ID.String(),
	}

	log.Println(args.Active)

	err := r.queries.UserUpdate(ctx, args)
	if err != nil {
		return fmt.Errorf("%w: %w", dbtools.ErrDb, err)
	}

	return nil
}

func sqlcToUser(u sqlc.AppUser) (users.User, error) {
	id, err := uuid.Parse(u.ID)
	if err != nil {
		return users.User{}, err
	}

	user := users.User{
		ID:             id,
		Name:           u.Name,
		LastSeen:       u.Lastseen,
		Active:         u.Active,
		PassHash:       u.Passhash,
		RefreshTokenID: u.RefreshTokenID,
	}

	return user, nil
}
