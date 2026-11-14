---
feature_code: feat-001
service: clover
locked_version: 1
last_updated: 2026-11-14
---

# feat-001 — Implementation Plan Overview

> Summary only — see `architecture.md`'s `## Implementation Plan` section
> for full detail.

## Overview

Five ordered parts — order of work only, no timeline or estimate. PART-1
to PART-4 build `clover-be` (executed by `/go-dev` against `feat/feat-001`
in the clover-be repo); PART-5 builds `clover-fe` (executed by `/fe-dev`
from `architecture.md`'s `## Frontend` section, after the API contract
from PART-2 and PART-3 exists).

## Goals

Ship Clover v1 end to end on the manager's stack decision — a Go API
without `nexcommon` and a separate React frontend without `nexcomponent`
— with the catalogue public, playback session-gated, the picked film
carried through sign-in, progressive video with seeking, complete
server-side sign-out, and the approved look from `UIMap.md` and the mock.
Every one of the 63 cases in `testcase.md` must be reachable from these
parts.

## Implementation Plan

| Part | What gets built | Key deliverable |
|---|---|---|
| PART-1 | Foundation and data layer | Go module, config, `pgx` pool, embedded migration runner, the three tables and their indexes, `cmd/seed` (20 films, 7 fixture accounts, posters from the mock, generated sample videos), `/healthz` and `/readyz`, Dockerfile and compose skeleton. |
| PART-2 | Auth and session core | Argon2id hashing, session create/verify/revoke/extend, `POST /v1/auth/login`, `POST /v1/auth/logout`, `GET /v1/auth/me`, the session cookie, permission middleware, the origin/CSRF check, the per-IP rate limit, and the error envelope with request id plus the English message catalogue. |
| PART-3 | Catalogue and media | Movies repository and service, `GET /v1/movies`, `GET /v1/movies/{movieId}`, the poster handler, and the video handler with Range support over faststart files — with the `401`/`404` semantics and cache headers each route needs. |
| PART-4 | Observability and hardening | Prometheus middleware and the internal `/metrics` listener, the two business counters, structured request logging, security headers, panic recovery, and the backend test suite mapped to the 63 scenarios. |
| PART-5 | Frontend application | `clover-fe` scaffolded from `UIMap.md` and the mock: tokens and components, auth provider and route guard, catalogue, sign-in with the `next` carry, the player with custom controls and error/retry, the empty/poster/not-found states, the nginx proxy and Dockerfile, and Vitest plus Playwright coverage. |

PART-2, PART-3 and PART-4 depend on PART-1. PART-3's video route depends on
PART-2's session check. PART-5 depends on PART-2's auth contract and
PART-3's response shapes; it is a separate repository and a separate
`/fe-dev` run.

One prerequisite is outside these parts and is recorded in
`architecture.md`'s "Open items": `clover-fe` is empty, so its `main` base
branch must exist with one commit before `/fe-dev` can start.

## Handoff values

Literal values for the build steps. Nothing below is inferred, defaulted or
re-derived at run time; the frontmatter of `architecture.md` carries the
same four repository values.

| Item | Value |
|---|---|
| Service | `clover` |
| Feature code | `feat-001` |
| Backend clone URL | `https://github.com/william22913/clover-be` |
| Backend base branch | `main` — the repository's default branch |
| Backend feature branch | `feat/feat-001` |
| Frontend clone URL | `https://github.com/william22913/clover-fe` |
| Frontend base branch | `main` — the repository's default branch, **created with one initial commit before `/fe-dev` starts**, because the repository is currently empty and has no `main` ref |
| Frontend feature branch | `feat/feat-001` |
| API base path | `/v1`; all other routes are operational only |
| Ports | API `8080` and metrics `9090` are internal to the compose network; the frontend publishes the only host port, `3000` → nginx `80` |
| Docs the build reads | `clover/feature/feat-001/`: `spec.md`, `testcase.md`, `architecture.md`, `database.md`, `TRD.md`, `BackendSchema.md`, `ImplementationPlan.md`, `UIMap.md`, `AppFlow.md`, `ui/mock/` |

Compose files, by repository:

| Repository | File | What runs in it |
|---|---|---|
| `clover-be` | `docker-compose.yml` | `postgres`, `be`, `fe`; the frontend service's build context is `../clover-fe`, so both repositories are cloned side by side under the same parent directory |
| `clover-be` | `docker-compose.verify.yml` | Override used by the stack verifier: `postgres` + `be` only, one replica, explicit small memory limits and a tiny seed, so the backend can be verified when the sibling `clover-fe` checkout is not available to the runner |
| `clover-fe` | `docker-compose.yml` | The frontend alone, against `API_UPSTREAM=http://localhost:8080`, for frontend-only work |
