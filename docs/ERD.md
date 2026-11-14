# clover — Entity Relationship Diagram

Cumulative ERD across every feature built for this service, maintained as part of each feature's own docs PR. It shows the combined picture only — each feature's own `database.md` holds the rationale.

```mermaid
erDiagram
    accounts {
        uuid id PK "app-generated v4"
        text username UK "lowercase, 3-50 [a-z0-9_]"
        text email UK "lowercase, syntax checked by app/seed"
        text password_hash "argon2id PHC string"
        timestamptz created_at
        timestamptz updated_at
    }

    movies {
        uuid id PK "app-generated v4"
        text slug UK "lowercase hyphenated public id"
        text title
        text description
        text poster_path "relative to MEDIA_DIR"
        text video_path "relative to MEDIA_DIR"
        integer duration_seconds "1..1800"
        integer catalogue_order UK "1..NN, film 01..NN"
        timestamptz created_at
        timestamptz updated_at
    }

    sessions {
        uuid id PK "app-generated v4"
        bytea token_hash UK "SHA-256 of the raw cookie token"
        uuid account_id FK "references accounts.id"
        timestamptz created_at
        timestamptz updated_at
        timestamptz last_seen_at
        timestamptz expires_at
    }

    schema_migrations {
        bigint version PK "from the migration filename"
        timestamptz applied_at
    }

    accounts ||--o{ sessions : "account_id holds many concurrent sessions"
```

## Tables by feature

| Feature | Tables added/changed |
|---|---|
| [feat-001](feature/feat-001/database.md) | movies, accounts, sessions, schema_migrations |
