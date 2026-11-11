package users

import (
	"context"
	"errors"
	"strings"
	"testing"
)

type fakeUsersRepo struct {
	callArgsInsert []User
	errInsert      error
}

func (r *fakeUsersRepo) Insert(ctx context.Context, user User) error {
	r.callArgsInsert = append(r.callArgsInsert, user)
	return r.errInsert
}

func TestNewUser(t *testing.T) {

	repo := &fakeUsersRepo{}

	svc := NewService(repo)

	user, err := svc.NewUser(t.Context(), "username", "password")
	if err != nil {
		t.Fatalf("error when creating new user: %v", err)
	}

	if len(repo.callArgsInsert) != 1 {
		t.Fatalf("unexpected call args for user insertion: %v", repo.callArgsInsert)
	}

	if user != repo.callArgsInsert[0] {
		t.Fatalf("returned user and inserted user do not match")
	}
}

func TestNewUserOnDbError(t *testing.T) {

	dbError := errors.New("")
	repo := &fakeUsersRepo{
		errInsert: dbError,
	}

	svc := NewService(repo)

	_, err := svc.NewUser(t.Context(), "username", "password")
	if !errors.Is(err, dbError) {
		t.Fatalf("expecting db error, got something else: %v", err)
	}
}

func TestNewUserWithBadName(t *testing.T) {

	svc := NewService(&fakeUsersRepo{})

	cases := []struct {
		desc          string
		username      string
		expectedError string
	}{
		{"is too short", "", "too short"},
		{"is too long", "aaaa_aaaa_aaaa_aaaa_aaaa_aaaa_aaaa_aaaa", "too long"},
		{"contains space", "aaaa aaaa", "whitespaces"},
		{"contains tab", "aaaa\taaaa", "whitespaces"},
		{"contains newline", "aaaa\naaaa", "whitespaces"},
		{"contains carriage return", "aaaa\raaaa", "whitespaces"},
		{"contains non-ascii chars", "ы", "non-ascii"},
	}

	for _, c := range cases {
		_, err := svc.NewUser(t.Context(), c.username, "password")
		if !errors.Is(err, ErrBadString) {
			t.Fatalf("when username %s: expecting validation error, got something else: %v", c.desc, err)
		}

		if !strings.Contains(err.Error(), c.expectedError) {
			t.Errorf("when username %s: expecting %q in the error message, but got %q", c.desc, c.expectedError, err)
		}
	}
}

func TestNewUserWithBadPassword(t *testing.T) {

	svc := NewService(&fakeUsersRepo{})

	cases := []struct {
		desc          string
		password      string
		expectedError string
	}{
		{"is too short", "1234567", "too short"},
		{"is too long", "aaaa_aaaa_aaaa_aaaa_aaaa_aaaa_aaaa_aaaa", "too long"},
	}

	for _, c := range cases {
		_, err := svc.NewUser(t.Context(), "bot777", c.password)
		if !errors.Is(err, ErrBadString) {
			t.Fatalf("when passowrd %s: expecting validation error, got something else: %v", c.desc, err)
		}

		if !strings.Contains(err.Error(), c.expectedError) {
			t.Errorf("when password %s: expecting %q in the error message, but got %q", c.desc, c.expectedError, err)
		}
	}
}
