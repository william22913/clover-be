package store

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
)

// Movie is one catalogue row. Files are filenames, never paths or URLs.
type Movie struct {
	Slug             string
	Title            string
	ShortDescription string
	ReleaseYear      int
	RuntimeSeconds   int
	PosterFile       string
	StillFile        *string
	VideoFile        string
}

// ListMovies returns every published film in catalogue order.
func (s *Store) ListMovies(ctx context.Context) ([]Movie, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT slug, title, short_description, release_year, runtime_seconds,
		       poster_file, still_file, video_file
		FROM movies
		WHERE is_published AND NOT is_deleted
		ORDER BY display_order, id`)
	if err != nil {
		return nil, fmt.Errorf("list movies: %w", err)
	}
	defer rows.Close()

	movies := make([]Movie, 0, 13)
	for rows.Next() {
		var m Movie
		if err := rows.Scan(&m.Slug, &m.Title, &m.ShortDescription, &m.ReleaseYear,
			&m.RuntimeSeconds, &m.PosterFile, &m.StillFile, &m.VideoFile); err != nil {
			return nil, fmt.Errorf("scan movie: %w", err)
		}
		movies = append(movies, m)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate movies: %w", err)
	}
	return movies, nil
}

// GetMovie returns one published film by slug. The predicate matches the
// catalogue index, so an unpublished or withdrawn film is absent everywhere.
func (s *Store) GetMovie(ctx context.Context, slug string) (Movie, error) {
	var m Movie
	err := s.pool.QueryRow(ctx, `
		SELECT slug, title, short_description, release_year, runtime_seconds,
		       poster_file, still_file, video_file
		FROM movies
		WHERE slug = $1 AND is_published AND NOT is_deleted`, slug,
	).Scan(&m.Slug, &m.Title, &m.ShortDescription, &m.ReleaseYear,
		&m.RuntimeSeconds, &m.PosterFile, &m.StillFile, &m.VideoFile)
	if errors.Is(err, pgx.ErrNoRows) {
		return Movie{}, ErrNotFound
	}
	if err != nil {
		return Movie{}, fmt.Errorf("get movie: %w", err)
	}
	return m, nil
}
