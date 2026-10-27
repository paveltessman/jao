package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/paveltessman/jao/keys"
	"github.com/paveltessman/jao/settings"
)

// runKey takes an optional container name. Without it, the node must have
// exactly one protocol container.
func runKey(ctx context.Context, cfg settings.Settings, args []string) error {
	if len(args) > 1 {
		return errors.New("usage: jao key [name]")
	}
	var name string
	if len(args) == 1 {
		name = args[0]
	}

	ctr, closeFn, err := newContainerService(ctx)
	if err != nil {
		return err
	}
	defer closeFn()

	clientConf, err := keys.New(ctr, cfg.PublicHost).CreateKey(ctx, name)
	if err != nil {
		return err
	}

	_, err = fmt.Fprint(os.Stdout, clientConf)
	return err
}
