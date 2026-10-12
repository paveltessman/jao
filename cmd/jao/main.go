package main

import (
	"context"
	"fmt"

	"github.com/paveltessman/jao/container"
)

func main() {
	svc, _ := container.New()
	defer func() {
		if err := svc.Close(); err != nil {
			fmt.Println(err)
		}
	}()

	result, err := svc.Ping(context.Background())
	fmt.Printf("ping: %s, err: %v\n", result, err)
}
