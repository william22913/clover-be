package store

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/william22913/clover-be/seed"
)

// fakeHash has the bcrypt shape the users CHECK constraint requires, without
// paying for a real hash at cost 12 in every row.
var fakeHash = "$2a$12$" + strings.Repeat("x", 53)

// newTestStore is the opt-in database fixture. TEST_DATABASE_URL must allow
// migrations, ownership changes and TRUNCATE (the development superuser).
func newTestStore(t *testing.T) *Store {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL not set: database-backed store tests skipped")
	}
	ctx := context.Background()
	if err := Migrate(ctx, dsn); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	s, err := Open(ctx, dsn)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(s.Close)
	if _, err := s.pool.Exec(ctx, `TRUNCATE sessions, users, movies RESTART IDENTITY CASCADE`); err != nil {
		t.Fatalf("truncate: %v", err)
	}
	return s
}

// The migration runner can run again on an already-migrated database without
// reapplying or failing.
func TestMigrate_IsIdempotent(t *testing.T) {
	s := newTestStore(t)
	dsn := os.Getenv("TEST_DATABASE_URL")
	if err := Migrate(context.Background(), dsn); err != nil {
		t.Fatalf("second migrate: %v", err)
	}
	var applied int
	if err := s.pool.QueryRow(context.Background(),
		`SELECT count(*) FROM schema_migrations`).Scan(&applied); err != nil {
		t.Fatal(err)
	}
	if applied == 0 {
		t.Fatal("no migration recorded")
	}
}

// TC-002 — the seed lists all 13 published films, and starting twice changes
// nothing.
func TestSeed_IsIdempotentAndListsEveryMovie(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()
	body, err := seed.FS.ReadFile("movies.sql")
	if err != nil {
		t.Fatal(err)
	}
	if err := Seed(ctx, s, string(body)); err != nil {
		t.Fatalf("first seed: %v", err)
	}
	first, err := s.ListMovies(ctx)
	if err != nil {
		t.Fatal(err)
	}
	var before, after string
	if err := s.pool.QueryRow(ctx,
		`SELECT string_agg(slug || ':' || updated_at::text, ',' ORDER BY id) FROM movies`).Scan(&before); err != nil {
		t.Fatal(err)
	}

	if err := Seed(ctx, s, string(body)); err != nil {
		t.Fatalf("second seed: %v", err)
	}
	second, err := s.ListMovies(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.pool.QueryRow(ctx,
		`SELECT string_agg(slug || ':' || updated_at::text, ',' ORDER BY id) FROM movies`).Scan(&after); err != nil {
		t.Fatal(err)
	}

	if len(first) != 13 || len(second) != 13 {
		t.Fatalf("rows = %d then %d, want 13 both times", len(first), len(second))
	}
	if first[0].Slug != "halcyon" {
		t.Fatalf("first catalogue row = %q, want halcyon", first[0].Slug)
	}
	if before != after {
		t.Fatal("re-seeding rewrote rows; the upsert must leave them untouched")
	}
}

// Unpublished and withdrawn films are absent from both the catalogue and the
// by-slug lookup.
func TestMovies_OnlyPublishedAreVisible(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()
	for _, row := range []struct {
		slug        string
		published   bool
		deleted     bool
		orderByName int
	}{
		{"shown", true, false, 1},
		{"hidden", false, false, 2},
		{"withdrawn", true, true, 3},
	} {
		_, err := s.pool.Exec(ctx, `
			INSERT INTO movies (slug, title, short_description, release_year, runtime_seconds,
			                    poster_file, video_file, display_order, is_published,
			                    is_deleted, deleted_at)
			VALUES ($1, $2, 'A test film.', 2024, 6000, 'p-test.jpg', 'test.mp4', $3, $4,
			        $5, CASE WHEN $5 THEN now() ELSE NULL END)`,
			row.slug, "Title "+row.slug, row.orderByName, row.published, row.deleted)
		if err != nil {
			t.Fatalf("insert %s: %v", row.slug, err)
		}
	}

	movies, err := s.ListMovies(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(movies) != 1 || movies[0].Slug != "shown" {
		t.Fatalf("movies = %+v, want only shown", movies)
	}
	if _, err := s.GetMovie(ctx, "hidden"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("hidden movie err = %v, want ErrNotFound", err)
	}
	if _, err := s.GetMovie(ctx, "shown"); err != nil {
		t.Fatalf("shown movie err = %v", err)
	}
}

// TC-040, TC-041 — uniqueness is case-insensitive and the lookup is too, both
// enforced by the database index.
func TestUsers_DuplicateEmailIsCaseInsensitive(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	if _, err := s.CreateUserAndSession(ctx, "ann@example.com", fakeHash, strings.Repeat("a", 64)); err != nil {
		t.Fatalf("first sign-up: %v", err)
	}
	if _, err := s.CreateUserAndSession(ctx, "Ann@Example.com", fakeHash, strings.Repeat("b", 64)); !errors.Is(err, ErrEmailTaken) {
		t.Fatalf("duplicate err = %v, want ErrEmailTaken", err)
	}

	u, err := s.UserByEmail(ctx, "ANN@EXAMPLE.COM")
	if err != nil {
		t.Fatalf("case-insensitive lookup: %v", err)
	}
	if u.Email != "ann@example.com" {
		t.Fatalf("email = %q", u.Email)
	}
	if _, err := s.UserByEmail(ctx, "nobody@example.com"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("unknown email err = %v, want ErrNotFound", err)
	}
}

// TC-012 — a session is accepted before the 7-day lifetime and refused after
// it; the boundary is the database clock, not application code.
func TestSessions_TC012_ExpiryBoundary(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()
	tokenHash := strings.Repeat("c", 64)
	if _, err := s.CreateUserAndSession(ctx, "viewer@example.com", fakeHash, tokenHash); err != nil {
		t.Fatal(err)
	}

	// Just before the lifetime: still inside the 7 days.
	if _, err := s.pool.Exec(ctx,
		`UPDATE sessions SET created_at = now() - interval '7 days' + interval '1 minute',
		 expires_at = now() + interval '1 minute' WHERE token_hash = $1`, tokenHash); err != nil {
		t.Fatal(err)
	}
	if _, err := s.ResolveSession(ctx, tokenHash); err != nil {
		t.Fatalf("before expiry: %v", err)
	}

	// Just after the lifetime.
	if _, err := s.pool.Exec(ctx,
		`UPDATE sessions SET created_at = now() - interval '7 days' - interval '1 minute',
		 expires_at = now() - interval '1 minute' WHERE token_hash = $1`, tokenHash); err != nil {
		t.Fatal(err)
	}
	if _, err := s.ResolveSession(ctx, tokenHash); !errors.Is(err, ErrNotFound) {
		t.Fatalf("after expiry err = %v, want ErrNotFound", err)
	}
}

// TC-015 — a revoked session is refused, and signing out twice is harmless.
func TestSessions_TC015_RevokedIsRefused(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()
	tokenHash := strings.Repeat("d", 64)
	if _, err := s.CreateUserAndSession(ctx, "viewer@example.com", fakeHash, tokenHash); err != nil {
		t.Fatal(err)
	}

	revoked, err := s.RevokeSession(ctx, tokenHash)
	if err != nil || revoked != 1 {
		t.Fatalf("revoke = %d, %v; want 1, nil", revoked, err)
	}
	if _, err := s.ResolveSession(ctx, tokenHash); !errors.Is(err, ErrNotFound) {
		t.Fatalf("revoked session err = %v, want ErrNotFound", err)
	}
	if again, err := s.RevokeSession(ctx, tokenHash); err != nil || again != 0 {
		t.Fatalf("second revoke = %d, %v; want 0, nil", again, err)
	}
}

// TC-013 — signing in on a second device leaves the first session live, and
// signing in with a presented cookie revokes only that cookie's session.
func TestSessions_TC013_MultipleDevices(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()
	first := strings.Repeat("e", 64)
	userID, err := s.CreateUserAndSession(ctx, "viewer@example.com", fakeHash, first)
	if err != nil {
		t.Fatal(err)
	}

	second := strings.Repeat("f", 64)
	if err := s.CreateSession(ctx, userID, second, ""); err != nil {
		t.Fatal(err)
	}
	for _, token := range []string{first, second} {
		if _, err := s.ResolveSession(ctx, token); err != nil {
			t.Fatalf("session %s: %v", token[:1], err)
		}
	}

	third := strings.Repeat("1", 64)
	if err := s.CreateSession(ctx, userID, third, second); err != nil {
		t.Fatal(err)
	}
	if _, err := s.ResolveSession(ctx, second); !errors.Is(err, ErrNotFound) {
		t.Fatalf("presented cookie's session should be revoked, got %v", err)
	}
	if _, err := s.ResolveSession(ctx, first); err != nil {
		t.Fatalf("other device must stay signed in: %v", err)
	}
}

// Security design — over-privileged credential: the runtime role can read and
// write rows but cannot create tables.
func TestMigrations_AppRoleCannotCreateTables(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()
	conn, err := s.pool.Acquire(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Release()

	if _, err := conn.Exec(ctx, "SET ROLE clover_app"); err != nil {
		t.Skipf("cannot assume clover_app with this DSN: %v", err)
	}
	if _, err := conn.Exec(ctx, "SELECT count(*) FROM movies"); err != nil {
		t.Fatalf("clover_app must read the catalogue: %v", err)
	}
	if _, err := conn.Exec(ctx, "CREATE TABLE should_not_exist (id integer)"); err == nil {
		t.Fatal("clover_app created a table; the runtime role is over-privileged")
	}
}

// Expired rows are purged; live ones survive.
func TestPurgeSessions_RemovesExpiredOnly(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()
	live := strings.Repeat("a", 64)
	expired := strings.Repeat("b", 64)
	userID, err := s.CreateUserAndSession(ctx, "viewer@example.com", fakeHash, live)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.CreateSession(ctx, userID, expired, ""); err != nil {
		t.Fatal(err)
	}
	if _, err := s.pool.Exec(ctx,
		`UPDATE sessions SET created_at = now() - interval '8 days',
		 expires_at = now() - interval '1 day' WHERE token_hash = $1`, expired); err != nil {
		t.Fatal(err)
	}

	purged, err := s.PurgeSessions(ctx)
	if err != nil || purged != 1 {
		t.Fatalf("purge = %d, %v; want 1, nil", purged, err)
	}
	if _, err := s.ResolveSession(ctx, live); err != nil {
		t.Fatalf("live session: %v", err)
	}
}
