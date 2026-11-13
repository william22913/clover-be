package auth

import (
	"strings"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

// TC-033 — the malformed addresses from the test design are all refused, and
// ordinary addresses are accepted.
func TestValidateEmail_TC033(t *testing.T) {
	for _, bad := range []string{
		"", "ann.example.com", "ann@", "@example.com", "ann@@example.com", "ann example@example.com",
		"ann\texample@example.com",
	} {
		if msg := ValidateEmail(bad); msg == "" {
			t.Fatalf("ValidateEmail(%q) accepted it", bad)
		}
	}
	for _, good := range []string{"ann@example.com", "ann.b@example.co.uk", "ann+tag@example.com"} {
		if msg := ValidateEmail(good); msg != "" {
			t.Fatalf("ValidateEmail(%q) = %q", good, msg)
		}
	}
	if msg := ValidateEmail(strings.Repeat("a", 250) + "@example.com"); msg == "" {
		t.Fatal("an over-long address must be refused")
	}
}

// TC-032, TC-035, TC-036, TC-037, TC-038, TC-039, TC-044 — the sign-up password
// rule and its inclusive 8-character boundary.
func TestValidatePassword_TCodes(t *testing.T) {
	cases := []struct {
		name     string
		password string
		want     string // "" means accepted
	}{
		{"TC-032 empty", "", "Enter a password."},
		{"TC-035 seven characters", "Abc1234", "at least 8 characters"},
		{"TC-036 eight characters", "Abc12345", ""},
		{"TC-037 no digit", "Abcdefgh", "one number"},
		{"TC-038 no letter", "12345678", "one letter"},
		{"TC-039 symbol", "Clover2026!", ""},
		{"TC-044 internal space", "Abc 1234", ""},
		{"bcrypt limit", strings.Repeat("a1", 37), "72 characters"},
	}
	for _, tc := range cases {
		msg := ValidatePassword(tc.password)
		if tc.want == "" {
			if msg != "" {
				t.Fatalf("%s: got %q, want accepted", tc.name, msg)
			}
			continue
		}
		if !strings.Contains(msg, tc.want) {
			t.Fatalf("%s: got %q, want it to mention %q", tc.name, msg, tc.want)
		}
	}
}

// TC-045 — sign-in only requires a non-empty password; composition rules are
// for account creation.
func TestValidateSignInPassword_TC045(t *testing.T) {
	if msg := ValidateSignInPassword(""); msg == "" {
		t.Fatal("an empty sign-in password must be refused")
	}
	if msg := ValidateSignInPassword("x"); msg != "" {
		t.Fatalf("a short sign-in password must be accepted, got %q", msg)
	}
}

func TestHashPassword_UsesRecordedCost(t *testing.T) {
	hash, err := HashPassword("Passw0rd1")
	if err != nil {
		t.Fatal(err)
	}
	cost, err := bcrypt.Cost([]byte(hash))
	if err != nil {
		t.Fatal(err)
	}
	if cost != BcryptCost {
		t.Fatalf("cost = %d, want %d", cost, BcryptCost)
	}
	if !CheckPassword(hash, "Passw0rd1") {
		t.Fatal("the hash must match its own password")
	}
	if CheckPassword(hash, "Passw0rd2") {
		t.Fatal("a different password must not match")
	}
}

// Tokens are 256-bit, unique, and stored only as a SHA-256 hex digest.
func TestNewToken(t *testing.T) {
	first, firstHash, err := NewToken()
	if err != nil {
		t.Fatal(err)
	}
	second, _, err := NewToken()
	if err != nil {
		t.Fatal(err)
	}
	if first == second {
		t.Fatal("two tokens must differ")
	}
	if len(first) < 43 {
		t.Fatalf("token length = %d, want at least 43 characters", len(first))
	}
	if len(firstHash) != 64 || strings.ToLower(firstHash) != firstHash {
		t.Fatalf("hash = %q, want 64 lowercase hex characters", firstHash)
	}
	if HashToken(first) != firstHash {
		t.Fatal("the stored hash must be reproducible from the token")
	}
	if HashToken(first) == first {
		t.Fatal("the token itself must never be the stored value")
	}
}

// The unknown-email path still burns one comparison, and must not panic.
func TestCheckDummyPassword(t *testing.T) {
	CheckDummyPassword("anything-at-all")
	CheckDummyPassword("")
	if len(dummyHash) == 0 || dummyHash == "anything-at-all" {
		t.Fatal("dummy hash must be a real bcrypt hash")
	}
}
