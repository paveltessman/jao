package main

import (
	"context"
	"fmt"
	"strings"

	database "github.com/paveltessman/jao/db"
	"github.com/paveltessman/jao/db/migrations"
	"github.com/paveltessman/jao/settings"
)

type actionFunc = func(ctx context.Context) (string, error)

var actions = []string{"up", "down", "version"}

func usageMigrate() error {
	return fmt.Errorf("migrage: invalid arguments; usage: jao migrate <%s>", strings.Join(actions, " | "))

}

func runMigrate(ctx context.Context, cfg settings.Settings, args []string) error {
	if len(args) != 1 {
		return usageMigrate()
	}

	db, err := database.New()
	if err != nil {
		return err
	}

	defer func() {
		if closeErr := db.Close(); closeErr != nil {
			fmt.Printf("error when closing db: %v\n", closeErr)
		}
	}()

	svc, err := migrations.NewService(db)
	if err != nil {
		return err
	}

	runAction, err := selectAction(svc, args[0])
	if err != nil {
		return err
	}
	result, err := runAction(ctx)
	if err != nil {
		return err
	}
	fmt.Println(result)
	return nil
}

func selectAction(svc *migrations.Service, action string) (actionFunc, error) {

	var runAction actionFunc

	switch action {
	case actions[0]:
		runAction = svc.Up
	case actions[1]:
		runAction = svc.Down
	case actions[2]:
		runAction = svc.Version
	default:
		return nil, usageMigrate()
	}

	return runAction, nil

}
