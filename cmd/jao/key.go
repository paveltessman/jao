package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/paveltessman/jao/keys"
	"github.com/paveltessman/jao/settings"
)

func runKey(ctx context.Context, cfg settings.Settings, args []string) error {
	if len(args) != 1 {
		return errors.New("usage: jao key <protocol name>")
	}
	name := args[0]

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
