package store

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
)

// SessionUser is the result of resolving a session token.
type SessionUser struct {
	UserID int64
	Email  string
}

// CreateSession issues a new session row and, when revokeTokenHash is not
// empty, revokes the session presented with the sign-in in the same
// transaction. It also records the sign-in time on the account.
func (s *Store) CreateSession(ctx context.Context, userID int64, tokenHash, revokeTokenHash string) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin sign-in: %w", err)
	}
	defer tx.Rollback(context.Background()) //nolint:errcheck

	if revokeTokenHash != "" {
		if _, err := tx.Exec(ctx,
			`UPDATE sessions SET revoked_at = now() WHERE token_hash = $1 AND revoked_at IS NULL`,
			revokeTokenHash); err != nil {
			return fmt.Errorf("revoke previous session: %w", err)
		}
	}
	if _, err := tx.Exec(ctx,
		`INSERT INTO sessions (token_hash, user_id) VALUES ($1, $2)`, tokenHash, userID); err != nil {
		return fmt.Errorf("insert session: %w", err)
	}
	if _, err := tx.Exec(ctx,
		`UPDATE users SET last_sign_in_at = now(), updated_at = now() WHERE id = $1`, userID); err != nil {
		return fmt.Errorf("touch user: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit sign-in: %w", err)
	}
	return nil
}

// ResolveSession returns the account behind a live session. Expiry and
// revocation are checked against the database clock in the same statement.
func (s *Store) ResolveSession(ctx context.Context, tokenHash string) (SessionUser, error) {
	var u SessionUser
	err := s.pool.QueryRow(ctx, `
		SELECT u.id, u.email
		FROM sessions s
		JOIN users u ON u.id = s.user_id
		WHERE s.token_hash = $1
		  AND s.revoked_at IS NULL
		  AND s.expires_at > now()
		  AND NOT u.is_deleted`, tokenHash,
	).Scan(&u.UserID, &u.Email)
	if errors.Is(err, pgx.ErrNoRows) {
		return SessionUser{}, ErrNotFound
	}
	if err != nil {
		return SessionUser{}, fmt.Errorf("resolve session: %w", err)
	}
	return u, nil
}

// RevokeSession signs out the session named by a token and reports how many
// live sessions it touched. It is idempotent: a missing or already-revoked
// token is not an error.
func (s *Store) RevokeSession(ctx context.Context, tokenHash string) (int64, error) {
	tag, err := s.pool.Exec(ctx,
		`UPDATE sessions SET revoked_at = now() WHERE token_hash = $1 AND revoked_at IS NULL`,
		tokenHash)
	if err != nil {
		return 0, fmt.Errorf("revoke session: %w", err)
	}
	return tag.RowsAffected(), nil
}

// PurgeSessions deletes expired rows and old revocations. It returns how many
// rows went.
func (s *Store) PurgeSessions(ctx context.Context) (int64, error) {
	tag, err := s.pool.Exec(ctx, `
		DELETE FROM sessions
		WHERE expires_at < now()
		   OR (revoked_at IS NOT NULL AND revoked_at < now() - interval '30 days')`)
	if err != nil {
		return 0, fmt.Errorf("purge sessions: %w", err)
	}
	return tag.RowsAffected(), nil
}
