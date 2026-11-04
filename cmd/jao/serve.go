package main

import (
	"context"
	"errors"

	"github.com/paveltessman/jao/httpapi"
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

	deps := httpapi.NewDeps(
		keys.NewService(ctr, cfg.PublicHost),
	)

	return httpapi.Serve(ctx, cfg, deps)
}
