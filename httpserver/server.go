package httpserver

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/paveltessman/jao/httpserver/shared"
	"github.com/paveltessman/jao/settings"
)

var ErrHTTP = errors.New("http server error")

func Serve(ctx context.Context, cfg settings.Settings, deps shared.Deps) error {

	listener, err := net.Listen("tcp", cfg.HttpListenAddr)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrHTTP, err)
	}

	srv := &http.Server{
		Handler:           mainRoute(deps),
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       10 * time.Second,
	}

	errs := make(chan error, 1)

	go func() {
		log.Printf("httpserver: serving on %s", cfg.HttpListenAddr)
		errs <- srv.Serve(listener)
	}()

	select {
	case err := <-errs:
		return ignoreServerClosed(err)
	case <-ctx.Done():
		break
	}

	log.Printf("httpserver: shutting down")
	shutDownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutDownCtx); err != nil {
		return errors.Join(ErrHTTP, err, ignoreServerClosed(<-errs))
	}

	return ignoreServerClosed(<-errs)
}

func ignoreServerClosed(err error) error {
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}
