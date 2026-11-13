# clover-be

Clover is a movie-streaming web app. This repository is the single service that
serves the JSON API under `/v1/`, the static UI, the posters and the
session-gated video stream, and owns all persistence.

## Stack

- Go 1.23+, standard library only: `net/http` with `http.ServeMux` method/path
  patterns, `http.ServeContent` for Range requests, `log/slog`, `embed.FS` for
  migrations. **No nexcommon, no nexcomponent, no web framework** — this is a
  stated product constraint.
- PostgreSQL 16 via `github.com/jackc/pgx/v5` (`pgxpool`), embedded SQL
  migrations, no ORM and no migration library.
- Passwords: `golang.org/x/crypto/bcrypt`, cost 12.
- Metrics: `github.com/prometheus/client_golang` on a separate internal
  listener (`METRICS_ADDR`, default `:9101`).
- UI: plain HTML/CSS/JS served from `web/`, no build step.

## Layout

Every package is `internal/` — there is one binary and no library consumer.

```
cmd/clover/main.go        composition root: config, pool, migrate, seed, server, shutdown
internal/config/          env configuration
internal/store/           pgx pool, embedded migration runner, seed loader, queries
internal/auth/            password rules + bcrypt, tokens, session resolution, rate limiter
internal/api/             router, middleware, error envelope, DTOs, handlers
internal/media/           session-gated video handler and static serving
internal/metrics/         collectors and the status-capturing middleware
migrations/               numbered .sql files
seed/                     idempotent catalogue seed
web/                      served UI
media/                    posters, stills and video
```

Conventions this project follows: parameterised SQL only; a real unique index,
never an application pre-check, for uniqueness; soft delete as the
`is_deleted`/`deleted_at`/`deleted_by` triple; the error envelope
`{"error":{"code","message","request_id","fields"}}` with a stable `code`;
English-only UI.

## Feature docs

- `FEAT-001` — movie streaming v1: public catalogue, combined sign-in /
  create-account, and the player. See `docs/FEAT-001/spec.md`,
  `docs/FEAT-001/architecture.md`, `docs/FEAT-001/database.md` and
  `docs/FEAT-001/testcase.md`.
- `docs/ERD.md` — the service's cumulative ERD.
