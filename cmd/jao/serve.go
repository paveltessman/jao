package main

import (
	"context"
	"errors"

	"github.com/paveltessman/jao/auth"
	"github.com/paveltessman/jao/db"
	"github.com/paveltessman/jao/db/repos/users"
	"github.com/paveltessman/jao/httpapi"
	"github.com/paveltessman/jao/httpapi/shared"
	"github.com/paveltessman/jao/keys"
	"github.com/paveltessman/jao/settings"
)

func runServe(ctx context.Context, cfg settings.Settings, args []string) error {
	if len(args) > 0 {
		return errors.New("usage: jao list")
	}

	ctr, closeFn, err := newContainerService(ctx)
	if err != nil {
		return err
	}
	defer closeFn()

	db, err := db.New()
	if err != nil {
		return err
	}

	deps := shared.NewDeps(
		keys.NewService(ctr, cfg.PublicHost),
		auth.NewService(users.New(db), cfg.SecretKey),
	)

	return httpapi.Serve(ctx, cfg, deps)
}
