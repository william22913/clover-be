---
feature_code: feat-001
service: clover
locked_version: 1
last_updated: 2026-11-14
---

# feat-001 — Backend Schema Overview

> Summary only — see `database.md` for the full schema-designer record
> and `architecture.md`'s `## Data design` for the query rationale.

## Overview

Three domain tables and one infrastructure table, all created by this
feature because both repositories are greenfield: `movies`, `accounts`,
`sessions` and the migration runner's `schema_migrations`. `clover/ERD.md`
shows the combined picture.

The model is deliberately small. There is no session-to-movie table, no
per-film ACL, and no watch history: playback authorisation is exactly "a
valid session exists", enforced by the application before any bytes are
written.

## Data Model

- `movies` — the seeded catalogue, read-only to the application. Public
  identifier is the unique `slug`; `catalogue_order` is unique and is what
  "Film 01 / 20" means; `duration_seconds` is bounded 1–1800 by a CHECK
  (REQ-029's 30-minute ceiling). `poster_path` and `video_path` are
  relative to `MEDIA_DIR`.
- `accounts` — pre-created accounts, written only by the seed command.
  `username` and `email` are stored lowercase with plain `UNIQUE`
  constraints plus `CHECK (value = lower(value))`, so case-insensitive
  uniqueness holds by construction rather than by a `citext` extension or
  a functional index. `password_hash` carries a `LIKE '$argon2id$%'` guard
  so no plaintext or foreign hash format can land in the column.
- `sessions` — one row per sign-in, keyed by the unique 32-byte
  `token_hash` (SHA-256 of the raw cookie token). `account_id` is a real
  foreign key with deliberate `RESTRICT`: the plan never deletes an
  account, and a cascade would silently destroy sessions. `expires_at`
  starts at `created_at + 30 days` and slides on use.
- **One surrogate UUID primary key per table**, with the natural keys
  (`slug`, `username`, `email`, `token_hash`) as `UNIQUE` constraints, per
  the house schema rule — the API never exposes a row number, and a later
  feature can foreign-key to `movies.id` without the slug becoming
  load-bearing.
- **No soft delete and no status-as-deletion.** v1 deletes no film and no
  account, and a session is removed for real on sign-out — lifecycle, not
  a business status, so no `is_deleted`/`status` column is introduced.
- **Two deliberate derived-but-stored columns**, stated rather than
  accidental: `poster_path` (the asset layout is a data decision, not a
  backend convention) and `catalogue_order` (editorial order is not
  derivable from any other column). Nothing else is denormalized; the
  catalogue count is computed at query time, never stored.
- Every foreign key and every filtered column is indexed: `movies.slug`,
  `movies.catalogue_order`, `accounts.username`, `accounts.email`,
  `sessions.token_hash`, plus non-unique indexes on `sessions.account_id`
  and `sessions.expires_at`. Nothing else is indexed — there is no search,
  filter or sort in v1.
- Migrations arrive in three independently reviewable batches —
  `accounts`, `movies`, `sessions` — applied by a small embedded runner
  that takes an advisory lock and runs one transaction per file. No DDL is
  written in this design.

## Access Control

- At the API level, the catalogue and posters are public; video requires a
  valid session (`401` before any bytes); `GET /v1/auth/me` and sign-out
  act only on the account the presented session belongs to.
- There is no row-level security and no user-supplied identifier on any
  scoped query. The "own" scope is satisfied by the query key: the session
  is resolved from the cookie's token hash to its `account_id`, and no
  endpoint accepts another account's id as input.
- The database roles are separated by function: the runtime role reads
  `movies` and `accounts`, reads and writes `sessions`, and holds no DDL
  rights; the seed command writes `movies` and `accounts` only; the
  migration step holds DDL rights and runs once at deploy.
- Fixture accounts seeded for the test design are test data by
  construction and must be replaced before any real deployment.
