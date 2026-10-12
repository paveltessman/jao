package main

import (
	"context"
	"fmt"

	"github.com/paveltessman/jao/container"
	"github.com/paveltessman/jao/installer"
)

func main() {
	ctrSvc, _ := container.New()
	defer func() {
		if err := ctrSvc.Close(); err != nil {
			fmt.Println(err)
		}
	}()
	result, err := ctrSvc.Ping(context.Background())
	fmt.Printf("ping: %s, err: %v\n", result, err)

	inst := installer.New(ctrSvc)

	err = inst.InstallAwg(context.Background())
	if err != nil {
		fmt.Println(err)
	}
}
