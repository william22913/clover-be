// Package movies owns the read-only seeded catalogue.
package movies

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ErrNotFound means no movie has the requested slug.
var ErrNotFound = errors.New("movies: not found")

// Movie is one catalogue entry.
type Movie struct {
	ID              uuid.UUID
	Slug            string
	Title           string
	Description     string
	PosterPath      string
	VideoPath       string
	DurationSeconds int
	Position        int
}

// List returns every movie in editorial order (REQ-002).
func List(ctx context.Context, pool *pgxpool.Pool) ([]Movie, error) {
	rows, err := pool.Query(ctx,
		`SELECT id, slug, title, description, poster_path, video_path, duration_seconds, catalogue_order
		   FROM movies
		  ORDER BY catalogue_order`)
	if err != nil {
		return nil, fmt.Errorf("movies: list: %w", err)
	}
	defer rows.Close()

	out := make([]Movie, 0, 20)
	for rows.Next() {
		var m Movie
		if err := rows.Scan(&m.ID, &m.Slug, &m.Title, &m.Description,
			&m.PosterPath, &m.VideoPath, &m.DurationSeconds, &m.Position); err != nil {
			return nil, fmt.Errorf("movies: scan: %w", err)
		}
		out = append(out, m)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("movies: list: %w", err)
	}
	return out, nil
}

// Get returns the movie with the given public slug.
func Get(ctx context.Context, pool *pgxpool.Pool, slug string) (Movie, error) {
	var m Movie
	err := pool.QueryRow(ctx,
		`SELECT id, slug, title, description, poster_path, video_path, duration_seconds, catalogue_order
		   FROM movies
		  WHERE slug = $1`, slug,
	).Scan(&m.ID, &m.Slug, &m.Title, &m.Description, &m.PosterPath, &m.VideoPath, &m.DurationSeconds, &m.Position)
	if errors.Is(err, pgx.ErrNoRows) {
		return Movie{}, ErrNotFound
	}
	if err != nil {
		return Movie{}, fmt.Errorf("movies: get: %w", err)
	}
	return m, nil
}
