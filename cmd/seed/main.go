// Command seed applies the migrations and inserts the fixed v1 catalogue, the
// demo accounts and verifies the committed media files. It is idempotent.
package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/william22913/clover-be/internal/config"
	"github.com/william22913/clover-be/internal/db"
	"github.com/william22913/clover-be/internal/seed"
	"github.com/william22913/clover-be/migrations"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	cfg, err := config.Load()
	if err != nil {
		logger.Error("seed: configuration", "error", err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	pool, err := db.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.Error("seed: connect", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	migrateCtx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	if err := db.Migrate(migrateCtx, pool, migrations.FS); err != nil {
		cancel()
		logger.Error("seed: migrations", "error", err)
		os.Exit(1)
	}
	cancel()

	if err := seed.Run(ctx, pool, cfg.MediaDir); err != nil {
		logger.Error("seed: run", "error", err)
		os.Exit(1)
	}
	logger.Info("seed: complete", "films", len(seed.Films), "accounts", len(seed.Accounts))
}
