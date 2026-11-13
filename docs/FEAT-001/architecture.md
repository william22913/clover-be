---
feature_code: FEAT-001
service: clover
version: 1
status: complete
spec_version: 1
code_repo: https://github.com/william22913/clover-be.git
release_branch: main
integration_branch: feat/FEAT-001
stack: go
dev_feedback: []
code_prep_feedback: []
sources:
  - path: knowledge/design-reference.md
    source_version: unversioned
    note: The manager's Dribbble reference. Read and mapped by /ui-preparation (UIMap.md, DesignBrief.md); cited here, not re-analysed.
  - path: ui/mock
    source_version: v2 (PR #58, merged 785cc5f)
    note: The approved static mock is the implementation source for the served UI (markup, copy, states, assets). It replaces DESIGN.md, which still describes the withdrawn draft.
pr_url: https://github.com/william22913/sandbox-docs/pull/59
last_updated: 2026-11-13
---

# FEAT-001 — Architecture (clover)

## Overview

Clover is a small movie-streaming web app for one business owner. v1 is one
journey end to end: open the site, browse a public catalogue, pick a film, and —
because watching needs an account — sign in or create one, then land on that
film's player and watch it.

One new codebase, one new database, no third party:

- `clover-be` — a Go HTTP service built on the standard library only. It serves
  the JSON API, the static UI, the poster images and the video stream, and owns
  all persistence.
- PostgreSQL 16 — the only data store, three tables (`users`, `movies`,
  `sessions`).
- The UI is plain HTML/CSS/JS served by the same Go process. There is no
  React/Vite app and no second repository.

### Stack decision

`stack: go`. The feature is a request/response HTTP service with relational
data, a login gate and byte-range media serving — nothing scheduled, batched or
columnar, so Airflow does not apply. The approved `spec.md` fixes Go as a hard
constraint and forbids `nexcommon` and `nexcomponent`; the service is therefore
built from `net/http` and the Go standard library, with three small modules
whose need is stated in `## Technology Stack` (bcrypt, the PostgreSQL driver,
the Prometheus client). `/go-dev` implements this feature.

## Technology Stack

| Layer | Choice | Reason / constraint |
|---|---|---|
| Language | Go 1.23+ | REQ hand-off: Go, standard library only — no nexcommon, no nexcomponent, no web framework. |
| HTTP | `net/http` + `http.ServeMux` method/path patterns (`GET /v1/movies/{slug}`) | Go 1.22+ mux covers method routing and path parameters; no router dependency needed. |
| HTML render | none — static files served as-is | The UI is the approved mock's HTML/CSS/JS; no template engine. |
| Database | PostgreSQL 16 | Relational, transactional, real unique constraints. House storage engine; works with the existing compose/verify stack. |
| DB driver | `github.com/jackc/pgx/v5` (`pgxpool`) | Standard PostgreSQL driver; no ORM. |
| Migrations | embedded `.sql` files + a ~60-line runner (`embed.FS`, `schema_migrations` table) | Versioned, ordered, applied once; no migration library needed for one schema file. |
| Password hashing | `golang.org/x/crypto/bcrypt`, cost 12 | Named by the spec. Per-hash salt, work factor recorded here. |
| Session | opaque 32-byte random token (base64url), SHA-256 stored | Revocable server-side, as REQ-020 requires. |
| Metrics | `github.com/prometheus/client_golang` | House metric stack; exposes the standard text format. |
| Logging | `log/slog` (stdlib) | Structured, one line per request; no logging library. |
| UI assets | Static files copied from the approved mock (`ui/mock/**`) | The manager approved this look; the build must not re-invent it. |
| Container | Dockerfile (multi-stage, distroless/static base) + `docker-compose.yml` (app + postgres) | Nothing extra to run for video: the Go process serves it. |

## Repository layout

### `clover-be` (new — currently an empty repository)

```
cmd/clover/main.go        composition root: config, pool, migrate, seed, server, shutdown
internal/config/          env configuration (one struct, no library)
internal/store/           pgx pool, embedded migration runner, seed loader,
                          users.go / movies.go / sessions.go queries
internal/auth/            password rules + bcrypt, token issue/hash, session resolve, rate limiter
internal/api/             router, middleware (request id, log, session, CSRF, security headers,
                          metrics), error envelope, DTOs, handlers
internal/media/           session-gated video handler (http.ServeContent) and public static serving
internal/metrics/         collectors + the status-capturing middleware
migrations/0001_init.sql  users, movies, sessions
seed/movies.sql           the 13 v1 films, idempotent upsert by slug
web/                      index.html, app.css, app.js, assets/ (fonts, icons) — from the approved mock
media/poster/*.jpg        13 posters, 3 stills
media/video/*.mp4         the sample clip
Dockerfile, docker-compose.yml, README.md
```

Every package is `internal/`: there is exactly one binary and no library
consumer, so nothing needs to be importable from outside the module.

`clover-be` exists but is **empty** — created on 2026-11-13 with `main` as its
default branch and no commits. The first `/go-dev` session therefore
establishes `main` with its first commit and cuts `feat/FEAT-001` from it;
there is no earlier code, CI configuration or deployment to preserve, and no
consumer to keep compatible.

## Components

| Component | New / touched | Boundary and responsibility |
|---|---|---|
| `cmd/clover` | new | Reads config, opens the pool, applies migrations and the seed, builds the router and the metrics listener, serves until SIGTERM with a bounded graceful shutdown. Owns process wiring only. |
| `internal/config` | new | `DATABASE_URL`, `MIGRATE_DATABASE_URL` (optional; the migrator role's DSN), `LISTEN_ADDR` (default `:8080`), `METRICS_ADDR` (default `:9101`), `APP_ENV`, `PUBLIC_ORIGIN`. No secret has a default. Nothing prints the struct. |
| `internal/store` | new | Parameterised SQL only. Three query files, one per aggregate. The unique-email guarantee is the index, not a pre-check; a `23505` violation is mapped to `email_taken`. |
| `internal/auth` | new | Password rules and bcrypt, token generation and hashing, session resolution, the sign-in rate limiter. Pure logic plus the store it is given — no HTTP knowledge. |
| `internal/api` | new | All routes, the JSON envelope, every handler, and the middleware chain in one fixed order: request id → log → metrics → security headers → CSRF origin check → route → session (only on routes that need it). |
| `internal/media` | new | The only place that touches the filesystem for user-facing assets. Video: resolve `movies.video_file` by slug, open, `http.ServeContent`. Posters/static: `http.FileServer(http.Dir(...))`. |
| `internal/metrics` | new | Four collectors and the response-status capture; owns nothing else. |
| `web/` | new | The served UI: shell, client router (`/`, `/auth`, `/play/{slug}`), API calls, all states from `UIMap.md`. |
| `migrations/` + `seed/` | new | Schema and the fixed catalogue, embedded and idempotent. |
| Untouched | — | Greenfield: no existing service, contract, schema or consumer to preserve. |

No component other than `internal/auth` hashes or compares a password, and no
component other than `internal/media` reads a media path from a database row —
so the two decisions most likely to drift cannot drift between handlers.

## Backend — Endpoints

All JSON routes are mounted under `/v1/`. The UI's own pages (`/`, `/auth`,
`/play/{slug}`) and `/static/` are not part of the API; they are listed in
`## Frontend`. Unknown paths under `/v1/` return the JSON 404 envelope; any
other unknown `GET` returns the shell, and the client router falls back to the
catalogue (UIMap "One silent fallback"). An unknown method on a known path
returns 405 with `Allow`.

Authorization is one of two things, never implicit:

- **public** — no session is read. Catalogue reads, sign-up, sign-in.
- **session** — the `clover_session` cookie is resolved to a live row in
  `sessions` (`revoked_at IS NULL AND expires_at > now()`, joined to a
  non-deleted `users` row, in one statement against the database clock);
  otherwise `401 session_expired` before the handler runs.

`consumer: fe` means a screen listed in `## Frontend` calls it. There is no
`service:` consumer: Clover has no sibling service, and `spec.md`'s External
Integrations section is deliberately empty.

| Method | Path | Auth | Success | Consumer | Purpose |
|---|---|---|---|---|---|
| GET | `/v1/movies` | public | 200 `{items:[MovieSummary]}` | fe | The whole published catalogue in display order — no paging, no query parameters (REQ-001, REQ-002). |
| GET | `/v1/movies/{slug}` | public | 200 `MovieDetail`, 404 `movie_not_found` | fe | One film for the player page and the auth screen's film block (REQ-003, REQ-017). |
| GET | `/v1/movies/{slug}/video` | session | 200/206 `video/mp4`, 401 `session_expired`, 404 `movie_not_found`, 404 `video_unavailable` | fe | The stream. `http.ServeContent` answers `Range` with 206; `HEAD` is supported for the same handler (REQ-008, REQ-009, REQ-018, REQ-019, REQ-020). |
| POST | `/v1/auth/sign-up` | public | 201 `{user}` + `Set-Cookie`, 422 `validation_failed`, 409 `email_taken`, 429 `rate_limited` | fe | Create an account and start a session in one step (REQ-005, REQ-014, REQ-015, REQ-022, REQ-024). |
| POST | `/v1/auth/sign-in` | public | 200 `{user}` + `Set-Cookie`, 401 `invalid_credentials`, 422 `validation_failed`, 429 `rate_limited` | fe | Sign in; one generic failure for both unknown email and wrong password (REQ-006, REQ-013, REQ-023). |
| POST | `/v1/auth/sign-out` | public (cookie) | 204 + cleared cookie | fe | Revoke the session named by the cookie, if any (REQ-010, REQ-020). Idempotent: a missing or dead cookie is still 204. |
| GET | `/v1/auth/me` | session | 200 `{user}` | fe | Session probe on every page load and the player's session-expiry probe (REQ-004, REQ-012, REQ-019). |
| GET | `/healthz` | public | 200 `{status:"ok"}`, 503 on a failed DB ping | external | Container/orchestrator liveness probe. Not part of the API contract. |
| GET | `/metrics` | internal listener | 200 Prometheus text | external | Scraped by Prometheus. Bound to `METRICS_ADDR`, never registered on the public mux — see `## Observability`. |

Notes on the contract:

- `MovieSummary` = `{slug, title, short_description, release_year, runtime_seconds, poster_url, still_url}`.
  `MovieDetail` is the same object; the detail route exists so the player and
  the auth screen can load one film without fetching the catalogue. Whether the
  film can actually play is decided by the stream route, not by a field.
- `user` = `{email}`. The email is the only account field v1 renders. No
  internal `users.id` is ever serialized, so no public id is needed for it; the
  film's public id is its `slug`, and poster/video files are resolved
  server-side from that slug.
- There is no filter or search endpoint in v1, by spec. The rule that a list
  page's filter data is scoped to the consuming page's permission has nothing
  to apply to here: the catalogue has no filters, no facets and no secondary
  lookup. `GET /v1/movies` is itself the whole catalogue.
- Both the detail route and the stream route resolve the film with
  `slug = $1 AND is_published AND NOT is_deleted`, the same predicate as the
  catalogue index, so an unpublished or withdrawn film is `404` on every
  route — not just absent from the list.
- Request bodies are limited to 1 MiB (`http.MaxBytesReader`); a body over that
  is `413 invalid_request`.

## Frontend

**This feature has UI scope, but there is no separate frontend repository.**
The manager's constraint is that the UI is served by the same Go service as
static files, so `frontend_repo` and `frontend_base_branch` are deliberately
absent from the frontmatter. `/go-dev` builds the UI as part of this feature;
`/fe-dev` does not apply, and no handoff branch exists for it.

Source of truth: `ui/mock/` (merged PR #58) — `index.html`, `app.css`,
`app.js`, `assets/` — with `UIMap.md` for the screen contract and
`DesignBrief.md` for the design system. When the mock and `DesignBrief.md`
disagree the mock wins, per `DesignBrief.md` itself.

| Page | Route | Who | What it shows |
|---|---|---|---|
| Catalogue | `/` | anyone | Hero, facts rail, film rail, stat and step cards, and the poster grid — every published film with poster, title and short description (REQ-001, REQ-002). |
| Sign in / Create account | `/auth?next={slug}` | anyone | One screen, two modes. `next` carries the film the visitor wants; the film's still, title and description are shown when it is present. |
| Player | `/play/{slug}` | session (client gate, server-enforced stream) | The film with the five controls and one message panel for every failure. |

Client decisions:

- **Routing** is path-based, matching the mock's own route table: the server
  serves the same shell for `/`, `/auth` and `/play/{slug}`; `app.js` reads
  `location.pathname` and `location.search`. A route it does not recognise
  renders the catalogue without an error. (`UIMap.md` allows hash routes or
  server paths; paths are chosen so a pasted link, bookmark or refresh is a
  real URL, which REQ-012 and REQ-017 are written against.)
- **Session state** comes from one `GET /v1/auth/me` on boot; a 401 means
  anonymous. The header swaps between the "Sign in" capsule and the account
  disc from that single fact.
- **API access** is one small `api()` helper (same-origin `fetch`,
  `credentials: 'same-origin'`, JSON, error envelope → typed error). Every page
  uses it; no page builds a URL by hand.
- **Text is never injected as HTML.** Catalogue, film and error strings are
  rendered through the mock's `esc()` helper or `textContent`; the only
  `innerHTML` is for fixed markup with no server value in it.
- **States** are exactly the ones in `UIMap.md`: catalogue signed-out and
  signed-in; auth sign-in, sign-in error, sign-up, invalid input, duplicate
  email, expired session; player paused, playing, idle, video-failed,
  film-missing. The preview query flags in the mock (`?demo=1`, `err=bad`,
  `u=1`, …) are **not** ported — each real state now comes from a real
  condition.
- **Responsive**: the mock's own breakpoints; the volume slider hides under
  600 px, mute stays. No separate mobile design.
- **No build step**, no framework, no external request: fonts, images, video
  and icons are local files, so the page works on a first paint with the
  service's own CSP.

## Data Design

The full schema, ERD, normalization decisions, seed rows and migration batches
are in `database.md` — the confirmed output of the schema gate (draft ERD →
advisory normalization review → batched migration plan). It contains no DDL;
`/go-dev` writes the migrations in the batch order recorded there.

Summary of the decision:

- Three tables: `users`, `movies`, `sessions`.
- Email uniqueness is a **real unique index on `lower(email)`** (partial on
  `NOT is_deleted`), never an application pre-check — that is what makes
  REQ-022 case-insensitive uniqueness hold under concurrent sign-ups.
- `movies` stores **filenames**, not URLs (`poster_file`, `still_file`,
  `video_file`). The service resolves the path and the video route is keyed by
  `slug`, so no request ever carries a filesystem path.
- `sessions` holds only the **SHA-256 of the token**; the raw token exists only
  in the cookie. Sign-out sets `revoked_at`; expiry is absolute at 7 days.
- Soft delete is the `is_deleted`/`deleted_at`/`deleted_by` triple on `users`
  and `movies`. No `status` column in v1 carries a deletion state.

## Security design

| Threat | Control | BE / FE | Proof |
|---|---|---|---|
| Brute force and credential stuffing on sign-in | bcrypt cost 12; per-IP fixed-window limit of 10 sign-in/sign-up attempts per minute in `internal/auth`, keyed by the TCP peer address only (`X-Forwarded-For` is not read) | BE | Test: 11th attempt in a minute → `429 rate_limited`, no password comparison performed. |
| Account enumeration from the sign-in response | One byte-identical `401 invalid_credentials` body for unknown email and wrong password; when the email is unknown, bcrypt still runs against a fixed dummy hash so the two paths cost the same | BE | Test: both responses compared field by field; TC-024, TC-025. |
| Account enumeration from sign-up | Deliberate and required: REQ-014 makes "this email is taken" a user-facing message. Accepted exposure, limited by the same rate limit | BE | TC-026. |
| Session theft or replay | 256-bit token from `crypto/rand`, `HttpOnly`, `Secure`, `SameSite=Lax`, `Path=/`, 7-day `MaxAge` and `Expires`; only its SHA-256 is stored; one atomic query re-checks `revoked_at IS NULL AND expires_at > now()` against the **database** clock on every request, and the stream endpoint re-checks on every `Range` request | BE | TC-012, TC-015, TC-020; SC-009. |
| Session fixation / token reuse across sign-ins | A new token is always issued on sign-in or sign-up; a session presented during sign-in is revoked in the same transaction | BE | Test: two consecutive sign-ins with the first cookie produce different cookie values, and the first is 401 afterwards. |
| CSRF on the three mutating routes | `SameSite=Lax` cookies (cross-site POST carries no cookie) **plus** an explicit check on every `POST`: if `Sec-Fetch-Site` is present it must be `same-origin`; if `Origin` is present it must equal `PUBLIC_ORIGIN` (or the request host when unset). A mismatch is `403 invalid_request` | BE | Test: `POST /v1/auth/sign-out` with `Origin: https://evil.example` → 403 and the session is still live. |
| XSS through catalogue text or error messages | All server strings rendered through `esc()`/`textContent`; `Content-Type: application/json` + `X-Content-Type-Options: nosniff`; CSP `default-src 'self'; script-src 'self'; style-src 'self'; img-src 'self' data:; media-src 'self'; font-src 'self'; connect-src 'self'; object-src 'none'; base-uri 'none'; frame-ancestors 'none'; form-action 'self'` | BE + FE | FE unit test: a film title containing `<img src=x onerror=alert(1)>` renders as text; response-header test asserts the CSP. |
| Clickjacking | `frame-ancestors 'none'` in the CSP and `X-Frame-Options: DENY` | BE | Header test. |
| Path traversal through media | The video path never comes from the URL: `{slug}` → row → `video_file` filename. Posters and static files are served by `http.FileServer(http.Dir(...))`, whose own path cleaning refuses `..` above the root; `Content-Disposition` is not user-controlled | BE | Test: `/media/poster/../../etc/passwd` → 404; `/static/../../etc/passwd` → 404. |
| Watching without an account | The stream route is behind the session middleware; there is no public copy of the file (no object storage, no second listener) | BE | TC-006, TC-027; SC-007. |
| Session expiring mid-film | Every range request re-resolves the session; a 401 on the media element makes the player probe `GET /v1/auth/me` and route to `/auth?next={slug}` | BE + FE | TC-031. |
| Open redirect through `next` | `next` must match `^[a-z0-9-]{1,64}$` **and** resolve to a published film; anything else is replaced by `/` | BE + FE | Test: `/auth?next=https%3A%2F%2Fevil.example` renders the catalogue on success. |
| Password storage | bcrypt only; the hash never leaves `internal/auth` and is not part of any DTO; no log line carries a password or a hash | BE | Test: no field in any response contains the hash; log capture on sign-up contains neither. |
| PII in logs and metrics | Log lines carry request id, method, route template, status, duration and user id only — never an email, a password or a token; metric labels are route templates and fixed enums | BE | Log-review test around a full sign-up/sign-in cycle. |
| Secrets in the repository | Configuration from the environment only; no credential, token or DSN defaulted in code; `.env` and local compose values are development-only placeholders | BE | Review gate: no secret-shaped literal in the tree. |
| Over-privileged database credential | Two roles: `clover_migrator` owns the schema, `clover_app` is the runtime login with `SELECT/INSERT/UPDATE` and no DDL; `PUBLIC` loses `CREATE` on the schema | BE | Migration test: `clover_app` cannot `CREATE TABLE`. |
| Unauthenticated metric scraping | `/metrics` is not on the public mux at all; it is served by a second `http.Server` on `METRICS_ADDR`, unpublished in compose | BE | Router test: `GET /metrics` on the public port → 404 envelope. |
| Oversized or malformed request bodies | `http.MaxBytesReader` at 1 MiB, `DisallowUnknownFields`, `Content-Type: application/json` enforced | BE | Test: 2 MiB body → 413; `text/plain` → 422. |

## Error contract

Every response carries `X-Request-Id`; the same value is in the JSON body and
in the server's log line, so any user-visible failure can be traced.

Envelope:

```json
{"error":{"code":"invalid_credentials","message":"That email and password do not match an account.","request_id":"…","fields":{"email":"…"}}}
```

`code` is the stable contract; `message` is English and safe to display;
`fields` is present only on validation failures and maps a field name to one
message. Success bodies are the plain object described per endpoint (no
envelope), so the FE reads `movie` / `items` / `user` directly.

| HTTP | `code` | When | FE message and behaviour |
|---|---|---|---|
| 400 | `invalid_request` | Malformed JSON, wrong `Content-Type`, unknown field, or a body over 1 MiB | "Something went wrong. Please try again." + request id. POSTs only. |
| 401 | `invalid_credentials` | Unknown email or wrong password at sign-in — identical body for both | Notice above the form: "That email and password do not match an account." Email kept, password cleared, stay on `/auth` (REQ-013). |
| 401 | `session_expired` | No cookie, unknown token, expired or revoked session on a session route | Player: go to `/auth?next={slug}&reason=expired`, gold "Your session ended. Sign in again to keep watching." After sign-in, back to the film (REQ-019). Catalogue: swap the header to "Sign in" and stay put. |
| 403 | `invalid_request` | CSRF origin/`Sec-Fetch-Site` mismatch on a POST | Generic "Something went wrong." + request id; no retry loop. |
| 404 | `movie_not_found` | `{slug}` not in the published catalogue | Player shows "This film is not available." / "It may have left the catalogue, or the link may be wrong." + "Back to the catalogue" (REQ-017). |
| 404 | `video_unavailable` | The movie row exists but its file is missing or unreadable | Player shows "The film could not be loaded." / "Check your connection, then try again." + "Try again" + "Back to the catalogue" (REQ-018). |
| 404 | `not_found` | Unknown path or method under `/v1/` with no more specific code | Generic message + request id; a page-level 404 renders the catalogue. |
| 405 | `method_not_allowed` | Known path, wrong method; `Allow` header set | Not user-reachable from the UI. |
| 409 | `email_taken` | Sign-up with an email that already exists, case-insensitively | Inline on the email field: "An account already uses this email. Sign in instead." with the link switching to the sign-in tab (REQ-014). |
| 422 | `validation_failed` | Field rules: email empty/invalid/too long, password empty/too short/too long/no letter/no number | Inline message under each invalid field, first invalid field focused, everything typed kept (REQ-015). |
| 429 | `rate_limited` | More than 10 sign-in/sign-up attempts per minute from one IP | Notice: "Too many attempts. Try again in a minute." |
| 500 | `internal_error` | Any unexpected failure; the cause is logged with the request id and never returned | "Something went wrong. Please try again." + request id. |

A code that is not in this table is shown only as the 500 row, with the request
id; the FE never renders a raw `message` it does not recognise. The FE also
treats a network failure with no response as `video_unavailable` on the player
(REQ-018) and as a retryable notice everywhere else.

## Observability

**Current state.** `clover-be` is an empty repository — no instrumentation, no
`/metrics`, no HTTP middleware, and by constraint no `nexcommon` (so there is
no library middleware whose status-label behaviour needs checking; the
hardcoded-200 trap recorded in the guard does not apply here).

**Decision: wire observability once, in this feature.** `internal/metrics`
registers the collectors on the default registry and exposes
`promhttp.Handler()` from a second `http.Server` bound to `METRICS_ADDR`
(default `:9101`). The public mux has no `/metrics` route, and the compose file
does not publish `METRICS_ADDR` — it is reachable only inside the compose
network or by an operator on the host.

**Implementation boundary.**

- `internal/api` composes one middleware that wraps the whole public handler. It
  captures the status actually written (an implicit 200 on first write, plus
  the optional `Flusher`/`ReaderFrom` interfaces preserved so range streaming
  is not broken), and labels by the **route template** from the mux
  (`r.Pattern` → `/v1/movies/{slug}/video`), never the raw path, so the
  `{slug}` value cannot explode label cardinality.
- An unmatched route uses the single label `unmatched`; `/healthz` is
  instrumented, `/metrics` is not.
- Handlers report business errors through one helper, so the error counter is
  incremented with the same `code` the response carries.
- Metrics: `clover_http_requests_total{method,route,status}`,
  `clover_http_request_duration_seconds{method,route,status}` (histogram),
  `clover_http_errors_total{code}`, and three business counters the spec makes
  worth watching — `clover_signins_total{result="ok|failed|rate_limited"}`,
  `clover_signups_total{result="ok|duplicate|invalid"}`, and
  `clover_sessions_revoked_total`. No dashboards are invented here beyond these.
- Logging: one line per request via `log/slog` with request id, method, route
  template, status, duration and (when signed in) user id. Unexpected errors
  log the cause at `Error` with the request id.

**Verification.** A router test scrapes the metrics listener after one 200
(`GET /v1/movies`) and one 401 (`GET /v1/movies/halcyon/video` with no cookie)
and asserts two distinct `status` labels and the template route label; a
second test asserts `GET /metrics` on the public listener returns the 404
envelope. The `## Security design` row for metric scraping is proved by the
same test.

## Implementation Plan

- PART-1: Service foundation — Go module, `internal/config`, pgx pool, embedded migration runner, `log/slog`, request id + logging + metrics + security-header + CSRF middleware, error envelope, `/healthz`, the internal metrics listener, `Dockerfile` and a compose stack (app + postgres), README.
- PART-2: Data layer — `0001_init.sql` (`users`, `movies`, `sessions`) with the two database roles and grants, the idempotent `seed/movies.sql` for the 13 films, the media assets under `media/`, and store queries for all three aggregates with their tests.
- PART-3: Identity — password rules and bcrypt, sign-up / sign-in / sign-out / me, session issue and resolution, the rate limiter, and the CSRF check on the three POST routes.
- PART-4: Catalogue and streaming — `GET /v1/movies`, `GET /v1/movies/{slug}`, the session-gated `GET /v1/movies/{slug}/video` over `http.ServeContent`, poster and static serving, and the faststart check on the video asset.
- PART-5: UI shell — the three served routes, the client router, the API helper, the session boot, the catalogue page (hero, rails, cards) wired to `GET /v1/movies`, and the signed-in header.
- PART-6: UI auth and player — the combined sign-in / create-account screen with every state, the player with the five controls and the three failure states, the expired-session recovery path, and the end-to-end checks in the compose stack.

Parts 5 and 6 are the UI half. `/go-dev` implements them in this repository —
there is no `/fe-dev` handoff for this feature.

## Spec Traceability

- **REQ-001** — `GET /v1/movies` is public (no session middleware on the route) and the catalogue page renders from it on first load; the shell at `/` is served without any account check. TC-001.
- **REQ-002** — One query returns every `is_published AND NOT is_deleted` row ordered by `display_order, id`; the endpoint accepts no paging, search or sort parameters, and unknown query strings are ignored rather than interpreted. TC-002, TC-003.
- **REQ-003** — Selecting a film navigates to `/play/{slug}`; the server serves the shell for that route and the page loads `GET /v1/movies/{slug}`. TC-004.
- **REQ-004** — A signed-out selection goes to `/auth?next={slug}`; a successful sign-in or sign-up sets the session cookie and the client navigates to `/play/{slug}` — never to `/`. `next` survives both modes. TC-006, TC-007, TC-010.
- **REQ-005** — `POST /v1/auth/sign-up` creates the row and a session and answers 201 with the cookie; no verification step exists anywhere in the flow. TC-007, TC-008.
- **REQ-006** — `POST /v1/auth/sign-in` looks the account up with `WHERE lower(email) = lower($1)` and compares the submitted password with bcrypt. TC-009, TC-010.
- **REQ-007** — The cookie is persistent (`MaxAge` and `Expires` set to 7 days) and the session row's `expires_at` is creation + 7 days; nothing is stored in browser storage. TC-011, TC-012.
- **REQ-008** — FE only: `web/app.js` implements play, pause, the scrubber, volume/mute and fullscreen over one `<video>` element with native controls off; the server's job is to deliver a stream the element can drive. TC-016 to TC-019.
- **REQ-009** — `http.ServeContent` answers `Range` with `206 Partial Content` and `Accept-Ranges: bytes`; the MP4 is H.264/AAC with the `moov` atom before `mdat` (faststart), so playback starts before the file is complete and seeking is a new range request. TC-020 to TC-022.
- **REQ-010** — `POST /v1/auth/sign-out` sets `revoked_at` on the session and clears the cookie; the FE returns to the catalogue with anonymous state. TC-014.
- **REQ-011** — Every string in `web/` and every error `message` is English; there is no i18n layer and no locale selection in v1. TC-005.
- **REQ-012** — The shell for `/play/{slug}` is served to anyone, and the page's boot `GET /v1/auth/me` returning 401 routes to `/auth?next={slug}`; the stream itself is unreachable without a session regardless of what the client does. TC-023.
- **REQ-013** — Both failure paths return the same `401 invalid_credentials` body; the FE renders one notice and keeps the email. TC-024, TC-025.
- **REQ-014** — The unique index on `lower(email)` rejects the second row; the handler maps `23505` to `409 email_taken`, issues no session, and the FE shows the linked field error. TC-026.
- **REQ-015** — Field rules are enforced in the request validator and reported as `422 validation_failed` with a `fields` map; the FE renders each message under its field and never clears an input. TC-032, TC-034, TC-035, TC-037, TC-038, TC-045.
- **REQ-016** — No cookie is issued unless sign-up or sign-in succeeds, and the stream route is session-gated, so abandoning `/auth` leaves the visitor anonymous with the film locked. TC-027.
- **REQ-017** — An unknown `{slug}` is `404 movie_not_found` from both the detail and the stream route; the player renders the "This film is not available." panel with a route back to the catalogue. TC-028.
- **REQ-018** — A movie whose file is missing or unreadable is `404 video_unavailable`; any media-element error triggers the "The film could not be loaded." panel with "Try again", and there is no spinner state that outlives a failed load. TC-029, TC-030.
- **REQ-019** — Every range request re-resolves the session; a 401 makes the player probe `GET /v1/auth/me`, then route to `/auth?next={slug}&reason=expired`, and the post-sign-in redirect returns to the same film. TC-031.
- **REQ-020** — Sign-out writes `revoked_at`; the session middleware and the stream route both require `revoked_at IS NULL`, so a copied cookie is refused immediately. TC-015; SC-009.
- **REQ-021** — Every sign-in creates a new `sessions` row and revokes only the session presented in that request, so other devices keep their rows and stay signed in. TC-013.
- **REQ-022** — Required and syntactically valid email is a request validator rule (non-empty, ≤254 bytes, `local@domain.tld` shape); surrounding whitespace is trimmed once at that boundary so two spellings cannot become two accounts; case-insensitive uniqueness is the partial unique index on `lower(email)`, not a pre-check. TC-032, TC-033, TC-040, TC-041.
- **REQ-023** — Sign-in resolves the account with `lower(email) = lower($1)` and passes the submitted password bytes to `bcrypt.CompareHashAndPassword` with no trim and no case folding; the same function is used for sign-in only, never for changing a stored hash. TC-041 to TC-045.
- **REQ-024** — Sign-up-only validator: at least 8 bytes (and at most 72, bcrypt's input limit, so no password is silently truncated), at least one Unicode letter, at least one digit, on the exact submitted bytes. No matching rule runs at sign-in. TC-034 to TC-039, TC-044.

## Planned breaking changes

None. `clover-be` is empty and this is its first feature; no existing consumer,
endpoint or payload changes.

## Deferred / explicitly out of scope

- Recommendations, subscriptions/payments, search, filtering and sorting — no
  endpoint, column or index exists for them.
- Any screen or route that writes movie data; the catalogue is seed-only.
- Email verification, password reset and account recovery — no token table, no
  mail dependency.
- Profiles, watchlists, viewing history, resume position, ratings, comments —
  `sessions` stores no viewing state.
- Non-English UI, native apps.
- Sliding session renewal: the 7-day lifetime is absolute, as TC-012 tests.
- Account deletion: `users.is_deleted` exists for the house convention, but no
  v1 route sets it.
