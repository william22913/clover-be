package accounts

import (
	"strings"
	"testing"
)

// TestPassword_TC045_CaseSensitive covers TC-045: verification is exact, so a
// lowercased variant never matches.
func TestPassword_TC045_CaseSensitive(t *testing.T) {
	hash, err := HashPassword("Watch1234")
	if err != nil {
		t.Fatalf("hash: %v", err)
	}
	ok, err := VerifyPassword("watch1234", hash)
	if err != nil {
		t.Fatalf("verify: %v", err)
	}
	if ok {
		t.Fatal("lowercased password matched")
	}
}

// TestPassword_TC046_TrailingSpaceIsSignificant covers TC-046 at the hash
// layer: the stored value is not trimmed or normalised.
func TestPassword_TC046_TrailingSpaceIsSignificant(t *testing.T) {
	hash, err := HashPassword("Watch1234 ")
	if err != nil {
		t.Fatalf("hash: %v", err)
	}
	if ok, _ := VerifyPassword("Watch1234 ", hash); !ok {
		t.Fatal("exact password with trailing space did not match")
	}
	if ok, _ := VerifyPassword("Watch1234", hash); ok {
		t.Fatal("password without the trailing space matched")
	}
}

// TestPassword_Argon2idPHC checks the stored form and a round trip.
func TestPassword_Argon2idPHC(t *testing.T) {
	hash, err := HashPassword("Abcdefg1")
	if err != nil {
		t.Fatalf("hash: %v", err)
	}
	if !strings.HasPrefix(hash, "$argon2id$v=19$m=19456,t=2,p=1$") {
		t.Fatalf("unexpected PHC string: %s", hash)
	}
	if ok, err := VerifyPassword("Abcdefg1", hash); err != nil || !ok {
		t.Fatalf("round trip failed: ok=%v err=%v", ok, err)
	}
	if ok, _ := VerifyPassword("Abcdefg2", hash); ok {
		t.Fatal("a different password matched")
	}
}

// TestPassword_MalformedHashIsAnError keeps a corrupt row from being read as a
// successful match.
func TestPassword_MalformedHashIsAnError(t *testing.T) {
	for _, encoded := range []string{"", "plain", "$argon2i$v=19$m=1,t=1,p=1$aa$bb", "$argon2id$v=19$m=x$aa$bb"} {
		if ok, err := VerifyPassword("Abcdefg1", encoded); err == nil || ok {
			t.Fatalf("malformed hash %q returned ok=%v err=%v", encoded, ok, err)
		}
	}
}

// TestVerifyDummy_AlwaysFails covers the account-enumeration defence (S3).
func TestVerifyDummy_AlwaysFails(t *testing.T) {
	VerifyDummy("Watch1234")
	ok, err := VerifyPassword("Watch1234", dummyHash)
	if err != nil {
		t.Fatalf("dummy hash is not a valid PHC string: %v", err)
	}
	if ok {
		t.Fatal("the dummy hash matched a real password")
	}
}
