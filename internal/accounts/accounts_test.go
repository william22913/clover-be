package accounts

import (
	"strings"
	"testing"
)

// TestParseIdentifier_TC025_TC027_EmptyIsRejected covers TC-025 and TC-027: an
// empty or whitespace-only identifier is empty after trimming.
func TestParseIdentifier_TC025_TC027_EmptyIsRejected(t *testing.T) {
	for _, raw := range []string{"", "    ", "\t\n "} {
		if _, err := ParseIdentifier(raw); err == nil {
			t.Fatalf("ParseIdentifier(%q) accepted a blank identifier", raw)
		}
	}
}

// TestParseIdentifier_TC028_TrimsSurroundingSpaces covers TC-028.
func TestParseIdentifier_TC028_TrimsSurroundingSpaces(t *testing.T) {
	ident, err := ParseIdentifier("  alex  ")
	if err != nil {
		t.Fatalf("trimmed identifier rejected: %v", err)
	}
	if ident.Value != "alex" || ident.IsEmail {
		t.Fatalf("got %+v, want alex as a username", ident)
	}
}

// TestParseIdentifier_TC029_TC030_UsernameLength covers the 3–50 rule.
func TestParseIdentifier_TC029_TC030_UsernameLength(t *testing.T) {
	if _, err := ParseIdentifier("abc"); err != nil {
		t.Fatalf("TC-029: 3-character username rejected: %v", err)
	}
	if _, err := ParseIdentifier("ab"); err == nil {
		t.Fatal("TC-030: 2-character username accepted")
	}
}

// TestParseIdentifier_TC031_TC032_UsernameMaximum covers the upper boundary.
func TestParseIdentifier_TC031_TC032_UsernameMaximum(t *testing.T) {
	if _, err := ParseIdentifier(strings.Repeat("a", 50)); err != nil {
		t.Fatalf("TC-031: 50-character username rejected: %v", err)
	}
	if _, err := ParseIdentifier(strings.Repeat("a", 51)); err == nil {
		t.Fatal("TC-032: 51-character username accepted")
	}
}

// TestParseIdentifier_TC033_TC034_AllowedCharacters covers the character set.
func TestParseIdentifier_TC033_TC034_AllowedCharacters(t *testing.T) {
	ident, err := ParseIdentifier("mira_k9")
	if err != nil || ident.Value != "mira_k9" {
		t.Fatalf("TC-033: mira_k9 rejected: %v", err)
	}
	for _, raw := range []string{"mira-k", "mira k", "mira.k", "mira@k"} {
		if _, err := ParseIdentifier(raw); err == nil {
			t.Fatalf("TC-034: %q accepted as a username", raw)
		}
	}
}

// TestParseIdentifier_TC035_TC036_Email covers valid and malformed emails.
func TestParseIdentifier_TC035_TC036_Email(t *testing.T) {
	ident, err := ParseIdentifier("mira.k@example.com")
	if err != nil {
		t.Fatalf("TC-035: valid email rejected: %v", err)
	}
	if !ident.IsEmail || ident.Value != "mira.k@example.com" {
		t.Fatalf("TC-035: got %+v, want an email", ident)
	}
	for _, raw := range []string{"alex@", "@example.com", "alex@@example.com", "alex example.com", "Alex <alex@example.com>"} {
		if _, err := ParseIdentifier(raw); err == nil {
			t.Fatalf("TC-036: malformed email %q accepted", raw)
		}
	}
}

// TestParseIdentifier_TC037_TC038_CaseInsensitive covers case folding for both
// identifier kinds; folding happens on the way in, so the lookup is exact.
func TestParseIdentifier_TC037_TC038_CaseInsensitive(t *testing.T) {
	user, err := ParseIdentifier("ALEX")
	if err != nil || user.Value != "alex" || user.IsEmail {
		t.Fatalf("TC-037: got %+v, %v", user, err)
	}
	email, err := ParseIdentifier("ALEX.RIVERA@EXAMPLE.COM")
	if err != nil || email.Value != "alex.rivera@example.com" || !email.IsEmail {
		t.Fatalf("TC-038: got %+v, %v", email, err)
	}
}

// TestValidatePassword_TC039_TC042_Length covers the 8–128 boundary.
func TestValidatePassword_TC039_TC042_Length(t *testing.T) {
	if err := ValidatePassword("Abcdefg1"); err != nil {
		t.Fatalf("TC-039: 8-character password rejected: %v", err)
	}
	if err := ValidatePassword("Abcdef1"); err == nil {
		t.Fatal("TC-040: 7-character password accepted")
	}
	if err := ValidatePassword("A1" + strings.Repeat("x", 126)); err != nil {
		t.Fatalf("TC-041: 128-character password rejected: %v", err)
	}
	if err := ValidatePassword("A1" + strings.Repeat("x", 127)); err == nil {
		t.Fatal("TC-042: 129-character password accepted")
	}
}

// TestValidatePassword_TC043_TC044_Composition covers the letter-and-digit rule.
func TestValidatePassword_TC043_TC044_Composition(t *testing.T) {
	if err := ValidatePassword("Abcdefgh"); err == nil {
		t.Fatal("TC-043: letters-only password accepted")
	}
	if err := ValidatePassword("12345678"); err == nil {
		t.Fatal("TC-044: digits-only password accepted")
	}
}

// TestValidatePassword_TC046_WhitespaceIsSignificant covers TC-046: the value
// is never trimmed, so a trailing space is part of the password.
func TestValidatePassword_TC046_WhitespaceIsSignificant(t *testing.T) {
	if err := ValidatePassword("Watch1234 "); err != nil {
		t.Fatalf("password with a trailing space rejected: %v", err)
	}
}
