// Package seed owns the only writes to movies and accounts: the fixed v1
// catalogue, the demo accounts the test design needs, and the media files they
// point at. It is idempotent and safe to run on every environment start.
package seed

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/william22913/clover-be/internal/accounts"
	"github.com/william22913/clover-be/internal/media"
)

// Film is one seeded catalogue row. Titles and descriptions are copied exactly
// from the approved mock (ui/mock/app.js); durations match the committed files.
type Film struct {
	Slug        string
	Title       string
	Description string
	Duration    int // seconds, 1..1800
}

// Films is the 20-film v1 catalogue in editorial order.
var Films = []Film{
	{"amber-drift", "Amber Drift", "A slow descent into a golden world where every surface is alive.", 600},
	{"motel-sun", "Motel Sun", "A stranded traveler waits out the night in a motel that should have closed years ago.", 90},
	{"long-evening", "The Long Evening", "One last summer evening stretches past midnight, and nobody wants to leave.", 960},
	{"neon-static", "Neon Static", "A city street, a missing signal, and a message that only appears in the rain.", 90},
	{"square-rain", "Square Rain", "On the wettest night of the year, two strangers miss the same train.", 90},
	{"after-midnight", "After Midnight", "The marquee lights stay on long after the last audience has gone home.", 90},
	{"orbit", "Orbit", "Ten minutes outside the station. No tether. One chance to look back.", 600},
	{"docking-day", "Docking Day", "A routine approach becomes anything but routine when the station goes silent.", 90},
	{"breakwater", "Breakwater", "The storm arrives on schedule. The lighthouse keeper does not.", 90},
	{"whitecap", "Whitecap", "A boat, a horizon, and a decision that cannot wait for calm water.", 90},
	{"last-light", "Last Light", "The sea gives up one last hour of sun before the weather closes in.", 90},
	{"the-hollow", "The Hollow", "Something moves between the trees, and the path keeps leading back.", 90},
	{"ember-road", "Ember Road", "A road through the fog, a car with no headlights, and a destination nobody names.", 90},
	{"tall-pines", "Tall Pines", "The forest is quiet. That is how you know someone is listening.", 90},
	{"northern-signal", "Northern Signal", "A green sky over a frozen waterfall, and a signal that answers back.", 90},
	{"green-sky", "Green Sky", "The lights come out early tonight, and they are not where they should be.", 90},
	{"snowbound", "Snowbound", "A cabin at the edge of the aurora, and a visitor who arrives without footprints.", 90},
	{"moonlit-runner", "Moonlit Runner", "A classic car, an open road, and a moon that follows too closely.", 90},
	{"blue-hour", "Blue Hour", "Between sunset and dark, a driver makes a deal with the quiet.", 90},
	{"descent", "Descent", "Below the surface, the light bends and the rules stop applying.", 90},
}

// Account is one seeded sign-in fixture from testcase.md's "Scope and test
// data" table. It must be reproduced exactly, trailing spaces included.
type Account struct {
	Username string
	Email    string
	Password string
}

// Accounts is the seven-account fixture set.
var Accounts = []Account{
	{"alex", "Alex.Rivera@example.com", "Watch1234"},
	{"mira_k9", "mira.k@example.com", "Play2025x"},
	{"abc", "abc@example.com", "Abcdefg1"},
	{strings.Repeat("a", 50), "max.user@example.com", "Abcdefg1"},
	{"minpw", "minpw@example.com", "Abcdefg1"},
	{"maxpw", "maxpw@example.com", "A1" + strings.Repeat("x", 126)},
	{"spacey", "spacey@example.com", "Watch1234 "},
}

// Run validates the seed data, verifies the media files and upserts the rows.
func Run(ctx context.Context, pool *pgxpool.Pool, mediaDir string) error {
	if err := validate(); err != nil {
		return err
	}
	if err := verifyMedia(mediaDir); err != nil {
		return err
	}
	if err := seedFilms(ctx, pool); err != nil {
		return err
	}
	return seedAccounts(ctx, pool)
}

func validate() error {
	if len(Films) != 20 {
		return fmt.Errorf("seed: the v1 catalogue must hold 20 films, got %d", len(Films))
	}
	seenSlugs := map[string]bool{}
	seenOrder := map[int]bool{}
	for i, f := range Films {
		order := i + 1
		if !media.ValidSlug(f.Slug) {
			return fmt.Errorf("seed: film %d has an invalid slug %q", order, f.Slug)
		}
		if seenSlugs[f.Slug] {
			return fmt.Errorf("seed: duplicate slug %q", f.Slug)
		}
		if seenOrder[order] {
			return fmt.Errorf("seed: duplicate catalogue order %d", order)
		}
		if strings.TrimSpace(f.Title) == "" || strings.TrimSpace(f.Description) == "" {
			return fmt.Errorf("seed: film %q needs a title and a description", f.Slug)
		}
		if f.Duration < 1 || f.Duration > 1800 {
			return fmt.Errorf("seed: film %q duration %ds is outside 1..1800", f.Slug, f.Duration)
		}
		seenSlugs[f.Slug] = true
		seenOrder[order] = true
	}

	seenUsernames := map[string]bool{}
	seenEmails := map[string]bool{}
	for _, a := range Accounts {
		ident, err := accounts.ParseIdentifier(a.Username)
		if err != nil || ident.IsEmail {
			return fmt.Errorf("seed: fixture username %q is not a valid username", a.Username)
		}
		if err := accounts.ValidatePassword(a.Password); err != nil {
			return fmt.Errorf("seed: fixture %q has an out-of-policy password: %w", a.Username, err)
		}
		if seenUsernames[strings.ToLower(a.Username)] {
			return fmt.Errorf("seed: duplicate username %q", a.Username)
		}
		if seenEmails[strings.ToLower(a.Email)] {
			return fmt.Errorf("seed: duplicate email %q", a.Email)
		}
		seenUsernames[strings.ToLower(a.Username)] = true
		seenEmails[strings.ToLower(a.Email)] = true
	}
	return nil
}

func verifyMedia(mediaDir string) error {
	for _, f := range Films {
		for _, rel := range []string{posterPath(f.Slug), videoPath(f.Slug)} {
			path, err := media.Resolve(mediaDir, rel)
			if err != nil {
				return fmt.Errorf("seed: %w", err)
			}
			info, err := os.Stat(path)
			if errors.Is(err, os.ErrNotExist) {
				return fmt.Errorf("seed: missing media file %s (posters and videos are committed under media/)", path)
			}
			if err != nil {
				return fmt.Errorf("seed: stat %s: %w", path, err)
			}
			if info.Size() == 0 {
				return fmt.Errorf("seed: media file %s is empty", path)
			}
		}
	}
	return nil
}

func posterPath(slug string) string { return filepath.Join("posters", slug+".webp") }
func videoPath(slug string) string  { return filepath.Join("videos", slug+".mp4") }

func seedFilms(ctx context.Context, pool *pgxpool.Pool) error {
	for i, f := range Films {
		_, err := pool.Exec(ctx,
			`INSERT INTO movies (id, slug, title, description, poster_path, video_path, duration_seconds, catalogue_order)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			 ON CONFLICT (slug) DO UPDATE
			    SET title = EXCLUDED.title,
			        description = EXCLUDED.description,
			        poster_path = EXCLUDED.poster_path,
			        video_path = EXCLUDED.video_path,
			        duration_seconds = EXCLUDED.duration_seconds,
			        catalogue_order = EXCLUDED.catalogue_order,
			        updated_at = now()`,
			uuid.New(), f.Slug, f.Title, f.Description,
			posterPath(f.Slug), videoPath(f.Slug), f.Duration, i+1,
		)
		if err != nil {
			return fmt.Errorf("seed: upsert film %q: %w", f.Slug, err)
		}
	}
	return nil
}

func seedAccounts(ctx context.Context, pool *pgxpool.Pool) error {
	for _, a := range Accounts {
		hash, err := accounts.HashPassword(a.Password)
		if err != nil {
			return fmt.Errorf("seed: hash password for %q: %w", a.Username, err)
		}
		_, err = pool.Exec(ctx,
			`INSERT INTO accounts (id, username, email, password_hash)
			 VALUES ($1, $2, $3, $4)
			 ON CONFLICT (username) DO NOTHING`,
			uuid.New(), strings.ToLower(a.Username), strings.ToLower(a.Email), hash,
		)
		if err != nil {
			return fmt.Errorf("seed: insert account %q: %w", a.Username, err)
		}
	}
	return nil
}
