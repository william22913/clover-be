package seed

import (
	"strings"
	"testing"
)

// TestSeed_TC062_CatalogueIsComplete covers the seed-data requirements: 20
// films, every field present, unique slugs and editorial order, and seven
// fixture accounts with unique identifiers.
func TestSeed_TC062_CatalogueIsComplete(t *testing.T) {
	if err := validate(); err != nil {
		t.Fatalf("seed data invalid: %v", err)
	}
	if len(Accounts) != 7 {
		t.Fatalf("fixture accounts = %d, want 7", len(Accounts))
	}
	usernames := map[string]bool{}
	emails := map[string]bool{}
	for _, a := range Accounts {
		if usernames[strings.ToLower(a.Username)] || emails[strings.ToLower(a.Email)] {
			t.Fatalf("duplicate fixture identifier for %q", a.Username)
		}
		usernames[strings.ToLower(a.Username)] = true
		emails[strings.ToLower(a.Email)] = true
	}
}

// TestSeed_TC063_DurationCeiling covers REQ-029: 30:00 is accepted, 30:01 is
// rejected before it can reach the catalogue.
func TestSeed_TC063_DurationCeiling(t *testing.T) {
	defer func() { Films[0].Duration = 600 }()

	Films[0].Duration = 1800
	if err := validate(); err != nil {
		t.Fatalf("a 30:00 film was rejected: %v", err)
	}
	Films[0].Duration = 1801
	if err := validate(); err == nil {
		t.Fatal("a 30:01 film was accepted")
	}
}

// TestSeed_TC056_MediaNamesMatchEveryFilm keeps the committed file names and
// the catalogue rows in step.
func TestSeed_TC056_MediaNamesMatchEveryFilm(t *testing.T) {
	seen := map[string]bool{}
	for _, f := range Films {
		if got := videoPath(f.Slug); got != "videos/"+f.Slug+".mp4" {
			t.Fatalf("video path for %q = %q", f.Slug, got)
		}
		if got := posterPath(f.Slug); got != "posters/"+f.Slug+".webp" {
			t.Fatalf("poster path for %q = %q", f.Slug, got)
		}
		if seen[f.Title] || strings.TrimSpace(f.Title) == "" {
			t.Fatalf("film title %q is empty or repeated", f.Title)
		}
		seen[f.Title] = true
	}
}
