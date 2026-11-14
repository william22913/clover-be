// Package db owns the Postgres connection pool and the embedded migration
// runner. It holds no domain queries — those live in the domain packages.
package db

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// migrationLockKey is the session-level advisory lock the runner takes so two
// replicas starting at once cannot apply the same file twice.
const migrationLockKey int64 = 710047001

// Connect opens a pool and verifies the database answers.
func Connect(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		// pgx quotes the connection string in parse errors, which would put
		// the password in the log; report the failure without it.
		return nil, errors.New("db: DATABASE_URL is not a valid connection string")
	}
	// Bound every statement so a stalled query cannot pin a pooled
	// connection and its handler goroutine indefinitely.
	if cfg.ConnConfig.RuntimeParams == nil {
		cfg.ConnConfig.RuntimeParams = map[string]string{}
	}
	cfg.ConnConfig.RuntimeParams["statement_timeout"] = "15000"

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("db: create pool: %w", err)
	}
	pingCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("db: ping: %w", err)
	}
	return pool, nil
}

// Migrate applies every pending migration from fsys in filename order, one
// transaction per file, under an advisory lock held for the whole run.
func Migrate(ctx context.Context, pool *pgxpool.Pool, fsys fs.FS) error {
	conn, err := pool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("db: acquire migration connection: %w", err)
	}
	defer conn.Release()

	if _, err := conn.Exec(ctx, "SELECT pg_advisory_lock($1)", migrationLockKey); err != nil {
		return fmt.Errorf("db: take migration lock: %w", err)
	}
	defer func() {
		_, _ = conn.Exec(context.WithoutCancel(ctx), "SELECT pg_advisory_unlock($1)", migrationLockKey)
	}()

	if _, err := conn.Exec(ctx, `CREATE TABLE IF NOT EXISTS schema_migrations (
		version bigint PRIMARY KEY,
		applied_at timestamptz NOT NULL DEFAULT now()
	)`); err != nil {
		return fmt.Errorf("db: create schema_migrations: %w", err)
	}

	applied := map[int64]bool{}
	rows, err := conn.Query(ctx, "SELECT version FROM schema_migrations")
	if err != nil {
		return fmt.Errorf("db: read applied versions: %w", err)
	}
	for rows.Next() {
		var v int64
		if err := rows.Scan(&v); err != nil {
			rows.Close()
			return fmt.Errorf("db: scan applied version: %w", err)
		}
		applied[v] = true
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return fmt.Errorf("db: read applied versions: %w", err)
	}

	entries, err := fs.Glob(fsys, "*.sql")
	if err != nil {
		return fmt.Errorf("db: list migrations: %w", err)
	}
	sort.Strings(entries)

	for _, name := range entries {
		version, err := migrationVersion(name)
		if err != nil {
			return err
		}
		if applied[version] {
			continue
		}
		body, err := fs.ReadFile(fsys, name)
		if err != nil {
			return fmt.Errorf("db: read migration %s: %w", name, err)
		}
		if err := applyOne(ctx, conn.Conn(), name, version, string(body)); err != nil {
			return err
		}
	}
	return nil
}

func applyOne(ctx context.Context, conn *pgx.Conn, name string, version int64, body string) error {
	tx, err := conn.Begin(ctx)
	if err != nil {
		return fmt.Errorf("db: begin migration %s: %w", name, err)
	}
	defer func() { _ = tx.Rollback(context.WithoutCancel(ctx)) }()

	if _, err := tx.Exec(ctx, body); err != nil {
		return fmt.Errorf("db: apply migration %s: %w", name, err)
	}
	if _, err := tx.Exec(ctx, "INSERT INTO schema_migrations (version) VALUES ($1)", version); err != nil {
		return fmt.Errorf("db: record migration %s: %w", name, err)
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("db: commit migration %s: %w", name, err)
	}
	return nil
}

func migrationVersion(name string) (int64, error) {
	base, _, _ := strings.Cut(name, "_")
	v, err := strconv.ParseInt(base, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("db: migration %q has no numeric version prefix", name)
	}
	return v, nil
}
