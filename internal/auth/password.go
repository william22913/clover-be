// Package auth holds password rules, hashing, opaque session tokens and the
// sign-in rate limiter. It has no HTTP knowledge.
package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"net/mail"
	"strings"
	"unicode"
	"unicode/utf8"
)

// BcryptCost is the recorded work factor for every stored password.
const BcryptCost = 12

// MaxPasswordBytes is bcrypt's own input limit.
const MaxPasswordBytes = 72

// ValidateEmail returns "" when the address is acceptable, otherwise one
// English message for the email field.
func ValidateEmail(email string) string {
	if email == "" {
		return "Enter your email address."
	}
	if len(email) > 254 {
		return "That email address is too long."
	}
	addr, err := mail.ParseAddress(email)
	if err != nil || addr.Address != email || strings.ContainsAny(email, " \t\r\n") {
		return "Enter a valid email address."
	}
	return ""
}

// ValidatePassword applies the sign-up rule from REQ-024: at least 8
// characters, at least one letter and at least one number, taken exactly as
// typed.
func ValidatePassword(password string) string {
	switch {
	case password == "":
		return "Enter a password."
	case utf8.RuneCountInString(password) < 8:
		return "Use at least 8 characters."
	case len(password) > MaxPasswordBytes:
		return "Use no more than 72 characters."
	}
	var hasLetter, hasNumber bool
	for _, r := range password {
		switch {
		case unicode.IsLetter(r):
			hasLetter = true
		case unicode.IsNumber(r):
			hasNumber = true
		}
	}
	switch {
	case !hasLetter:
		return "Include at least one letter."
	case !hasNumber:
		return "Include at least one number."
	}
	return ""
}

// ValidateSignInPassword only requires a non-empty password: composition rules
// belong to account creation, and an account made under another rule must
// still be able to sign in (REQ-024).
func ValidateSignInPassword(password string) string {
	if password == "" {
		return "Enter your password."
	}
	return ""
}

// HashPassword returns a bcrypt hash at the recorded cost.
func HashPassword(password string) (string, error) {
	hash, err := bcryptGenerate([]byte(password), BcryptCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// CheckPassword compares a plain password with a stored hash.
func CheckPassword(hash, password string) bool {
	return bcryptCompare([]byte(hash), []byte(password)) == nil
}

// NewToken returns a 256-bit URL-safe token and its SHA-256 hex digest. Only
// the digest is stored.
func NewToken() (token, hash string, err error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", "", err
	}
	token = base64.RawURLEncoding.EncodeToString(buf)
	return token, HashToken(token), nil
}

// HashToken is the stored form of a session token.
func HashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

// dummyHash is compared against when no account matched, so the unknown-email
// and wrong-password paths cost the same.
var dummyHash = func() string {
	h, err := HashPassword("clover-timing-equaliser-1")
	if err != nil {
		panic(err)
	}
	return h
}()

// CheckDummyPassword burns one bcrypt comparison. Call it when the email is
// unknown.
func CheckDummyPassword(password string) {
	CheckPassword(dummyHash, password)
}
