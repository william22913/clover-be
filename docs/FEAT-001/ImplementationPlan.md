---
feature_code: FEAT-001
service: clover
locked_version: 1
last_updated: 2026-11-13
---

# FEAT-001 — Implementation Plan Overview

> Summary only — see `architecture.md`'s `## Implementation Plan` section for
> full detail.

## Overview

Six ordered parts — order of work only, no timeline or estimate. All six build
the one `clover-be` repository (executed by `/go-dev` against `feat/FEAT-001`);
there is no `/fe-dev` handoff for this feature, because the front end is static
HTML/CSS/JS served by the same Go process. PART-1 to PART-4 build the API and
its data layer; PART-5 and PART-6 build the served UI on top of it. Parts are
sequencing, not gates — each can land as its own `/go-dev` PR on
`feat/FEAT-001`.

## Goals

Ship Clover v1 as a single Go standard-library service on PostgreSQL: a public
catalogue of thirteen films with posters, self-service bcrypt accounts with
opaque cookie sessions, a session-gated progressive MP4 stream over
`http.ServeContent`, and the approved mock's three screens — catalogue,
combined sign-in / create-account, and player — with every error and recovery
state the spec names.

## Implementation Plan

| Part | What gets built | Key deliverable |
|---|---|---|
| PART-1 | Service foundation | Go module, `internal/config`, pgx pool, embedded migration runner under an advisory lock, `log/slog`, request id + logging + metrics + security-header + CSRF middleware, the error envelope, `/healthz`, the internal metrics listener, `Dockerfile` and a compose stack (app + postgres), README. |
| PART-2 | Data layer | `0001_init.sql` (`users`, `movies`, `sessions`) with the two database roles and grants, the idempotent `seed/movies.sql` for the 13 films, the media assets under `media/`, and store queries for all three aggregates with their tests. |
| PART-3 | Identity | Password rules and bcrypt, sign-up / sign-in / sign-out / me, session issue and resolution, the rate limiter, and the CSRF check on the three POST routes. |
| PART-4 | Catalogue and streaming | `GET /v1/movies`, `GET /v1/movies/{slug}`, the session-gated `GET /v1/movies/{slug}/video` over `http.ServeContent`, poster and static serving, and the faststart check on the video asset. |
| PART-5 | UI shell | The three served routes, the client router, the API helper, the session boot, the catalogue page (hero, rails, cards) wired to `GET /v1/movies`, and the signed-in header. |
| PART-6 | UI auth and player | The combined sign-in / create-account screen with every state, the player with the five controls and the three failure states, the expired-session recovery path, and the end-to-end checks in the compose stack. |

PART-2 depends on PART-1; PART-3 depends on PART-2's tables; PART-4 depends on
PART-2 and PART-3's session middleware; PART-5 needs PART-4's public contract;
PART-6 needs PART-3, PART-4 and PART-5. PART-2's media assets and seed can be
prepared in parallel with PART-1 once the module exists.
