package users

import (
	"context"
	"fmt"
	"time"
	"uuid"

	"github.com/paveltessman/jao/password"
)

type User struct {
	ID             uuid.UUID
	Name           string
	LastSeen       time.Time
	Active         bool
	PassHash       string
	RefreshTokenID string
}

type UsersRepo interface {
	Insert(ctx context.Context, user User) error
}

type Service struct {
	repo UsersRepo
}

func NewService(repo UsersRepo) *Service {
	svc := &Service{repo}
	return svc
}

func (s *Service) NewUser(ctx context.Context, name string, passwd string) (User, error) {

	if err := validateName(name); err != nil {
		return User{}, err
	}

	if err := validatePassword(passwd); err != nil {
		return User{}, err
	}

	id := uuid.New()

	passwdHash, err := password.Hash(passwd)
	if err != nil {
		return User{}, fmt.Errorf("unable to create password hash: %w", err)
	}

	user := User{
		ID:       id,
		Name:     name,
		Active:   true,
		PassHash: passwdHash,
	}

	err = s.repo.Insert(ctx, user)
	if err != nil {
		return User{}, fmt.Errorf("unable to save new user: %w", err)
	}

	return user, nil
}
