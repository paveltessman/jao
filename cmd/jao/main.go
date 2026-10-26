package main

import (
	"context"
	"fmt"
	"os"

	"github.com/paveltessman/jao/container"
	"github.com/paveltessman/jao/installer"
	"github.com/paveltessman/jao/settings"
)

func main() {
	cfg := settings.New()

	ctrSvc, err := container.New()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer func() {
		if err := ctrSvc.Close(); err != nil {
			fmt.Println(err)
		}
	}()
	result, err := ctrSvc.Ping(context.Background())
	fmt.Printf("ping: %s, err: %v\n", result, err)

	inst := installer.New(ctrSvc, cfg.WorkDir)

	awgCfg := installer.DefaultAwgConfig()
	err = inst.InstallAwg(context.Background(), awgCfg)
	if err != nil {
		fmt.Println(err)
		return
	}

	clientConf, err := inst.AddAwgPeer(context.Background(), awgCfg, cfg.PublicHost)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Print(clientConf)
}
