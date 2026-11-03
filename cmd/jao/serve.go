package main

import (
	"context"
	"errors"

	"github.com/paveltessman/jao/httpapi"
	"github.com/paveltessman/jao/settings"
)

func runServe(ctx context.Context, cfg settings.Settings, args []string) error {
	if len(args) > 0 {
		return errors.New("usage: jao list")
	}

	return httpapi.Serve(ctx, cfg)
}
