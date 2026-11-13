package store

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// User is an account row. The hash never leaves this package's callers inside
// internal/auth.
type User struct {
	ID           int64
	Email        string
	PasswordHash string
}

// CreateUserAndSession inserts the account and its first session in one
// transaction. A duplicate email surfaces as ErrEmailTaken from the unique
// lower(email) index.
func (s *Store) CreateUserAndSession(ctx context.Context, email, passwordHash, tokenHash string) (int64, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("begin sign-up: %w", err)
	}
	defer tx.Rollback(context.Background()) //nolint:errcheck

	var id int64
	err = tx.QueryRow(ctx,
		`INSERT INTO users (email, password_hash) VALUES ($1, $2) RETURNING id`,
		email, passwordHash).Scan(&id)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" && pgErr.ConstraintName == "ux_users_email_lower" {
			return 0, ErrEmailTaken
		}
		return 0, fmt.Errorf("insert user: %w", err)
	}
	if _, err := tx.Exec(ctx,
		`INSERT INTO sessions (token_hash, user_id) VALUES ($1, $2)`, tokenHash, id); err != nil {
		return 0, fmt.Errorf("insert session: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return 0, fmt.Errorf("commit sign-up: %w", err)
	}
	return id, nil
}

// UserByEmail returns the account for an email, compared case-insensitively.
func (s *Store) UserByEmail(ctx context.Context, email string) (User, error) {
	var u User
	err := s.pool.QueryRow(ctx, `
		SELECT id, email, password_hash
		FROM users
		WHERE lower(email) = lower($1) AND NOT is_deleted`, email,
	).Scan(&u.ID, &u.Email, &u.PasswordHash)
	if errors.Is(err, pgx.ErrNoRows) {
		return User{}, ErrNotFound
	}
	if err != nil {
		return User{}, fmt.Errorf("get user by email: %w", err)
	}
	return u, nil
}
