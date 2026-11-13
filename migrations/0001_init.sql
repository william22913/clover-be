-- FEAT-001 batch 1: users, movies, sessions, their constraints and indexes,
-- then the two database roles and their grants.
-- Creation order is users -> movies -> sessions (the FK depends on users).

CREATE TABLE users (
    id             BIGSERIAL PRIMARY KEY,
    email          VARCHAR(254) NOT NULL,
    password_hash  VARCHAR(72)  NOT NULL,
    last_sign_in_at TIMESTAMPTZ NULL,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_by     BIGINT NULL,
    updated_by     BIGINT NULL,
    is_deleted     BOOLEAN NOT NULL DEFAULT FALSE,
    deleted_at     TIMESTAMPTZ NULL,
    deleted_by     BIGINT NULL,
    CONSTRAINT ck_users_email_shape CHECK (
        char_length(email) BETWEEN 3 AND 254
        AND position('@' in email) > 1
        AND email = btrim(email)
    ),
    CONSTRAINT ck_users_password_hash CHECK (password_hash ~ '^\$2[aby]\$[0-9]{2}\$.{53}$'),
    CONSTRAINT ck_users_soft_delete CHECK (is_deleted = (deleted_at IS NOT NULL))
);

CREATE TABLE movies (
    id                BIGSERIAL PRIMARY KEY,
    slug              VARCHAR(64)  NOT NULL,
    title             VARCHAR(200) NOT NULL,
    short_description VARCHAR(500) NOT NULL,
    release_year      SMALLINT     NOT NULL,
    runtime_seconds   INTEGER      NOT NULL,
    poster_file       VARCHAR(128) NOT NULL,
    still_file        VARCHAR(128) NULL,
    video_file        VARCHAR(128) NOT NULL,
    display_order     SMALLINT     NOT NULL,
    is_published      BOOLEAN NOT NULL DEFAULT TRUE,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_by        BIGINT NULL,
    updated_by        BIGINT NULL,
    is_deleted        BOOLEAN NOT NULL DEFAULT FALSE,
    deleted_at        TIMESTAMPTZ NULL,
    deleted_by        BIGINT NULL,
    CONSTRAINT ck_movies_slug CHECK (slug ~ '^[a-z0-9-]{1,64}$'),
    CONSTRAINT ck_movies_file_names CHECK (
        (poster_file ~ '^[a-z0-9][a-z0-9._-]{0,127}$' AND position('..' in poster_file) = 0)
        AND (video_file ~ '^[a-z0-9][a-z0-9._-]{0,127}$' AND position('..' in video_file) = 0)
        AND (still_file IS NULL OR (still_file ~ '^[a-z0-9][a-z0-9._-]{0,127}$' AND position('..' in still_file) = 0))
    ),
    CONSTRAINT ck_movies_release_year CHECK (release_year BETWEEN 1888 AND 2100),
    CONSTRAINT ck_movies_runtime CHECK (runtime_seconds > 0),
    CONSTRAINT ck_movies_display_order CHECK (display_order > 0),
    CONSTRAINT ck_movies_soft_delete CHECK (is_deleted = (deleted_at IS NOT NULL))
);

CREATE TABLE sessions (
    id         BIGSERIAL PRIMARY KEY,
    token_hash VARCHAR(64) NOT NULL,
    user_id    BIGINT      NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    expires_at TIMESTAMPTZ NOT NULL DEFAULT (now() + interval '7 days'),
    revoked_at TIMESTAMPTZ NULL,
    CONSTRAINT ck_sessions_token_hash CHECK (token_hash ~ '^[0-9a-f]{64}$'),
    CONSTRAINT ck_sessions_expiry CHECK (expires_at > created_at)
);

-- Every uniqueness rule is a real index, never an application pre-check.
CREATE UNIQUE INDEX ux_users_email_lower ON users (lower(email)) WHERE NOT is_deleted;
CREATE UNIQUE INDEX ux_movies_slug ON movies (slug) WHERE NOT is_deleted;
CREATE INDEX ix_movies_catalogue ON movies (display_order) WHERE is_published AND NOT is_deleted;
CREATE UNIQUE INDEX ux_sessions_token_hash ON sessions (token_hash);
CREATE INDEX ix_sessions_user_id ON sessions (user_id);
CREATE INDEX ix_sessions_expiry ON sessions (expires_at);

-- Two roles: the migrator owns the schema, the app is the runtime login.
-- The roles themselves (and their passwords) are created by the privileged
-- bootstrap, scripts/init-db.sh, which also makes the migrator the owner of
-- schema public; creating a login role requires CREATEROLE, which the
-- migrator deliberately does not have.
ALTER TABLE users OWNER TO clover_migrator;
ALTER TABLE movies OWNER TO clover_migrator;
ALTER TABLE sessions OWNER TO clover_migrator;
ALTER SEQUENCE users_id_seq OWNER TO clover_migrator;
ALTER SEQUENCE movies_id_seq OWNER TO clover_migrator;
ALTER SEQUENCE sessions_id_seq OWNER TO clover_migrator;

REVOKE ALL ON SCHEMA public FROM PUBLIC;
REVOKE CREATE ON SCHEMA public FROM PUBLIC;
GRANT USAGE ON SCHEMA public TO clover_app, clover_migrator;

GRANT SELECT, INSERT, UPDATE ON users, movies, sessions TO clover_app;
GRANT DELETE ON sessions TO clover_app;
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO clover_app;
GRANT ALL ON users, movies, sessions TO clover_migrator;
GRANT ALL ON ALL SEQUENCES IN SCHEMA public TO clover_migrator;
