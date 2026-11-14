package httpapi

import (
	"errors"
	"net/http"

	"github.com/william22913/clover-be/internal/media"
	"github.com/william22913/clover-be/internal/movies"
)

// moviePayload is the catalogue item shape from architecture.md.
type moviePayload struct {
	ID              string `json:"id"`
	Title           string `json:"title"`
	Description     string `json:"description"`
	PosterURL       string `json:"posterUrl"`
	VideoURL        string `json:"videoUrl"`
	DurationSeconds int    `json:"durationSeconds"`
	Position        int    `json:"position"`
}

type listMoviesResponse struct {
	Movies []moviePayload `json:"movies"`
}

type getMovieResponse struct {
	Movie moviePayload `json:"movie"`
}

func toPayload(m movies.Movie) moviePayload {
	return moviePayload{
		ID:              m.Slug,
		Title:           m.Title,
		Description:     m.Description,
		PosterURL:       apiPrefix + "/media/movies/" + m.Slug + "/poster",
		VideoURL:        apiPrefix + "/media/movies/" + m.Slug + "/video",
		DurationSeconds: m.DurationSeconds,
		Position:        m.Position,
	}
}

// listMovies returns the whole catalogue in editorial order. It is public and
// never empty-shaped: an empty catalogue is `{"movies":[]}`.
func (s *Server) listMovies(w http.ResponseWriter, r *http.Request) {
	if !s.authorize(w, r, PermCatalogView) {
		return
	}
	rows, err := movies.List(r.Context(), s.pool)
	if err != nil {
		s.writeInternal(w, r, err)
		return
	}
	out := make([]moviePayload, 0, len(rows))
	for _, m := range rows {
		out = append(out, toPayload(m))
	}
	w.Header().Set("Cache-Control", "public, max-age=60")
	writeJSON(w, http.StatusOK, listMoviesResponse{Movies: out})
}

// getMovie returns one film by its public slug.
func (s *Server) getMovie(w http.ResponseWriter, r *http.Request) {
	if !s.authorize(w, r, PermCatalogView) {
		return
	}
	slug := r.PathValue("movieId")
	if !media.ValidSlug(slug) {
		s.writeError(w, r, http.StatusNotFound, CodeMovieNotFound, nil)
		return
	}
	movie, err := movies.Get(r.Context(), s.pool, slug)
	if errors.Is(err, movies.ErrNotFound) {
		s.writeError(w, r, http.StatusNotFound, CodeMovieNotFound, nil)
		return
	}
	if err != nil {
		s.writeInternal(w, r, err)
		return
	}
	w.Header().Set("Cache-Control", "public, max-age=60")
	writeJSON(w, http.StatusOK, getMovieResponse{Movie: toPayload(movie)})
}

// getPoster streams a film's poster image. The path comes from the movie row,
// never from the request (S11).
func (s *Server) getPoster(w http.ResponseWriter, r *http.Request) {
	if !s.authorize(w, r, PermCatalogView) {
		return
	}
	slug := r.PathValue("movieId")
	if !media.ValidSlug(slug) {
		s.writeError(w, r, http.StatusNotFound, CodeMovieNotFound, nil)
		return
	}
	movie, err := movies.Get(r.Context(), s.pool, slug)
	if errors.Is(err, movies.ErrNotFound) {
		s.writeError(w, r, http.StatusNotFound, CodeMovieNotFound, nil)
		return
	}
	if err != nil {
		s.writeInternal(w, r, err)
		return
	}
	path, err := media.Resolve(s.mediaDir, movie.PosterPath)
	if err != nil {
		s.writeInternal(w, r, err)
		return
	}
	if err := media.ServePoster(w, r, path); err != nil {
		s.writeInternal(w, r, err)
	}
}

// getVideo streams the film with Range support. Without a valid session it
// answers 401 before a single byte of media is written (S1, REQ-011).
func (s *Server) getVideo(w http.ResponseWriter, r *http.Request) {
	id := identityFrom(r.Context())
	if !id.Valid || !id.Role.Allows(PermPlaybackWatch) {
		s.metrics.ObserveVideo("unauthorized")
		s.authorize(w, r, PermPlaybackWatch)
		return
	}

	slug := r.PathValue("movieId")
	if !media.ValidSlug(slug) {
		s.metrics.ObserveVideo("not_found")
		s.writeError(w, r, http.StatusNotFound, CodeMovieNotFound, nil)
		return
	}
	movie, err := movies.Get(r.Context(), s.pool, slug)
	if errors.Is(err, movies.ErrNotFound) {
		s.metrics.ObserveVideo("not_found")
		s.writeError(w, r, http.StatusNotFound, CodeMovieNotFound, nil)
		return
	}
	if err != nil {
		s.writeInternal(w, r, err)
		return
	}
	path, err := media.Resolve(s.mediaDir, movie.VideoPath)
	if err != nil {
		s.writeInternal(w, r, err)
		return
	}
	s.metrics.ObserveVideo("ok")
	if err := media.ServeVideo(w, r, path); err != nil {
		s.writeInternal(w, r, err)
	}
}
