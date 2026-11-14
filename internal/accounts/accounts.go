// Package accounts owns account records, identifier and password validation,
// and argon2id hashing. Accounts are seeded; v1 has no write path for them.
package accounts

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"regexp"
	"strings"
	"unicode"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ErrNotFound means no account matches the identifier.
var ErrNotFound = errors.New("accounts: not found")

// usernamePattern is REQ-025 after trimming and lowercasing.
var usernamePattern = regexp.MustCompile(`^[a-z0-9_]{3,50}$`)

// Account is the part of an account row the API and the sign-in check use.
type Account struct {
	ID           uuid.UUID
	Username     string
	Email        string
	PasswordHash string
}

// Identifier is a normalized sign-in identifier.
type Identifier struct {
	// Value is the trimmed, lowercased identifier.
	Value string
	// IsEmail reports which column Value should match.
	IsEmail bool
}

// ParseIdentifier trims the raw identifier and decides whether it is a
// username or an email address. A value that is neither is not a valid
// identifier; the login handler answers every failure with one generic error.
func ParseIdentifier(raw string) (Identifier, error) {
	value := strings.ToLower(strings.TrimSpace(raw))
	if value == "" {
		return Identifier{}, fmt.Errorf("accounts: identifier is empty")
	}
	if usernamePattern.MatchString(value) {
		return Identifier{Value: value}, nil
	}
	if validEmail(value) {
		return Identifier{Value: value, IsEmail: true}, nil
	}
	return Identifier{}, fmt.Errorf("accounts: identifier is neither a username nor an email address")
}

func validEmail(value string) bool {
	addr, err := mail.ParseAddress(value)
	if err != nil {
		return false
	}
	// Reject display names ("Alex <a@b.com>") and anything mail.ParseAddress
	// tolerates beyond a bare address.
	if addr.Address != value {
		return false
	}
	at := strings.LastIndex(value, "@")
	if at <= 0 || at == len(value)-1 {
		return false
	}
	domain := value[at+1:]
	return strings.Contains(domain, ".") && !strings.HasPrefix(domain, ".") && !strings.HasSuffix(domain, ".")
}

// ValidatePassword enforces REQ-026: 8–128 characters, at least one letter and
// at least one digit. The value is taken exactly as typed — never trimmed or
// case-folded.
func ValidatePassword(password string) error {
	runes := []rune(password)
	if len(runes) < 8 || len(runes) > 128 {
		return fmt.Errorf("accounts: password must be 8-128 characters")
	}
	var hasLetter, hasDigit bool
	for _, r := range runes {
		if unicode.IsLetter(r) {
			hasLetter = true
		}
		if unicode.IsDigit(r) {
			hasDigit = true
		}
	}
	if !hasLetter || !hasDigit {
		return fmt.Errorf("accounts: password must contain a letter and a digit")
	}
	return nil
}

// ByIdentifier loads the account whose username or email equals ident.Value.
// The identifier is matched lowercased, which is how both columns are stored.
func ByIdentifier(ctx context.Context, pool *pgxpool.Pool, ident Identifier) (Account, error) {
	var acc Account
	var err error
	if ident.IsEmail {
		err = pool.QueryRow(ctx,
			`SELECT id, username, email, password_hash FROM accounts WHERE email = $1`,
			ident.Value,
		).Scan(&acc.ID, &acc.Username, &acc.Email, &acc.PasswordHash)
	} else {
		err = pool.QueryRow(ctx,
			`SELECT id, username, email, password_hash FROM accounts WHERE username = $1`,
			ident.Value,
		).Scan(&acc.ID, &acc.Username, &acc.Email, &acc.PasswordHash)
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return Account{}, ErrNotFound
	}
	if err != nil {
		return Account{}, fmt.Errorf("accounts: lookup: %w", err)
	}
	return acc, nil
}

// ByID loads one account by its internal id.
func ByID(ctx context.Context, pool *pgxpool.Pool, id uuid.UUID) (Account, error) {
	var acc Account
	err := pool.QueryRow(ctx,
		`SELECT id, username, email, password_hash FROM accounts WHERE id = $1`, id,
	).Scan(&acc.ID, &acc.Username, &acc.Email, &acc.PasswordHash)
	if errors.Is(err, pgx.ErrNoRows) {
		return Account{}, ErrNotFound
	}
	if err != nil {
		return Account{}, fmt.Errorf("accounts: lookup by id: %w", err)
	}
	return acc, nil
}
