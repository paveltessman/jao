package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"text/tabwriter"

	"github.com/paveltessman/jao/keys"
	"github.com/paveltessman/jao/settings"
)

func runList(ctx context.Context, cfg settings.Settings, args []string) error {
	if len(args) > 0 {
		return errors.New("usage: jao list")
	}

	ctr, closeFn, err := newContainerService(ctx)
	if err != nil {
		return err
	}
	defer closeFn()

	installed, err := keys.New(ctr, cfg.PublicHost).Discover(ctx)
	if err != nil {
		return err
	}

	return printInstalled(os.Stdout, installed)
}

func printInstalled(w io.Writer, installed []keys.Installed) error {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintln(tw, "NAME\tPROTOCOL")
	for _, inst := range installed {
		_, _ = fmt.Fprintf(tw, "%s\t%s\n", inst.Name, inst.Protocol)
	}
	return tw.Flush()
}
