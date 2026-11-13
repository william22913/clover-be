---
feature_code: FEAT-001
service: clover
spec_version: 1
last_updated: 2026-11-13
---

# Clover — App Flow Overview (FEAT-001)

## Overview

The user-journey view of Clover v1. Summary only — `spec.md` is the
authoritative source, and each journey below cites the `REQ-*` it summarizes.
Screen design, layout and visual direction belong to `/ui-preparation`
(`UIMap.md`), not to this document.

<!-- BEGIN:DOCPREP -->

## Key User Journeys

### 1. Browse before signing up

A first-time visitor opens Clover and lands straight on the movie catalogue —
posters, titles and short descriptions, no login wall. They can look around
freely and leave without ever making an account.
(REQ-001, REQ-002, REQ-011, REQ-016)

### 2. Pick a movie and start watching (new account)

The visitor selects a movie. Because they are signed out, they are taken to a
combined login / sign-up screen. They create an account with an email and a
password, and the next thing they see is that same movie's player, signed in
and ready to play. Playback then starts, and they can pause, scrub, change the
volume and go fullscreen.
(REQ-003, REQ-004, REQ-005, REQ-008, REQ-009, REQ-024)

### 3. Return and carry on watching (existing account)

A returning visitor opens the site already signed in, picks a movie and
watches immediately — no login step. If their session has lapsed, they sign in
and are returned to the movie they were trying to play.
(REQ-006, REQ-007, REQ-012, REQ-019)

### 4. Sign out

A signed-in viewer signs out and is back to the anonymous browsing experience:
catalogue still visible, player locked behind the login screen again.
(REQ-010, REQ-020)

## Main pages and actions

| Page / action | What the visitor can do | Requirements |
|---|---|---|
| Catalogue (public) | See every movie with poster, title and short description; select one to watch | REQ-001, REQ-002, REQ-003 |
| Login / sign-up screen | Sign in with an existing account, or create one; see inline errors and retry without losing input | REQ-004, REQ-005, REQ-006, REQ-013, REQ-014, REQ-015, REQ-016 |
| Player page | Watch the selected movie with play, pause, scrub, volume and fullscreen; seek to any point | REQ-008, REQ-009 |
| Player page — recovery | See a clear message and a way back when the movie is missing or the video fails; retry playback after signing in again | REQ-017, REQ-018, REQ-019 |
| Sign out (any signed-in page) | End the session so nothing can be replayed afterwards | REQ-010, REQ-020, REQ-021 |

## Where this stops

This document describes what a person must be able to do. Visual design,
layout, components and motion are out of scope here and are mapped against the
manager's supplied design reference by `/ui-preparation` in `UIMap.md`.

<!-- END:DOCPREP -->

<!-- BEGIN:CODEPREP -->

## Route map

The browser talks only to the one `clover-be` service. The Go server answers
each page path with the same static shell from `web/`; the client then reads
`location.pathname` and renders the matching screen. Every screen below is a
server path, and each row lists the endpoints that load it. A route marked
**public** renders without a session; **session** routes send the visitor to
`/auth?next={slug}` and return them afterwards.

| Server path | Screen | Backend calls | Access |
|---|---|---|---|
| `/` | Catalogue | `GET /v1/movies`, `GET /v1/auth/me` (to pick the header state) | public |
| `/auth?next={slug}` | Combined sign-in / create account | `GET /v1/movies/{slug}`, `POST /v1/auth/sign-in`, `POST /v1/auth/sign-up`, `GET /v1/auth/me` | public |
| `/play/{slug}` | Player | `GET /v1/auth/me`, `GET /v1/movies/{slug}`, `GET /v1/movies/{slug}/video` (ranged), `POST /v1/auth/sign-out` | session |
| any unknown path | Not found — render the catalogue | none | public |

Static delivery from the same origin: `web/app.css`, `web/app.js` and the
fonts under `web/assets/fonts/`; posters and stills under
`/media/poster/` and `/media/still/`; the film itself only through
`/v1/movies/{slug}/video`, never as a direct file path.

## Request flow

```text
browser (static shell served from web/)
  -> fetch helper               credentials: same-origin, JSON body
  -> net/http.ServeMux          /v1/{resource}, method + path pattern
       -> middleware chain      request id -> logging -> metrics
                                -> security headers -> CSRF (POST only)
       -> route handler
            -> session middleware   (session-gated routes only)
            -> decode + validate    (MaxBytesReader 1 MiB, DisallowUnknownFields)
            -> store query          (pgx, parameterized SQL)
       -> JSON body, or the error envelope, plus X-Request-Id
  -> client router              renders the screen for location.pathname
```

The session travels in the `clover_session` cookie, so `credentials:
"same-origin"` is the whole auth plumbing — nothing is kept in local or session
storage. A `401 session_expired` on the player makes it probe
`GET /v1/auth/me`, then go to `/auth?next={slug}&reason=expired`; signing in
returns to the same film.

## Backend components

- **`internal/config`** — environment-only configuration (`DATABASE_URL`,
  optional `MIGRATE_DATABASE_URL`, `ADDR`, `METRICS_ADDR`, cookie settings),
  validated at startup with no secret defaulted in code.
- **`internal/store`** — pgx pool and the parameterized queries for the three
  aggregates: catalogue list, film by slug, user by lower(email), user insert,
  session insert, session resolve against `now()`, session revoke.
- **`internal/auth`** — password rules and bcrypt (cost 12), token generation
  and SHA-256 hashing, the session middleware, and the in-process sign-in
  rate limiter.
- **`internal/api`** — the `http.ServeMux` routes, the middleware chain, request
  decoding and validation, the error envelope, and the static and page shell
  handlers.
- **`internal/media`** — resolves a film row to its file inside the media
  directory and streams it with `http.ServeContent`; maps a missing or
  unreadable file to `404 video_unavailable`.
- **`internal/metrics`** — collectors and the second `http.Server` on
  `METRICS_ADDR`, with the status-capturing, route-template-labelled
  middleware.

## Frontend components

- **Client router** — maps `location.pathname` to the catalogue, auth or player
  screen; uses `history.pushState` and the three server paths above.
- **API helper** — one `fetch` wrapper that sets the JSON body, sends the
  cookie, and turns the error envelope into a typed failure the screen can
  switch on.
- **Catalogue page** — the header (anonymous or signed-in), the hero, the rails
  and the poster cards from `GET /v1/movies`, in `display_order`.
- **Auth screen** — the sign-in / create-account tabs over one form, inline
  field errors under their fields, the credential notice, the duplicate-email
  error with its sign-in link, and the `next` destination carried through both
  modes.
- **Player** — one `<video>` with native controls off and the five custom
  controls (play/pause, scrubber, volume/mute, fullscreen), plus the missing
  film, video error and expired session panels, each with its way back.

## Cross-cutting behaviour

- Every string the visitor sees is English; there is no i18n layer and no
  locale selection in v1.
- The header state comes from `GET /v1/auth/me` once at boot; a 401 is not an
  error on public pages, it just means anonymous.
- Sign-out posts, clears the cookie, and returns to `/` in the anonymous
  state; the catalogue stays browsable.
- A failed media element shows the "The film could not be loaded." panel with
  "Try again" — there is no spinner state that outlives a failed load.
- The mock's preview query flags (for example `?demo=1`) are not part of the
  shipped UI.

<!-- END:CODEPREP -->
