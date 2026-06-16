package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"
)

func newServer(addr string, handler http.Handler) *http.Server {
	return &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
	}
}

func (a *App) Run(ctx context.Context) error {
	errCh := make(chan error, 1)

	go a.serve(a.server, errCh)

	select {
	case <-ctx.Done():
		return a.shutdownWithTimeout(context.Background())
	case err := <-errCh:
		_ = a.shutdownWithTimeout(context.Background())
		return err
	}
}

func (a *App) Shutdown(ctx context.Context) error {
	if err := a.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("shutdown server: %w", err)
	}
	return nil
}

func (a *App) shutdownWithTimeout(ctx context.Context) error {
	shutdownCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return a.Shutdown(shutdownCtx)
}

func (a *App) serve(srv *http.Server, errCh chan<- error) {
	a.logger.Printf("server listening on %s", srv.Addr)

	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		errCh <- fmt.Errorf("server failed: %w", err)
	}
}
