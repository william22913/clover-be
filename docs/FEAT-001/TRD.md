---
feature_code: FEAT-001
service: clover
locked_version: 1
last_updated: 2026-11-13
---

# FEAT-001 — Technical Requirements Document

> Overview summary only — see `architecture.md` for the full design and its
> rationale, and `database.md` for the schema.

## Overview

FEAT-001 delivers Clover v1 as a single Go HTTP service (`clover-be`) with no
shared house library: `net/http` and the standard library for the server,
PostgreSQL 16 for storage, and the front end — static HTML, CSS and JavaScript
built from the approved mock — served by that same process from `web/`.
Visitors browse a public catalogue of thirteen seeded films, choose one, and
are sent through a combined sign-in / create-account screen that returns them
straight to that film. Playback streams the film from the same service over
`http.ServeContent`, so range requests and seeking work progressively.
Accounts are self-service email and password, sessions last seven days and are
revoked on sign-out, and every failure the visitor can hit — missing film,
broken video, expired session — has a named screen and a way back.

The repository is greenfield: `clover-be` is empty, so this feature creates its
`main` and its `feat/FEAT-001` integration branch. There is no separate
frontend repository.

## Technical Goals

- **One service, no front-end split.** The Go binary serves the `/v1/` JSON
  the UI calls, plus the static shell, posters and video. One deployable, one
  origin, no CORS layer.
- **Standard library first.** `http.ServeMux` method/path patterns,
  `http.ServeContent` for Range, `http.MaxBytesReader`, `log/slog`, `embed.FS`
  for migrations. The only third-party modules are bcrypt, the PostgreSQL
  driver and the Prometheus client, each named in `## Technology Stack` with
  the reason it is needed.
- **The database enforces the rules that can be broken concurrently.**
  Case-insensitive email uniqueness is a partial unique index on
  `lower(email) WHERE NOT is_deleted`; a duplicate sign-up is a `23505`, not a
  pre-check that races.
- **Opaque sessions, not JWTs.** A 32-byte random token in a `HttpOnly`,
  `Secure`, `SameSite=Lax` cookie; only its SHA-256 hash is stored, so the
  database never holds a usable credential and sign-out is a single `UPDATE`
  of `revoked_at`.
- **Playback is a file streamed by the standard library.** `http.ServeContent`
  answers `Range` with `206` and `Accept-Ranges: bytes`; the MP4 is
  faststart-encoded so playback and seeking need no transcoding, no HLS and no
  object storage.
- **Every error is a stable code.** One envelope,
  `{"error":{"code","message","request_id","fields"}}`, with codes the UI maps
  to a specific screen, never a raw server message.
- **Observable by default.** Prometheus metrics on a separate internal
  listener with real status codes and route-template labels; request ids in the
  body, the `X-Request-Id` header and the log line; no email, password or token
  in a log line or a metric label.

## Technology Stack

| Area | Choice | Purpose and reason |
|---|---|---|
| Language | Go 1.23+, module `github.com/william22913/clover-be` | Approved backend; 1.22+ mux patterns carry method and path template in one place. |
| HTTP | stdlib `net/http` with `http.ServeMux` | Manager constraint: no nexcommon and no nexcomponent. |
| Routing | `GET /v1/movies`, `GET /v1/movies/{slug}/video`, … | Methods and path variables without a third-party router. |
| Database | PostgreSQL 16 | Partial indexes, `timestamptz`, check constraints and two roles do work a file database cannot do safely. |
| Driver | `github.com/jackc/pgx/v5` (`pgxpool`) | The maintained driver for PostgreSQL 16. |
| Migrations | Embedded numbered SQL, one batch `0001_init.sql`, run at startup under an advisory lock | No migration dependency; the schema is small and new. |
| Password hashing | `golang.org/x/crypto/bcrypt`, cost 12 | Slow by design, no salt management, 72-byte input limit handled in validation. |
| Sessions | 32-byte random token, base64url in the cookie; SHA-256 hex in `sessions.token_hash`; 7-day absolute expiry | Revocation without server-side state beyond one indexed lookup. |
| Cookie | `clover_session`, `HttpOnly`, `Secure`, `SameSite=Lax`, `Path=/`, 7-day `MaxAge` | Not readable from script; cross-site POSTs blocked. |
| CSRF | `SameSite=Lax` plus an `Origin` / `Sec-Fetch-Site` check on every POST | Lax alone leaves a gap on some clients; the explicit check closes it. |
| Streaming | `http.ServeContent` over `os.File` | Range, `206`, conditional requests and `Content-Length` come for free. |
| Static | `http.FileServer(http.Dir(...))` for `web/` and `media/poster/` | Its path cleaning refuses traversal above the root. |
| Metrics | `github.com/prometheus/client_golang`, separate listener on `METRICS_ADDR` (default `:9101`) | Real status labels from a middleware that captures the written status. |
| Logging | `log/slog`, one line per request | Stdlib; structured, no PII. |
| Front end | Static HTML/CSS/JS from the approved mock under `web/` | Manager constraint; no React, no build step, no separate repo. |
| Packaging | Multi-stage `Dockerfile`; compose runs app + postgres | One command to review. |

## Key Technical Decisions

- **Path routes, not hash routes.** `/`, `/auth?next={slug}` and `/play/{slug}`
  are real paths the Go server answers with the same shell; the client reads
  `location.pathname` and renders the matching screen. The mock's hash-style
  ids map onto them one to one.
- **The catalogue is the only public data.** `GET /v1/movies` and
  `GET /v1/movies/{slug}` are open; the stream, `/v1/auth/me` and sign-out
  require a session. An unpublished or withdrawn film is a `404` on every
  route, using the same predicate as the catalogue index.
- **Three tables, no more.** `users`, `movies`, `sessions`. No viewing history,
  no watchlist, no tokens for password reset — those are out of scope and get
  no column.
- **Soft delete is three columns, never a status value.** `is_deleted`,
  `deleted_at`, `deleted_by` on `users` and `movies`; every uniqueness and
  lookup carries `WHERE NOT is_deleted`.
- **The session lookup is one statement against the database clock.** Session,
  user and expiry are resolved in a single query, so an expired or revoked
  session cannot pass between two reads and the app server's clock never
  decides expiry.
- **Validation at the boundary, hashes only in the database.** Email shape,
  length and password rules are checked in the request validator; password
  rules cannot be a CHECK constraint because only the bcrypt hash is stored.
- **Media is referenced by filename, resolved server-side.** The URL carries a
  slug; the row carries `poster_file`, `still_file` and `video_file`, each
  validated by a CHECK that pins the shape and forbids `..`.
- **Rate limiting is in-process.** Ten sign-in/sign-up attempts per minute per
  IP; a single replica in v1, so an in-memory limiter is sufficient and is
  recorded as the ceiling.
- **No account deletion in v1.** `users.is_deleted` exists for the house
  convention; no route sets it.

## Non-Functional Requirements

- **Security:** `architecture.md`'s `## Security design` carries fifteen threat
  rows, each with a control and a proof. The SQL security validator runs on the
  migration and the store queries before the feature is considered done.
- **Accessibility and motion:** the UI follows `UIMap.md` and the approved mock
  — keyboard access, visible focus, reduced-motion fallbacks, and accessible
  custom player controls over a single `<video>` element.
- **Performance:** the catalogue is one indexed read of thirteen rows; a poster
  is a static file; playback is a ranged read from disk. No caching layer is
  introduced for a catalogue this size.
- **Operability:** `/healthz` on the public listener, `/metrics` internal only,
  request id on every response, structured logs without PII, one container plus
  PostgreSQL.
- **Data integrity:** uniqueness, file-name shape, email shape, hash shape and
  soft-delete consistency are database constraints; the application never
  relies on a read-then-write check for them.
