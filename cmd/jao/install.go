package main

import (
	"context"

	"github.com/paveltessman/jao/installer"
	"github.com/paveltessman/jao/settings"
)

func runInstall(ctx context.Context, cfg settings.Settings, _ []string) error {
	inst, closeFn, err := newInstaller(ctx, cfg)
	if err != nil {
		return err
	}
	defer closeFn()

	return inst.InstallAwg(ctx, installer.DefaultAwgConfig())
}
