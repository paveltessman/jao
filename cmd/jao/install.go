package main

import (
	"context"

	"github.com/paveltessman/jao/installer"
	"github.com/paveltessman/jao/protocol/awg"
	"github.com/paveltessman/jao/settings"
)

func runInstall(ctx context.Context, cfg settings.Settings, _ []string) error {
	ctr, closeFn, err := newContainerService(ctx)
	if err != nil {
		return err
	}
	defer closeFn()

	return installer.New(ctr, cfg.WorkDir).InstallAwg(ctx, awg.DefaultConfig())
}
