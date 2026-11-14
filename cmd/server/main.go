// Command server runs the Clover API: JSON endpoints under /v1, poster images,
// the session-gated video stream, the operational probes and an internal
// metrics listener.
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/william22913/clover-be/internal/config"
	"github.com/william22913/clover-be/internal/db"
	"github.com/william22913/clover-be/internal/httpapi"
	"github.com/william22913/clover-be/internal/observability"
	"github.com/william22913/clover-be/migrations"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	cfg, err := config.Load()
	if err != nil {
		logger.Error("server: configuration", "error", err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	pool, err := db.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.Error("server: connect", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	migrateCtx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	err = db.Migrate(migrateCtx, pool, migrations.FS)
	cancel()
	if err != nil {
		logger.Error("server: migrations", "error", err)
		os.Exit(1)
	}

	metrics := observability.New()
	api := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           httpapi.New(httpapi.Deps{Pool: pool, Config: cfg, Metrics: metrics, Logger: logger}).Handler(),
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      5 * time.Minute, // a full video stream is one response
		IdleTimeout:       2 * time.Minute,
	}
	internal := &http.Server{
		Addr:              cfg.MetricsAddr,
		Handler:           metrics.Handler(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       2 * time.Minute,
	}

	errs := make(chan error, 2)
	go func() {
		logger.Info("api listening", "addr", cfg.HTTPAddr)
		if err := api.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errs <- err
		}
	}()
	go func() {
		logger.Info("metrics listening", "addr", cfg.MetricsAddr)
		if err := internal.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errs <- err
		}
	}()

	exitCode := 0
	select {
	case <-ctx.Done():
		logger.Info("shutdown requested")
	case err := <-errs:
		logger.Error("listener stopped", "error", err)
		// A listener that dies is a crash: the supervisor must see it.
		exitCode = 1
	}

	shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancelShutdown()
	_ = api.Shutdown(shutdownCtx)
	_ = internal.Shutdown(shutdownCtx)
	logger.Info("server stopped")
	if exitCode != 0 {
		os.Exit(exitCode)
	}
}
