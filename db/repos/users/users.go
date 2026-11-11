package users

import (
	"context"
	"database/sql"
	"fmt"
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

func (r *Repo) Get(ctx context.Context, id uuid.UUID) (users.User, error) {

	u, err := r.queries.UserGetByID(ctx, id.String())
	if err != nil {
		return users.User{}, fmt.Errorf("%w: %w", dbtools.ErrDb, err)
	}

	return sqlcToUser(u)
}

func (r *Repo) Update(ctx context.Context, user users.User) error {

	args := sqlc.UserUpdateParams{
		Lastseen:       user.LastSeen.Unix(),
		Active:         user.Active,
		Passhash:       user.PassHash,
		RefreshTokenID: user.RefreshTokenID,
		UpdatedAt:      time.Now().Unix(),
	}

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
		LastSeen:       dbtools.Int64ToUnixTime(u.Lastseen),
		Active:         u.Active,
		PassHash:       u.Passhash,
		RefreshTokenID: u.RefreshTokenID,
	}

	return user, nil
}
