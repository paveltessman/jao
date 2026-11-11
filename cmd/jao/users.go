package main

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/paveltessman/jao/db"
	usersdb "github.com/paveltessman/jao/db/repos/users"
	"github.com/paveltessman/jao/settings"
	"github.com/paveltessman/jao/users"
)

var userActions = map[string]func(args []string) error{
	"create": createUser,
}

func usageUsers() error {
	actions := slices.Collect(maps.Keys(userActions))
	slices.Sort(actions)

	return fmt.Errorf("usage: jao users <%s>", strings.Join(actions, " | "))
}

func usageUsersCreate() error {
	return errors.New("usage: jao users create <name> <password>")
}

func runUsers(ctx context.Context, cfg settings.Settings, args []string) error {
	if len(args) < 1 {
		return usageUsers()
	}
	action, found := userActions[args[0]]
	if !found {
		return usageUsers()
	}
	return action(args[1:])
}

func createUser(args []string) error {

	if len(args) != 2 {
		return usageUsersCreate()
	}

	db, err := db.New()
	if err != nil {
		return err
	}
	svc := users.NewService(usersdb.New(db))

	user, err := svc.NewUser(context.Background(), args[0], args[1])
	if err != nil {
		return err
	}

	fmt.Println(user.ID)
	return nil
}
