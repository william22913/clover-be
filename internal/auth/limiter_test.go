package auth

import (
	"testing"
	"time"
)

// The limiter allows limit attempts per key per fixed window and resets exactly
// at the window edge.
func TestLimiter_FixedWindow(t *testing.T) {
	l := NewLimiter(2, int64(time.Minute))
	now := int64(0)
	l.now = func() int64 { return now }

	if !l.Allow("10.0.0.1") || !l.Allow("10.0.0.1") {
		t.Fatal("the first two attempts must be allowed")
	}
	if l.Allow("10.0.0.1") {
		t.Fatal("the third attempt in the window must be refused")
	}
	if !l.Allow("10.0.0.2") {
		t.Fatal("keys are independent")
	}

	now = int64(time.Minute)
	if !l.Allow("10.0.0.1") {
		t.Fatal("a new window must reset the budget")
	}
}

// The sweep keeps a growing key set bounded without dropping live entries.
func TestLimiter_SweepDropsOnlyExpiredKeys(t *testing.T) {
	l := NewLimiter(1, int64(time.Minute))
	now := int64(0)
	l.now = func() int64 { return now }

	for i := 0; i < 10001; i++ {
		l.Allow(string(rune('a'+i%26)) + string(rune(i)))
	}
	now = int64(2 * time.Minute)
	l.Allow("fresh-key")
	if len(l.entries) > 2 {
		t.Fatalf("entries = %d, want the expired keys swept", len(l.entries))
	}
}
