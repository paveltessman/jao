package main

import (
	"context"
	"fmt"
	"os"

	"github.com/paveltessman/jao/protocol/awg"
	"github.com/paveltessman/jao/settings"
)

func runKey(ctx context.Context, cfg settings.Settings, _ []string) error {
	inst, closeFn, err := newInstaller(ctx, cfg)
	if err != nil {
		return err
	}
	defer closeFn()

	clientConf, err := inst.AddAwgPeer(ctx, awg.DefaultConfig(), cfg.PublicHost)
	if err != nil {
		return err
	}

	_, err = fmt.Fprint(os.Stdout, clientConf)
	return err
}
