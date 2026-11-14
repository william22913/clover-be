---
feature_code: feat-001
service: clover
locked_version: 1
last_updated: 2026-11-14
---

# feat-001 — Technical Requirements Document

> Overview summary only — see `architecture.md` for the full design and
> its rationale.

## Overview

Clover v1 is the whole product, so this feature builds both applications
from nothing: a Go API (`clover-be`) that serves the catalogue JSON,
poster images and the video stream, and a React + Vite frontend
(`clover-fe`) that renders the Catalogue, Sign-in, Player and Film not
found screens and is reverse-proxied to the API on one origin.

Seven API routes under `/v1` carry the product: the catalogue list and a
single film, sign-in, sign-out, the signed-in account, the video stream
and the poster. The catalogue and posters are public; the video and the
account require a valid session. There is no self sign-up, no admin, no
search, filter or pagination, and no third party of any kind.

Data is three domain tables — `movies`, `accounts`, `sessions` — plus the
migration runner's bookkeeping. All content is seeded: 20 films with real
sample videos and 7 test-fixture accounts. Sessions are opaque tokens
stored only as a SHA-256 hash with a sliding 30-day expiry.

Two constraints shape the build and are not open questions: the manager's
stack decision (Go without `nexcommon`, a separate frontend without
`nexcomponent`) and the approved look, which comes from `UIMap.md` and the
mock in `ui/mock/` rather than from any house UI layer.

## Technical Goals

- Serve the full catalogue to an anonymous visitor: one `GET /v1/movies`
  returning every film, ordered by the catalogue's editorial position, with
  no authentication and no pagination.
- Gate playback on a valid session, never on a client-side check: the video
  route answers `401` before writing a byte, and the frontend's route guard
  is a convenience, not the control.
- Carry the picked film through sign-in so a signed-out visitor who selects
  a film lands on that film's player, not on the catalogue — including when
  the player address was opened directly.
- Stream progressively with real Range support from faststart-encoded MP4s,
  so playback starts before the file arrives and seeking resumes at the
  requested point.
- Keep sign-in honest without account enumeration or lockout: one generic
  refusal for every failure mode, constant-cost verification even when the
  account does not exist, and no per-account lockout (REQ-028).
- Make sign-out complete on the server: the session row is deleted, so the
  old cookie fails both the account and the video route, and video
  responses are never cacheable.
- Hash passwords with argon2id, because the spec allows 128-character
  passwords and bcrypt truncates at 72 bytes; store identifiers lowercase
  so uniqueness is case-insensitive by construction.
- Emit one JSON error envelope with a request id in the body, in the
  `X-Request-Id` header and on the server log line, and never surface a raw
  server string in the UI.
- Wire Prometheus once on an internal listener with low-cardinality labels
  and two business counters; publish only the frontend port.
- Keep the interface English-only, with every string in a message catalogue
  on each side, and accessibility treated as design (labelled inputs, focus
  rings, keyboard-operable player, reduced motion).

## Technology Stack

| Area | Tool | Purpose |
|---|---|---|
| Backend language/runtime | Go 1.23 | Standard library `net/http` with the ServeMux method/path patterns — seven routes do not justify a router dependency. |
| Shared libraries | none | `nexcommon` is deliberately excluded by the manager's decision; the service is self-contained. |
| Frontend | React 18 + Vite (JavaScript) | Four screens and one player; `react-router` for routes, no UI framework and no `nexcomponent`. |
| Persistence | PostgreSQL 16 + `pgx` v5 | Parameterized queries only; no ORM. |
| Migrations | Embedded SQL + a small Go runner | `embed.FS`, `schema_migrations`, one transaction per file, advisory lock at startup — no migration framework. |
| Passwords | argon2id (`x/crypto/argon2`) | PHC-encoded hash carrying its own parameters, so they can be raised by rehashing. |
| Sessions | Opaque token (32 random bytes) + SHA-256 in the database | `HttpOnly`/`SameSite=Lax` cookie; the raw token never reaches storage. |
| Video delivery | `http.ServeContent` over an open file | Range, `206` and conditional requests with constant memory; faststart MP4s. |
| Observability | `prometheus/client_golang` | One HTTP middleware and two business counters on an internal `:9090` listener. |
| Logging | `log/slog` | One structured line per request, carrying the request id. |
| Frontend packaging | Multi-stage Dockerfile → nginx | SPA fallback, `/v1/` reverse proxy with buffering off for media, security headers and CSP. |
| Local stack | Docker Compose | `postgres`, `be`, `fe` in `clover-be`; a frontend-only compose in `clover-fe`. Only the frontend port is published. |
