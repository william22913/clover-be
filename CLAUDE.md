# clover-be

The Go API for Clover, a small movie-streaming web app. It serves the JSON
API under `/v1`, poster images and the video stream, plus `/healthz`,
`/readyz` and an internal `/metrics` listener. It does not serve or render the
UI — the browser app lives in the separate `clover-fe` repository and
reverse-proxies `/v1/` here.

## Stack and conventions

Decided in `docs/feat-001/architecture.md` (feat-001, spec v1) and followed by
every change:

- Go 1.23, stdlib `net/http` with `http.ServeMux` method/path patterns. No web
  framework.
- No `nexcommon` and no `nexcomponent` — a hard constraint from the product
  owner, not an omission.
- PostgreSQL 16 through `github.com/jackc/pgx/v5/pgxpool`; parameterized SQL
  only, no ORM.
- Migrations: embedded SQL files under `migrations/`, applied by a small
  runner that records versions in `schema_migrations` (one transaction per
  file, advisory lock at startup).
- Passwords: argon2id via `golang.org/x/crypto/argon2`, PHC-encoded.
- Sessions: opaque 32-byte token in an HttpOnly cookie; only its SHA-256 is
  stored in `sessions`.
- Metrics: `prometheus/client_golang` on a separate internal `:9090`
  listener; logging with stdlib `log/slog` (JSON, one line per request with
  the request id).

## Layout

- `cmd/server` — entrypoint (config, pool, router, listeners, shutdown)
- `cmd/seed` — idempotent seed of the catalogue, accounts and media
- `internal/httpapi` — router, middlewares, handlers, error envelope
- `internal/accounts`, `internal/sessions`, `internal/movies`,
  `internal/media` — domain code
- `internal/config`, `internal/observability`, `internal/i18n`, `internal/db`
- `migrations/` — embedded SQL; `media/posters/`, `media/videos/` — seeded
  assets

## Feature docs

- `feat-001` — Clover v1 movie streaming: public catalogue, session-gated
  playback with video Range support, seeded accounts and films. See
  `docs/feat-001/spec.md` (requirements), `docs/feat-001/architecture.md`
  (design authority), `docs/feat-001/database.md` (schema),
  `docs/feat-001/testcase.md` (63 acceptance cases).
- `docs/ERD.md` — the service's cumulative entity-relationship diagram.
