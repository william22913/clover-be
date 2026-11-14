package config

import "testing"

func TestParseTrustedProxies(t *testing.T) {
	prefixes, err := parseTrustedProxies(" 10.0.0.0/8 , 203.0.113.7 ,")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(prefixes) != 2 {
		t.Fatalf("parsed %d prefixes, want 2", len(prefixes))
	}
	if got := prefixes[0].String(); got != "10.0.0.0/8" {
		t.Fatalf("first = %q", got)
	}
	// A bare address becomes a single-host prefix.
	if got := prefixes[1].String(); got != "203.0.113.7/32" {
		t.Fatalf("second = %q", got)
	}

	if _, err := parseTrustedProxies("not-an-address"); err == nil {
		t.Fatal("an invalid entry was accepted")
	}
	if prefixes, err := parseTrustedProxies(""); err != nil || len(prefixes) != 0 {
		t.Fatalf("empty input: %v, %d prefixes", err, len(prefixes))
	}
}
