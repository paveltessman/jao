package users

import (
	"testing"
	"uuid"

	"github.com/paveltessman/jao/db/migrations"
	"github.com/paveltessman/jao/users"
)

func newUser() users.User {
	u := users.User{
		ID:   uuid.New(),
		Name: "username",
	}
	return u
}

func insert(t *testing.T, repo *Repo, user users.User) {
	t.Helper()

	err := repo.Insert(t.Context(), user)
	if err != nil {
		t.Fatalf("error when inserting new user: %v", err)
	}

}

func selectByID(t *testing.T, repo *Repo, id uuid.UUID) users.User {
	t.Helper()

	user, err := repo.Get(t.Context(), id)
	if err != nil {
		t.Fatalf("error when selecting user: %v", err)
	}

	return user
}

func TestInsertGet(t *testing.T) {

	repo := New(migrations.DBForTestsWithMigrations(t))

	insertedUser := newUser()

	insert(t, repo, insertedUser)

	loadedUser := selectByID(t, repo, insertedUser.ID)

	if loadedUser != insertedUser {
		t.Errorf("inserted and loaded users do not match: loaded=%v, inserted=%v", loadedUser, insertedUser)
	}
}

func TestUpdateSetsActive(t *testing.T) {

	repo := New(migrations.DBForTestsWithMigrations(t))

	user := newUser()
	user.Active = true

	insert(t, repo, user)

	user.Active = false

	err := repo.Update(t.Context(), user)
	if err != nil {
		t.Fatalf("error when updating user: %v", err)
	}

	updatedUser := selectByID(t, repo, user.ID)

	if updatedUser.Active {
		t.Errorf("active flag wasn't changed")
	}
}

func TestUpdateSetsRefreshTokenID(t *testing.T) {

	repo := New(migrations.DBForTestsWithMigrations(t))

	user := newUser()
	user.RefreshTokenID = "opa"

	insert(t, repo, user)

	user.RefreshTokenID = "opa x2"

	err := repo.Update(t.Context(), user)
	if err != nil {
		t.Fatalf("error when updating user: %v", err)
	}

	updatedUser := selectByID(t, repo, user.ID)

	if updatedUser.RefreshTokenID != "opa x2" {
		t.Errorf("active flag wasn't changed")
	}
}
