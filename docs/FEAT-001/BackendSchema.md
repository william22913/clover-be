---
feature_code: FEAT-001
service: clover
locked_version: 1
last_updated: 2026-11-13
---

# FEAT-001 — Backend Schema Overview

> Summary only — see `database.md` for the full schema-designer record
> (constraints, indexes, normalization decisions, migration batches and seed
> data) and `architecture.md`'s `## Data Design` for the rationale.

## Overview

FEAT-001 creates the entire `clover` schema: three new tables in one
PostgreSQL 16 database, with no pre-existing tables to touch. They fall into
two groups — identity (`users`, `sessions`) and catalogue (`movies`) — and fit
in a single migration batch, `0001_init.sql`, created in the order users →
movies → sessions so the foreign key exists before it is referenced.
`clover/ERD.md` holds the cumulative ERD with a `## Tables by feature` index.

## Data Model

- **`users`** — `id` bigserial, `email` `VARCHAR(254)` unique on
  `lower(email) WHERE NOT is_deleted`, bcrypt `password_hash` `VARCHAR(72)`,
  `last_sign_in_at`, and the house audit and soft-delete columns. CHECKs pin the
  email shape (trimmed once at the boundary, so
  `email = btrim(email)`), the bcrypt `$2a$`/`$2b$`/`$2y$` hash shape, and
  soft-delete consistency (`is_deleted` ⇔ `deleted_at IS NOT NULL`). No
  plaintext password column exists, so password rules are request-validator
  rules, not constraints.
- **`movies`** — `id` bigserial, public `slug` `VARCHAR(64)` unique among live
  rows, `title`, `short_description`, `release_year`, `runtime_seconds`,
  `poster_file`, `still_file`, `video_file`, `display_order`, `is_published`,
  and the audit and soft-delete columns. File-name CHECKs pin
  `^[a-z0-9][a-z0-9._-]{0,127}$` and forbid `..`, so the row can never name a
  path outside the media directory. A partial index
  `(display_order) WHERE is_published AND NOT is_deleted` is the catalogue
  read; `display_order` is deliberately not unique, and ordering is
  `display_order, id` so the list is deterministic.
- **`sessions`** — `id` bigserial, `token_hash` `VARCHAR(64)` unique (SHA-256
  hex of the cookie token), `user_id` FK to `users` with `ON DELETE CASCADE`,
  `created_at`, `expires_at` defaulting to `now() + interval '7 days'` from the
  database clock, and nullable `revoked_at`. No audit triple: a session is
  system-originated and is never soft-deleted, only revoked. Indexes on
  `user_id` and on `expires_at` serve the auth lookup and the expiry sweep.
- **The authentication query** is one statement —
  `SELECT s.user_id, u.email FROM sessions s JOIN users u ON u.id = s.user_id
  WHERE s.token_hash = $1 AND s.revoked_at IS NULL AND s.expires_at > now()
  AND NOT u.is_deleted` — so revocation, expiry and a soft-deleted owner are
  one atomic decision on the database clock. Sign-out is idempotent:
  `UPDATE sessions SET revoked_at = now() WHERE token_hash = $1 AND
  revoked_at IS NULL`.
- **Seed data** — thirteen films (slug, title, year, runtime seconds, poster,
  still and the shared placeholder `clover-sample.mp4`) inserted by an
  idempotent `seed/movies.sql` with `ON CONFLICT (slug) WHERE NOT is_deleted DO
  UPDATE`. Test fixtures for the missing-file and unavailable-video cases are
  not seeded.
- PostgreSQL conventions: `BIGSERIAL` surrogate keys, a public identifier
  (`slug`) separate from the internal id, `TIMESTAMPTZ` audit columns, TEXT or
  `VARCHAR` plus CHECK for enum-like values, partial unique indexes for
  live-row uniqueness.

## Access Control

- **Public reads:** the catalogue (`GET /v1/movies`), a single film
  (`GET /v1/movies/{slug}`), posters and the static shell. No session, no
  account lookup.
- **Session-gated:** the video stream (`GET /v1/movies/{slug}/video`),
  `GET /v1/auth/me` and sign-out (`POST /v1/auth/sign-out`). The stream route
  re-resolves the session on every range request, so revocation takes effect
  mid-playback.
- **Writes:** only `POST /v1/auth/sign-up`, `sign-in` and `sign-out` write, and
  they write only `users`, `sessions` and `users.last_sign_in_at`. The
  catalogue is seed-only — no route inserts, updates or deletes a film.
- **Two database roles.** `clover_migrator` owns the schema and runs DDL;
  `clover_app` is the runtime login with `SELECT/INSERT/UPDATE` on the three
  tables and `DELETE` only on `sessions`, and no DDL. `PUBLIC` loses `CREATE`
  on the schema.
- **Soft delete, not hard delete.** `users` and `movies` carry the
  `is_deleted`/`deleted_at`/`deleted_by` triple and are never physically
  removed; `sessions` rows are retained with `revoked_at` until the expiry
  sweep.
- **Nothing prints or logs a password, a password hash, a session token or an
  email.** Log lines carry request id, method, route template, status,
  duration and user id only.
