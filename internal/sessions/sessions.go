// Package sessions owns server-side sessions: opaque tokens, their stored
// SHA-256 hashes, sliding expiry, and revocation. The raw token never reaches
// the database.
package sessions

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ErrInvalid means the presented token has no live session.
var ErrInvalid = errors.New("sessions: invalid session")

const (
	// TTL is the fixed session lifetime (REQ-007, database.md).
	TTL = 30 * 24 * time.Hour
	// slideAfter is how old last_seen_at must be before a request extends the
	// session, so the row is not written on every request.
	slideAfter = 24 * time.Hour
)

// Session is a live session resolved from a token.
type Session struct {
	ID       uuid.UUID
	Account  uuid.UUID
	Username string
	// Slid reports that Resolve extended the server-side expiry, so the
	// caller must also re-issue the cookie with a fresh lifetime.
	Slid bool
}

// Create inserts a new session for accountID and returns the raw token. The
// cookie carries this token; only its SHA-256 is stored.
func Create(ctx context.Context, pool *pgxpool.Pool, accountID uuid.UUID) (string, error) {
	token, err := newToken()
	if err != nil {
		return "", err
	}
	hash := HashToken(token)
	id := uuid.New()
	_, err = pool.Exec(ctx,
		`INSERT INTO sessions (id, token_hash, account_id, created_at, updated_at, last_seen_at, expires_at)
		 VALUES ($1, $2, $3, now(), now(), now(), now() + $4::interval)`,
		id, hash, accountID, TTL.String(),
	)
	if err != nil {
		return "", fmt.Errorf("sessions: create: %w", err)
	}
	return token, nil
}

// Resolve returns the live session for token, sliding its expiry when it has
// not been used for a day. An unknown, expired or revoked token is ErrInvalid.
func Resolve(ctx context.Context, pool *pgxpool.Pool, token string) (Session, error) {
	if token == "" {
		return Session{}, ErrInvalid
	}
	hash := HashToken(token)

	var (
		session  Session
		expires  time.Time
		lastSeen time.Time
	)
	err := pool.QueryRow(ctx,
		`SELECT s.id, s.account_id, a.username, s.expires_at, s.last_seen_at
		   FROM sessions s
		   JOIN accounts a ON a.id = s.account_id
		  WHERE s.token_hash = $1`,
		hash,
	).Scan(&session.ID, &session.Account, &session.Username, &expires, &lastSeen)
	if errors.Is(err, pgx.ErrNoRows) {
		return Session{}, ErrInvalid
	}
	if err != nil {
		return Session{}, fmt.Errorf("sessions: resolve: %w", err)
	}

	now := time.Now()
	if !expires.After(now) {
		_, _ = pool.Exec(ctx, `DELETE FROM sessions WHERE id = $1`, session.ID)
		return Session{}, ErrInvalid
	}
	if now.Sub(lastSeen) >= slideAfter {
		_, err = pool.Exec(ctx,
			`UPDATE sessions
			    SET expires_at = now() + $2::interval, last_seen_at = now(), updated_at = now()
			  WHERE id = $1`,
			session.ID, TTL.String(),
		)
		if err != nil {
			return Session{}, fmt.Errorf("sessions: slide expiry: %w", err)
		}
		session.Slid = true
	}
	return session, nil
}

// Revoke deletes the session for token. It is safe to call with an unknown or
// already-revoked token: sign-out is idempotent.
func Revoke(ctx context.Context, pool *pgxpool.Pool, token string) error {
	if token == "" {
		return nil
	}
	if _, err := pool.Exec(ctx, `DELETE FROM sessions WHERE token_hash = $1`, HashToken(token)); err != nil {
		return fmt.Errorf("sessions: revoke: %w", err)
	}
	return nil
}

// DeleteExpired removes every expired row. Login calls it opportunistically so
// no cron or queue is needed.
func DeleteExpired(ctx context.Context, pool *pgxpool.Pool) error {
	if _, err := pool.Exec(ctx, `DELETE FROM sessions WHERE expires_at <= now()`); err != nil {
		return fmt.Errorf("sessions: delete expired: %w", err)
	}
	return nil
}

// HashToken returns the SHA-256 of the raw token, which is what the database
// stores and looks sessions up by.
func HashToken(token string) []byte {
	sum := sha256.Sum256([]byte(token))
	return sum[:]
}

func newToken() (string, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", fmt.Errorf("sessions: read token: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(raw), nil
}
