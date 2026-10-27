package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/paveltessman/jao/container"
	"github.com/paveltessman/jao/settings"
)

type command struct {
	name    string
	summary string
	run     func(ctx context.Context, cfg settings.Settings, args []string) error
}

func commands() []command {
	commands := []command{
		{"install", "install the AWG container", runInstall},
		{"key", "create a key in a protocol container and print its client config", runKey},
		{"list", "list the protocol containers on the node", runList},
	}
	return commands
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := do(ctx, os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "fatal: %v\n", err)
		stop()
		os.Exit(1)
	}
}

func do(ctx context.Context, args []string) error {
	if len(args) == 0 {
		usage()
		return errors.New("no subcommand given")
	}

	for _, c := range commands() {
		if c.name == args[0] {
			return c.run(ctx, settings.New(), args[1:])
		}
	}

	usage()
	return fmt.Errorf("unknown subcommand %q", args[0])
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage: jao <command> [args]")
	fmt.Fprintln(os.Stderr, "\ncommands:")
	for _, c := range commands() {
		fmt.Fprintf(os.Stderr, "  %-8s %s\n", c.name, c.summary)
	}
}

// newContainerService connects to the Docker engine and makes sure that it replies.
// The caller must call the returned close function when it is done.
func newContainerService(ctx context.Context) (*container.Service, func(), error) {
	ctrSvc, err := container.New()
	if err != nil {
		return nil, nil, err
	}
	closeFn := func() {
		if err := ctrSvc.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "close docker client: %v\n", err)
		}
	}

	if _, err := ctrSvc.Ping(ctx); err != nil {
		closeFn()
		return nil, nil, err
	}

	return ctrSvc, closeFn, nil
}
