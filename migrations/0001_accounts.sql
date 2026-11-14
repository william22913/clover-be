CREATE TABLE accounts (
    id uuid PRIMARY KEY,
    username text NOT NULL,
    email text NOT NULL,
    password_hash text NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT accounts_username_key UNIQUE (username),
    CONSTRAINT accounts_email_key UNIQUE (email),
    CONSTRAINT accounts_username_format CHECK (username ~ '^[a-z0-9_]{3,50}$'),
    CONSTRAINT accounts_email_lowercase CHECK (email = lower(email) AND btrim(email) <> ''),
    CONSTRAINT accounts_password_hash_argon2 CHECK (password_hash LIKE '$argon2id$%')
);
