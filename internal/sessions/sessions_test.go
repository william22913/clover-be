package sessions

import (
	"bytes"
	"testing"
)

// TestSessions_TokenIsOpaqueAndHashed checks the two properties the schema
// relies on: the token is raw-URL base64 of 32 random bytes and the stored
// value is a 32-byte SHA-256 that is not the token itself.
func TestSessions_TokenIsOpaqueAndHashed(t *testing.T) {
	token, err := newToken()
	if err != nil {
		t.Fatalf("newToken: %v", err)
	}
	if len(token) != 43 {
		t.Fatalf("token length = %d, want 43 base64url characters", len(token))
	}
	other, err := newToken()
	if err != nil {
		t.Fatalf("newToken: %v", err)
	}
	if token == other {
		t.Fatal("two tokens were identical")
	}

	hash := HashToken(token)
	if len(hash) != 32 {
		t.Fatalf("token hash length = %d, want 32", len(hash))
	}
	if !bytes.Equal(hash, HashToken(token)) {
		t.Fatal("token hash is not stable")
	}
	if HashToken("") == nil || len(HashToken("")) != 32 {
		t.Fatal("empty token must still hash deterministically")
	}
}
