# clover — Entity Relationship Diagram

Cumulative ERD across every feature built for this service, maintained by /code-preparation as part of each feature's own PR.
It is not a substitute for the feature's own `database.md`, which carries the rationale.

```mermaid
erDiagram
    users ||--o{ sessions : "has"

    users {
        bigserial id PK
        varchar email "unique on lower(email)"
        varchar password_hash "bcrypt"
        timestamptz last_sign_in_at
        timestamptz created_at
        timestamptz updated_at
        bigint created_by
        bigint updated_by
        boolean is_deleted
        timestamptz deleted_at
        bigint deleted_by
    }

    movies {
        bigserial id PK
        varchar slug "public id, unique"
        varchar title
        varchar short_description
        smallint release_year
        integer runtime_seconds
        varchar poster_file
        varchar still_file
        varchar video_file
        smallint display_order
        boolean is_published
        timestamptz created_at
        timestamptz updated_at
        bigint created_by
        bigint updated_by
        boolean is_deleted
        timestamptz deleted_at
        bigint deleted_by
    }

    sessions {
        bigserial id PK
        varchar token_hash "sha256 hex, unique"
        bigint user_id FK
        timestamptz created_at
        timestamptz expires_at
        timestamptz revoked_at
    }
```

## Tables by feature

| Feature | Tables added/changed |
|---|---|
| [FEAT-001](feature/FEAT-001/database.md) | users, movies, sessions |
