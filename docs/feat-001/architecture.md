---
feature_code: feat-001
service: clover
version: 1
status: complete
spec_version: 1
code_repo: https://github.com/william22913/clover-be
release_branch: main
integration_branch: feat/feat-001
stack: go
frontend_repo: https://github.com/william22913/clover-fe
frontend_base_branch: main
dev_feedback: []
code_prep_feedback: []
sources:
  - path: knowledge/design-reference.md
    source_version: unversioned
    note: Manager's Dribbble reference; the approved mock (ui/mock/) and UIMap.md are the visual truth this design hands to /fe-dev.
pr_url: https://github.com/william22913/sandbox-docs/pull/71
last_updated: 2026-11-14
---

# feat-001 — Movie Streaming (Clover v1) — architecture

Design authority for `/go-dev` (backend) and `/fe-dev` (frontend). Requirements
live in `spec.md` v1; schema detail in `database.md`; screen detail in
`UIMap.md` and the clickable mock `ui/mock/`. This document does not restate
either — it decides how the two applications are built.

## Overview

Clover v1 is **two applications**, not one:

- **`clover-be`** — one Go service that serves only the JSON API, the poster
  images and the video stream, plus `/healthz`, `/readyz` and `/metrics`. It
  never renders or serves the UI.
- **`clover-fe`** — a React + Vite single-page app (no `nexcomponent`) that
  renders every screen from `UIMap.md` and calls the API. Its nginx serves the
  built SPA and reverse-proxies `/v1/*` to the backend, so the browser sees one
  origin and the session cookie is first-party everywhere.

The manager's hard constraints are honoured literally: Go backend **without
`nexcommon` and without `nexcomponent`**, a separate browser frontend **without
`nexcomponent`**, and video served as plain files by Clover itself — no video
platform, no transcoding pipeline, no CDN.

## Technology Stack

`stack: go`. The target repository `clover-be` already declares a single Go
service (its README) and contains no pipeline artifacts (no `dags/`, no Spark
jobs, no `mappings/`), so the stack is settled by the existing repository, not
chosen here.

| Area | Choice | Reason |
|---|---|---|
| Language | Go 1.23 | Manager's constraint; single small service |
| HTTP | stdlib `net/http` + `http.ServeMux` method/path patterns | No framework needed for 7 routes; no `nexcommon` allowed; stdlib gives Range support for free |
| Database | PostgreSQL 16 via `github.com/jackc/pgx/v5/pgxpool` | House default for a service with persistent state; plain SQL, no ORM |
| Migrations | Embedded SQL files (`migrations/*.sql`) applied by a ~80-line runner using `schema_migrations` | No migration dependency; batched and idempotent |
| Password hashing | `golang.org/x/crypto/argon2` (argon2id, m=19456 KiB, t=2, p=1, PHC encoding) | bcrypt silently truncates at 72 bytes and REQ-026 allows 128-character passwords |
| Sessions | Opaque 32-byte random token in an HttpOnly cookie; SHA-256 of the token stored in `sessions` | Server-side revocation (REQ-021) without storing the secret |
| IDs | UUIDv4 (`github.com/google/uuid`) internal PKs; movie public id is its readable slug | Public ids are never the database row number |
| Metrics | `github.com/prometheus/client_golang` on a separate internal listener | Canonical exposition; no `nexcommon` available to reuse |
| Logging | stdlib `log/slog`, JSON to stdout, one line per request with `request_id` | Structured, dependency-free |
| Frontend | React 18 + Vite (JavaScript/JSX), `react-router-dom`, plain CSS built from the UIMap tokens | No `nexcomponent`; the approved mock is the design source |
| Frontend tests | Vitest + Testing Library; Playwright for the end-to-end cases | Matches the FE repo conventions to be scaffolded |
| Packaging | Multi-stage Docker images; compose stack (see Deployment) | One command runs the app |

Nothing is assumed to exist that this feature does not build. `clover-be` is
effectively greenfield (one README, no `go.mod`, no `CLAUDE.md`/`PROJECT.md`),
and `clover-fe` is empty. Every mechanism below — auth middleware, session
store, migration runner, metrics wiring, proxy — is decided here, not deferred
to "assumed to exist by implementation time".

## Scale

`app_scale: small` (spec.md). Confirmed numbers: a few hundred accounts (about
30 concurrent viewers), 20 films now and 20–50 within a year, films up to
30 minutes, one tenant, a short maintenance break acceptable.

The profile as it applies here:

- **One backend replica**, one PostgreSQL instance, **no cache and no queue**.
- Slow work is done in the request: the only slow operation is streaming bytes,
  which `http.ServeContent` pipes from disk (`io.Copy`) without buffering the
  file or blocking a request slot for long.
- **Sign-in is rate-limited per client IP** (30 attempts/min, burst 30) by an
  in-memory token bucket. This is correct only while there is one replica;
  moving to 2+ replicas moves the limiter to Postgres or Redis — noted, not
  built in v1.
- The frontend is a static build served by nginx; no server-side rendering.
- Media is local disk (`MEDIA_DIR`), mounted as a volume. At 30 concurrent
  viewers and ~1 Mbps per sample file this is ~30 Mbps outbound and light disk
  read — comfortably inside a small instance. There is no CDN in v1 by
  decision, not by omission.
- Monitoring: `/metrics` plus structured logs (see Observability). Backups:
  the Postgres volume and the media volume.
- Deliberate step down from the profile: none. Deliberate step up: none.

## Localization

English only (`en`), which is also the default and the fallback. There is no
locale switcher.

- **Language choice**: the request's language is ignored for v1 because only
  `en` exists — `Accept-Language` never changes the response (TC-006, TC-007).
  The mechanism is in place so adding a language later needs no rebuild of the
  API contract.
- **Backend dictionary**: `internal/i18n/en.json` maps every stable error code
  to its English message. Every error response carries both
  `error.code` (stable, machine-readable) and `error.message` (resolved from
  that dictionary) — see the Error contract.
- **Frontend dictionary**: `src/locales/en.json` holds all UI copy; components
  read text through one helper, never inline literals. The approved copy in
  `UIMap.md` is the source.
- **Formatting**: the only formatted values in v1 are timecodes (`mm:ss` /
  `h:mm:ss`) and the catalogue count; both are formatted with English rules.
  There are no dates, money or currency values in this feature.

## Permissions

A permission is `<path>:<action>`; a role is a list of grants. There are no
deny rules and no record-level ACLs in v1; **scope** is where a record belongs
to someone, and only the account and session endpoints are scoped (`own`).

| Permission | Meaning |
|---|---|
| `catalog:view` | See the movie catalogue, a movie's metadata and its poster |
| `playback:watch` | Request a movie's video stream |
| `account:view` | Read the signed-in account's own data |
| `session:create` | Sign in |
| `session:delete` | Sign out |

| Role | Grants |
|---|---|
| `visitor` (no valid session) | `catalog:view`, `session:create` |
| `member` (valid session) | `catalog:view`, `playback:watch`, `account:view`, `session:create`, `session:delete` |

Enforcement lives in one middleware: it resolves the session cookie to a role
and each route declares the permission it requires (a constant next to the
handler, not a string scattered in handlers). A signed-out request to a
`member` route returns `401 UNAUTHENTICATED`.

**Filter options**: not applicable — the catalogue has no search, filter, sort
or pagination (REQ-002), so there is no filter-options endpoint and no
secondary permission to design.

## Components

**New — backend (`clover-be`), all greenfield:**

| Component | Path | Responsibility |
|---|---|---|
| Entrypoint | `cmd/server` | Config, pool, router, listeners, graceful shutdown |
| Seed tool | `cmd/seed` | Idempotent seed of 20 movies + 7 accounts + media files |
| HTTP layer | `internal/httpapi` | Router, request-id/logging/recover/security-header/origin-check/rate-limit/auth middlewares, handlers, error envelope |
| Accounts | `internal/accounts` | User lookup (username or email), argon2id hash/verify, validation |
| Sessions | `internal/sessions` | Create/verify/revoke/extend, token hashing, expiry |
| Movies | `internal/movies` | Catalogue and single-movie queries, catalogue ordering |
| Media | `internal/media` | Poster serving, video streaming with Range, safe path resolution |
| Config | `internal/config` | Env parsing and validation |
| Observability | `internal/observability` | Prometheus registry, HTTP middleware, metrics listener |
| i18n | `internal/i18n` | Error code → English message catalogue |
| Migrations | `migrations/` + `internal/db` | Embedded SQL, `schema_migrations` runner |
| Media assets | `media/posters/`, `media/videos/` | Seeded posters (from the approved mock) and sample videos |

**New — frontend (`clover-fe`), all greenfield:**

| Component | Responsibility |
|---|---|
| `src/api/client.js` | `fetch` wrapper (`credentials: 'same-origin'`), error-envelope parsing, request-id capture |
| `src/auth/AuthProvider.jsx` | Boot-time `GET /v1/auth/me`, `signIn`, `signOut`, loading/anonymous/member state |
| `src/routes/*` | Catalogue, Sign-in, Player, Film-not-found, route guard |
| `src/components/*` | Header, film card, primary capsule button, ghost/quiet/icon buttons, facts card, browse chip, field, form alert, target chip, player stage, timeline/volume, toast, poster fallback, skeleton, spinner, empty panel |
| `src/styles/*` | UIMap design tokens and screen rules ported from `ui/mock/styles.css` |
| `src/locales/en.json` | All UI copy |
| `nginx.conf` + `Dockerfile` | SPA fallback, `/v1/` reverse proxy, security headers |

**Touched / untouched:** nothing existing is touched — both repositories have
no application code. `clover-be`'s README currently says the Go service serves
"the static UI"; that sentence is stale against this two-stack decision and is
`/go-dev`'s to correct in the code repo (this design does not edit it).

**Explicit boundaries:**

- The backend returns JSON and bytes only. No `html/template`, no embedded
  HTML, no `static/` UI folder served by Go.
- The frontend holds no catalogue data of its own and no credentials; every
  screen reads the API.
- Video bytes are never proxied through the frontend build or bundled into
  either image; nginx streams them from the API.

## Backend — Endpoints

All routes are mounted under the literal prefix **`/v1`** (except the
operational routes, which stay at the root). The prefix is a constant in
`internal/httpapi/router.go`, not configuration, because the frontend proxy and
the OpenAPI contract both pin it. Paths below are final.

| # | Method | Path | Consumer | Permission | Scope / auth | Notes |
|---|---|---|---|---|---|---|
| 1 | GET | `/v1/movies` | `fe` — Catalogue | `catalog:view` | public | All movies, ordered by `position`; `{"movies":[]}` when empty (REQ-023) |
| 2 | GET | `/v1/movies/{movieId}` | `fe` — Player, Film not found | `catalog:view` | public | One movie; `404 MOVIE_NOT_FOUND` for an unknown slug (REQ-018) |
| 3 | POST | `/v1/auth/login` | `fe` — Sign in | `session:create` | public | Body `{identifier, password}`; sets the cookie; `401 AUTH_INVALID_CREDENTIALS` on any refusal (REQ-005, REQ-016) |
| 4 | POST | `/v1/auth/logout` | `fe` — Header | `session:delete` | member; `own` session | Deletes the presented session and clears the cookie; idempotent `204` (REQ-008, REQ-021) |
| 5 | GET | `/v1/auth/me` | `fe` — boot, player revalidation | `account:view` | member; `own` account | `{user:{id,username}}` or `401 UNAUTHENTICATED` (REQ-007, TC-021) |
| 6 | GET | `/v1/media/movies/{movieId}/video` | `fe` — Player `<video>` | `playback:watch` | member | `206` with `Content-Range` for Range requests; no bytes without a session (REQ-010, REQ-011, TC-020, TC-061) |
| 7 | GET | `/v1/media/movies/{movieId}/poster` | `fe` — Catalogue cards | `catalog:view` | public | WebP bytes; public cache; `404` for an unknown slug (REQ-024, TC-062) |
| — | GET | `/healthz` | infra | public | internal network | Liveness, no database access |
| — | GET | `/readyz` | infra | public | internal network | Readiness, `SELECT 1` on the pool |
| — | GET | `/metrics` | infra | token | separate internal port | Prometheus exposition — never published publicly |

Notes that `/go-dev` must implement exactly:

- **Route mount prefix** is `/v1` on the API port (`:8080`); `/metrics` is on
  `:9090` and is not published by compose.
- `GET /v1/movies` response item shape:
  `{"id":"amber-drift","title":"Amber Drift","description":"…","posterUrl":"/v1/media/movies/amber-drift/poster","videoUrl":"/v1/media/movies/amber-drift/video","durationSeconds":600,"position":1}`.
  `posterUrl`/`videoUrl` are relative paths on the same origin and `position`
  is `movies.catalogue_order` in JSON form.
- `POST /v1/auth/login` accepts a maximum body of 4 KiB, rejects unknown JSON
  fields, and returns `{"user":{"id":"…","username":"alex"}}` — never the
  password hash, never the email.
- Video responses: `Content-Type: video/mp4`, `Accept-Ranges: bytes`,
  `Cache-Control: private, no-store`, `Vary: Cookie`; `HEAD` supported.
  Poster responses: `Cache-Control: public, max-age=86400`.
- Catalogue responses: `Cache-Control: public, max-age=60`.

## Frontend

UI scope: **yes** (`has_ui: true`). The FE repository is
`https://github.com/william22913/clover-fe`, base branch `main`. It is
currently **empty** (no commits, no `main` ref), so `/fe-dev` must create the
feature branch `feat/feat-001` from a base branch that has to exist first; see
"Open items" for the one prerequisite this design cannot perform itself.

`UIMap.md` and the mock in `ui/mock/` are the visual and interaction truth —
this section records only the engineering decisions the design leaves open.

**Routes and screens**

| Route | Screen(s) from `UIMap.md` | API calls |
|---|---|---|
| `/` | `#catalogue`, `#catalogue-signed-in`, `#catalogue-loading`, `#catalogue-empty`, `#catalogue-poster-error` | `GET /v1/movies`, `GET /v1/auth/me` (boot), `POST /v1/auth/logout` |
| `/sign-in?next=/watch/{movieId}` | `#signin`, `#signin-direct`, `#signin-required`, `#signin-error`, `#signin-loading` | `GET /v1/movies/{id}` (target chip), `POST /v1/auth/login` |
| `/watch/{movieId}` | `#player`, `#player-loading`, `#player-error` | `GET /v1/auth/me`, `GET /v1/movies/{id}`, `GET /v1/media/movies/{id}/video` |
| `/watch/{movieId}` (unknown id) | `#not-found` | `GET /v1/movies/{id}` → 404 |
| `*` | — | redirect to `/` |

**State and navigation**

- One `AuthProvider` holds `loading | anonymous | member`; it calls
  `GET /v1/auth/me` once on boot and exposes `signIn(identifier, password)`,
  `signOut()`, and `revalidate()`.
- The picked film is carried in the `next` query parameter
  (`/sign-in?next=/watch/{slug}`), so it survives reload and direct addresses
  (REQ-004, REQ-011). `next` is accepted **only** if it matches
  `^/watch/[a-z0-9-]+$`; anything else is discarded and the user goes to `/`
  (no open redirect).
- On successful sign-in: `next` present → navigate there with the "Signed in."
  toast; absent → `/` with "Signed in. Pick a film to watch." (TC-013).
- `/watch/:movieId` is wrapped in a session guard: anonymous → redirect to
  `/sign-in?next=…`; member → render the player.
- Three request states per screen: loading (skeletons/spinner per UIMap),
  error (screen-specific copy), ready. No global store; no `react-query` —
  one small `fetch` wrapper and per-screen hooks are enough for four screens.

**Player**

- One `<video preload="metadata" poster={posterUrl}>` element, custom controls
  from `UIMap.md`: play/pause disc, timeline `range` 0–1000 mapped to
  `currentTime`, elapsed/total timecodes, volume `range` 0–100, mute with
  restore-previous-volume, fullscreen via the Fullscreen API.
- Seeking sets `currentTime`; the browser issues a Range request and playback
  resumes from that point (REQ-010, TC-052).
- Error paths: a `<video>` `error`/`stalled` event or a 10-second load timeout
  shows the `#player-error` card with "Try again". Before showing it, the
  player issues `HEAD` on the video URL: `401` → session expired, so follow
  REQ-020 to sign-in; anything else → media error with retry (REQ-019,
  TC-059, TC-060).
- Before any playback action after 60 seconds of the last check, call
  `GET /v1/auth/me`; a `401` sends the viewer to
  `/sign-in?next=/watch/{id}` (REQ-020, TC-016).

**Responsive / accessibility**: breakpoints, stacking and the below-620 px
rules come from `UIMap.md`; every control is a real `<button>`/`<input>` with
an accessible name, focus rings use `--c-focus`, and
`prefers-reduced-motion: reduce` disables the entrance and transitions
(MO-018).

**Frontend tests**: Vitest + Testing Library for the button/field/alert
states and the `next`-param validation; Playwright for the end-to-end cases
(catalogue → sign-in → player, sign-out gating, error states, responsive
screenshots at 1440/1024/768/390 px). The FE runs against the real backend in
compose, never only a mock.

**Deployment**: multi-stage Dockerfile (node build → nginx). nginx serves the
SPA with `try_files $uri /index.html`, proxies `/v1/` to the API service with
`proxy_buffering off` for the media routes (so Range responses stream), and
sets the security headers in the Security design section.

## Data design

Full detail — ERD, normalization review, index list, batched migration plan
and seed rules — is in `database.md`, produced through the schema-design gate.
In summary the model is three tables:

- `movies` — seed-owned catalogue rows (slug public id, title, description,
  poster path, video path, duration, catalogue position).
- `accounts` — pre-created accounts (username, email, argon2id password hash),
  both identifiers stored lowercase with case-insensitive uniqueness.
- `sessions` — one row per sign-in (token SHA-256 hash, user reference,
  expiry, last-seen), hard-deleted on sign-out or expiry.

No soft delete in v1: nothing deletes a movie or an account, and sessions are
removed for real on sign-out — that is lifecycle, not a business status, and
no table carries a `status` value for it.

## Security design

Every row names the control, where it lives, and the test or check that proves
it. A row without a proof would be an open item; there are none below.

| # | Threat | Control | Where | Proof |
|---|---|---|---|---|
| S1 | Anonymous viewer watches video (REQ-011) | Video route requires `playback:watch`; `401` before any bytes; `Cache-Control: private, no-store`, `Vary: Cookie` | BE `internal/httpapi` + `internal/media` | TC-061, TC-020 |
| S2 | Stolen/old cookie reused after sign-out (REQ-021) | Sign-out hard-deletes the session row; every request re-verifies the token hash | BE `internal/sessions` | TC-020, TC-021 |
| S3 | Account enumeration (REQ-016) | One `401 AUTH_INVALID_CREDENTIALS` for unknown identifier, wrong password, malformed identifier and out-of-policy password; dummy argon2id verify when the user is missing | BE `internal/accounts` | TC-023, TC-024, TC-030, TC-034, TC-036 |
| S4 | Password cracking after a database leak | Argon2id (m=19456 KiB, t=2, p=1) with a per-hash random 16-byte salt, PHC-encoded; parameters travel with the hash so they can be raised later | BE `internal/accounts` | Unit test: encode/verify round-trip + wrong password fails |
| S5 | Password length truncation | Argon2id instead of bcrypt (no 72-byte limit); rune-count validation 8–128 | BE `internal/accounts` | TC-041, TC-042 |
| S6 | Session token theft from the database | Only the SHA-256 of the token is stored; the raw token exists only in the cookie | BE `internal/sessions` | Unit test: token not present in the row |
| S7 | Session cookie stolen by script | `HttpOnly`, `SameSite=Lax`, `Path=/`, `Secure` when `COOKIE_SECURE=true` (required in production), 30-day `Max-Age` | BE login handler | API test asserts cookie attributes |
| S8 | CSRF on sign-in/sign-out | `SameSite=Lax` **plus** an `Origin`/`Sec-Fetch-Site` same-origin check on every `POST` (`403 CSRF_REJECTED`); the check is against the configured public origin | BE middleware | API test: wrong `Origin` → 403 |
| S9 | Sign-in brute force without lockout (REQ-028) | Per-IP token bucket 30 requests/min, burst 30 — deliberately above TC-047's 11 attempts so a correct retry is never delayed; no per-account counter, no backoff | BE middleware | TC-047; API test: 11 attempts then success |
| S10 | Oversized or hostile request bodies | `http.MaxBytesReader` 4 KiB on login, unknown JSON fields rejected, strict content type | BE handlers | Unit test: 8 KiB body → 413; unknown field → 400 |
| S11 | Path traversal through media ids | Media ids are validated against `^[a-z0-9-]{1,64}$` and resolved by database lookup; file paths come from the row, joined under `MEDIA_DIR` and verified to stay inside it | BE `internal/media` | Unit test: `../` id → 404 |
| S12 | Open redirect through `next` | FE accepts `next` only when it matches `^/watch/[a-z0-9-]+$` | FE sign-in | Vitest: external URL discarded |
| S13 | Script injection / clickjacking in the browser | React escaping (no `dangerouslySetInnerHTML`); CSP, `X-Content-Type-Options: nosniff`, `Referrer-Policy: strict-origin-when-cross-origin`, `frame-ancestors 'none'` | FE nginx + React | Playwright/header check |
| S14 | Secrets in the repository | Only environment variables (`DATABASE_URL`, `PUBLIC_ORIGIN`, `COOKIE_SECURE`, `MEDIA_DIR`, `METRICS_ADDR`); `.env.example` with placeholders, no real values | BE config | Repository scan |
| S15 | SQL injection | Parameterized pgx queries only; no string-built SQL | BE repositories | Code review + scan |
| S16 | Log injection / untraceable errors | Server generates the request id (incoming `X-Request-Id` is ignored), echoes it in the body and header, and logs it on one line per request | BE middleware | API test: error body + header share the id |

CSP served by nginx:
`default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; media-src 'self'; font-src 'self'; connect-src 'self'; frame-ancestors 'none'; base-uri 'none'; form-action 'self'`.
`'unsafe-inline'` is limited to styles because the mock's entrance animation is
driven by inline style attributes; scripts stay strict.

## Error contract

Every error response is JSON with the same envelope and always carries a
request id in both the body and the `X-Request-Id` response header; the same id
is on the matching server log line.

```json
{
  "error": {
    "code": "AUTH_INVALID_CREDENTIALS",
    "message": "Incorrect username/email or password.",
    "requestId": "9f2c6a1b8d4e4f0a",
    "fields": { "identifier": "Enter your username or email." }
  }
}
```

| HTTP | Code | When | Fields | Message the user sees | What the screen does |
|---|---|---|---|---|---|
| 401 | `AUTH_INVALID_CREDENTIALS` | Login refused: unknown identifier, wrong password, malformed identifier, password outside 8–128/letter+digit | — | "Incorrect username/email or password." | Alert above the actions, identifier kept, password cleared, form usable (TC-016, TC-023) |
| 422 | `VALIDATION_ERROR` | Identifier or password empty after the identifier is trimmed (direct API use; the UI blocks it first) | `identifier` / `password` | "Enter your username or email." / "Enter your password." | Inline field error, focus the field (TC-025, TC-026) |
| 401 | `UNAUTHENTICATED` | Missing, expired or revoked session on `/v1/auth/me`, the video route | — | — | Go to `/sign-in?next=/watch/{id}` and return there after sign-in (REQ-004, REQ-020) |
| 403 | `CSRF_REJECTED` | `Origin`/`Sec-Fetch-Site` not same-origin on a POST | — | "Something went wrong. Please try again." + request id | Keep the form state; allow retry |
| 404 | `MOVIE_NOT_FOUND` | Unknown movie slug on `/v1/movies/{id}`, poster or video | — | — | Film not found screen with "Back to the catalogue" (REQ-018) |
| 404 | `NOT_FOUND` | Unknown API route | — | — | Not reached by the UI |
| 405 | `METHOD_NOT_ALLOWED` | Wrong verb on a known route | — | — | Not reached by the UI |
| 429 | `RATE_LIMITED` | Login limit exceeded (above the TC-047 threshold) | — | "Too many attempts. Please wait a moment and try again." + request id | Keep the form usable; re-enable after `Retry-After` |
| 500 | `INTERNAL_ERROR` | Unexpected server fault | — | "Something went wrong. Please try again." + request id | Error state with retry; the request id is shown so support can find the log line |

Anything not in this table is shown by the frontend only as the generic
message **with its request id** — never as a raw server string. Ids in URLs are
public slugs, never database row numbers, and the frontend treats every id as
an opaque string. Every foreign key and every filtered column is indexed — see
`database.md`.

## Observability

- **Current state**: greenfield. `clover-be` has no instrumentation, no router
  and no `/metrics` route; `nexcommon` is deliberately not used, so no library
  behaviour needs accounting for.
- **Decision**: wire Prometheus **once**, using `prometheus/client_golang` and
  one HTTP middleware, on a separate internal listener `:9090`
  (`METRICS_ADDR`). Not reused (nothing exists), not "not applicable".
- **Implementation boundary**: the middleware labels `route` (the registered
  pattern, not the raw path — no cardinals from ids), `method` and `status`,
  and records `http_requests_total` and
  `http_request_duration_seconds`; the error writer increments
  `http_errors_total{code}`. Two business counters only, per the spec:
  `clover_logins_total{result="ok|failed"}` and
  `clover_video_requests_total{result="ok|unauthorized|not_found"}`. No
  dashboards, no per-user labels.
- **Verification**: after a successful login and a refused video request, a
  scrape of `:9090/metrics` shows the two counters and a per-status
  `http_requests_total` sample for `/v1/auth/login` and the video route.
  Compose publishes only the frontend port.

## Non-functional concerns

- **Idempotency / re-run safety**: all GETs are safe. `POST /v1/auth/login`
  creates a new session per success (intentional — multi-device, REQ-022).
  `POST /v1/auth/logout` is idempotent (`204` even with no valid session) and
  deletes only the presented session. The seed command is idempotent: upserts
  by slug/username and skips media files that already exist, so it can run on
  every environment start.
- **Session expiry**: 30 days, sliding — a request after 24 hours extends
  `expires_at` once (`last_seen_at` is the throttle). Expired rows are deleted
  opportunistically on login; no cron and no queue.
- **Concurrency**: `accounts.username`/`accounts.email` uniqueness is a database
  constraint, not a check-then-insert. Session creation is a single insert.
  There is no shared mutable state in the process except the rate limiter and
  the metrics registry, both safe under one replica.
- **Deployment**: two images, one compose file. `clover-be`'s
  `docker-compose.yml` runs `postgres`, `be` and `fe` (frontend build context
  `../clover-fe`, both repos checked out side by side); `clover-fe`'s own
  `docker-compose.yml` runs the frontend alone against
  `API_UPSTREAM=http://localhost:8080` for frontend-only work. Only the
  frontend port is published publicly. Healthchecks: `/healthz` for `be`,
  HTTP 200 on `/` for `fe`.
- **Video delivery**: `http.ServeContent` with an open `*os.File`, so Range
  and conditional requests work and memory stays constant. MP4s are encoded
  faststart (`moov` before `mdat`) so playback begins before the download
  completes (REQ-010, TC-055).
- **Seed media**: posters are the approved mock assets
  (`ui/mock/assets/img/*.webp`, licences in `CREDITS.md`), copied by
  `/go-dev` into `media/posters/`. Sample videos are generated by the seed
  step with `ffmpeg` (1 fps, 320×180, silent AAC, low bitrate) so the files are
  small but real: **17 films 60–120 s, one 16-minute film** (so TC-052 can seek
  past 15 minutes) **and two 10-minute films** — `amber-drift` (TC-055) and
  `orbit` (whose copy already describes ten minutes). If `ffmpeg` is not
  available in the build environment, pre-generate the files once and commit
  them; the seed step then only verifies their presence.
- **Data volume**: 20 movies × one poster and one video; the 16-minute file at
  low bitrate is a few MB. Well inside the media volume at this scale.
- **Accessibility** is part of the design, not polish: labelled inputs, focus
  rings, reduced-motion, keyboard-operable player controls (UIMap MO-018).

## Test strategy

`testcase.md` (63 scenarios) is the authority; this table says **who proves
what** so the architecture can be exercised at every level.

| Testcase group | Level | Owner | Notes |
|---|---|---|---|
| TC-001 – TC-007 (catalogue, empty, language) | API + UI | `/go-dev` (API), `/fe-dev` (UI) | `GET /v1/movies` returns all 20 in order and `[]` when empty; `Accept-Language` ignored |
| TC-008 – TC-022 (sign-in gate, session, sign-out, multi-device) | API + end-to-end | `/go-dev` (login/me/logout/revocation), `/fe-dev` (navigation and carry) | Session revocation and video refusal proven at API level (TC-020, TC-021) |
| TC-023 – TC-048 (validation, boundaries, enumeration) | API + UI | `/go-dev` (handler rules), `/fe-dev` (pre-submit checks) | Boundary cases TC-029..TC-046 asserted directly against `POST /v1/auth/login` |
| TC-049 – TC-056 (player) | UI + end-to-end | `/fe-dev`; `/go-dev` proves Range/faststart at API level | TC-052 needs the 16-minute seeded film; TC-055 needs the 10-minute one |
| TC-057 – TC-061 (player errors, access control) | API + end-to-end | `/go-dev` (404/401 semantics), `/fe-dev` (screens) | TC-061: no bytes without a session, plain and Range |
| TC-062 – TC-063 (seed data) | API | `/go-dev` | Seed validation plus `GET /v1/movies` field checks; duration ≤ 1800 s enforced at seed and by DB CHECK |

Unit tests `/go-dev` must write: identifier/password validators, argon2id
encode/verify, token hashing, Range handler against a temp file, media path
traversal, error envelope + request id, origin/CSRF check, rate limiter
threshold, i18n completeness (every code has a message).

## Implementation Plan

- PART-1: Foundation and data layer — Go module, config, pgx pool, embedded migration runner, the three tables and their indexes, seed command (20 films, 7 accounts, posters, generated sample videos), `/healthz` and `/readyz`, Dockerfile and compose skeleton.
- PART-2: Auth and session core — argon2id hashing, session create/verify/revoke/extend, `POST /v1/auth/login`, `POST /v1/auth/logout`, `GET /v1/auth/me`, session cookie, permission middleware, origin/CSRF check, per-IP rate limit, error envelope with request id and the English message catalogue.
- PART-3: Catalogue and media — movies repository and service, `GET /v1/movies`, `GET /v1/movies/{movieId}`, poster handler, video handler with Range and faststart files, the 401/404 semantics and cache headers each route needs.
- PART-4: Observability and hardening — Prometheus middleware and the internal `/metrics` listener, business counters, structured request logging, security headers, panic recovery, and the backend test suite mapped to the 63 scenarios.
- PART-5: Frontend application — scaffold `clover-fe` from `UIMap.md` and the mock, design tokens and components, auth provider and route guard, catalogue, sign-in with `next` carry, player with custom controls and error/retry, empty/poster/not-found states, nginx proxy and Dockerfile, Vitest and Playwright coverage.

## Planned breaking changes

None — this is the first release of a new service; no existing consumer
depends on any endpoint.

## Spec traceability

One line per active `REQ-*` in `spec.md` v1, each with the concrete
architectural answer (where the rule is enforced or the flow is built).

- REQ-001 — `GET /v1/movies` is public and called by the Catalogue screen on first paint; no session check, no gate, poster/title/description all in the response.
- REQ-002 — The catalogue query returns every row ordered by `movies.catalogue_order`; the API has no pagination, search, filter or sort parameters and the frontend renders all 20 cards in one grid.
- REQ-003 — A film card links to `/watch/{slug}`; when `AuthProvider` is `member`, the route guard renders the Player directly with no sign-in screen.
- REQ-004 — While `anonymous`, the card links to `/sign-in?next=/watch/{slug}`; a successful login navigates to that validated `next` path, never to `/`.
- REQ-005 — `POST /v1/auth/login` takes one `identifier` and one `password`; the identifier is trimmed, lowercased and matched against `accounts.username` **or** `accounts.email` in one parameterized query.
- REQ-006 — No registration route or handler exists; unknown paths return the API's `404`; accounts exist only through `cmd/seed`.
- REQ-007 — Login sets a persistent `clover_session` cookie (HttpOnly, SameSite=Lax, 30-day Max-Age) backed by a `sessions` row with sliding expiry; `GET /v1/auth/me` restores the signed-in state on every boot.
- REQ-008 — Sign-out calls `POST /v1/auth/logout`, which deletes the session row and clears the cookie; the frontend drops to `anonymous`, keeps the catalogue and gates cards again.
- REQ-009 — The Player is an HTML5 `<video>` driven by the UIMap's custom controls: play/pause, timeline range mapped to `currentTime`, volume range, mute with restore, Fullscreen API.
- REQ-010 — The video handler serves through `http.ServeContent` with `Accept-Ranges: bytes` and `206` responses; MP4s are faststart-encoded so playback starts before the full file arrives and a seek issues a Range request.
- REQ-011 — The video route requires a valid session (`401` before any bytes) and the `/watch/:movieId` guard redirects anonymous visitors to sign-in with the film carried.
- REQ-012 — The Player fetches `GET /v1/movies/{movieId}` and renders `title` beside the stage; the catalogue card uses the same record so the title always matches the film.
- REQ-013 — The seed creates 20 films each with its own sample MP4 under `media/videos/`; `GET /v1/media/movies/{slug}/video` streams exactly that file.
- REQ-014 — English-only: the backend resolves every error text from `internal/i18n/en.json`, the frontend holds all copy in `src/locales/en.json`, and `Accept-Language` never changes a response.
- REQ-015 — "Cancel and keep browsing" navigates to `/` without calling the API; no session exists, so every card still routes through sign-in.
- REQ-016 — Any refused login returns the single `401 AUTH_INVALID_CREDENTIALS`; the frontend shows one alert, keeps the identifier and clears the password.
- REQ-017 — The frontend blocks submission when a field is blank after trimming (no request sent); the backend independently answers `422 VALIDATION_ERROR` with per-field messages as a safety net.
- REQ-018 — An unknown slug returns `404 MOVIE_NOT_FOUND` from `GET /v1/movies/{movieId}`; the frontend renders the Film not found screen with a link back to the catalogue.
- REQ-019 — The player bounds loading (video `error`/`stalled` or a 10-second timeout), shows the error card with "Try again", and distinguishes a `401` (session gone → sign-in) from a media failure by a `HEAD` on the video URL.
- REQ-020 — Playback actions revalidate `GET /v1/auth/me` at most once every 60 seconds; a `401` sends the viewer to `/sign-in?next=/watch/{slug}` and the player reopens for the same film after sign-in.
- REQ-021 — Sign-out hard-deletes the session row, so the old cookie fails `GET /v1/auth/me` and the video route; video responses carry `Cache-Control: private, no-store` and `Vary: Cookie` so no cached copy is served.
- REQ-022 — Sessions are per-login rows keyed by token hash, many per user; signing in on a second device inserts another row and `logout` deletes only the presented one.
- REQ-023 — An empty `movies` table returns `{"movies":[]}` with `200`; the frontend renders the empty panel and hides the count and facts card.
- REQ-024 — Posters are served from `/v1/media/movies/{slug}/poster`; the card's `<img onError>` swaps to the poster-unavailable tile while the title and description stay rendered.
- REQ-025 — Enforced in the login handler (trim, username `^[A-Za-z0-9_]{3,50}$` or a `net/mail`-valid address, lowercase before lookup, format failures answered generically) and backed by `accounts` CHECK constraints that stored values are lowercase plus unique constraints.
- REQ-026 — Enforced in the login handler: the password is never trimmed, its rune count must be 8–128 with at least one letter and one digit, and it is verified with argon2id (no bcrypt truncation); an empty password is `422`, any other violation the generic `401`.
- REQ-027 — `accounts.username` and `accounts.email` are stored lowercase with `UNIQUE` constraints, and `cmd/seed` verifies case-insensitive uniqueness before insert; the constraint, not application discipline, is the guarantee.
- REQ-028 — No account lockout, per-account counter or backoff exists; the only limit is the per-IP 30/min token bucket, set above TC-047's 11 attempts so the correct retry is never delayed.
- REQ-029 — `cmd/seed` requires a non-empty title, description, poster and video for every film and rejects a duration over 1800 s; DB `NOT NULL` plus `CHECK (duration_seconds BETWEEN 1 AND 1800)` enforce the boundary at rest.

## Open items

- **`clover-fe` has no base branch.** The repository is empty, so `/fe-dev`'s
  preflight cannot pass until `main` exists with one initial commit. This design
  does not create it — a code-repo write needs the manager's instruction. The
  intended values are recorded above (`frontend_repo`, `frontend_base_branch:
  main`); once the branch exists, `/fe-dev` creates `feat/feat-001` itself.
- **`clover-be`'s README still describes "the static UI" in the Go service.**
  That sentence predates the approved two-stack decision and is a code-repo
  edit for `/go-dev`, not a docs edit for this session.
- **`UIMap.md` frontmatter still says `status: draft` with `pr_url: null`**
  although PR #69 is merged. That is `/ui-preparation`'s file; the design
  content is complete and was used as the visual source, but the frontmatter
  is stale and its owner should refresh it.
