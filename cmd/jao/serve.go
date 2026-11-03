package main

import (
	"context"
	"errors"

	"github.com/paveltessman/jao/settings"
)

func runServe(ctx context.Context, cfg settings.Settings, args []string) error {
	if len(args) > 0 {
		return errors.New("usage: jao list")
	}

	return errors.New("the 'serve' command is not implemented yet")
}
