package api

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/william22913/clover-be/internal/media"
	"github.com/william22913/clover-be/internal/store"
)

func (h *Handler) listMovies(w http.ResponseWriter, r *http.Request) {
	movies, err := h.store.ListMovies(r.Context())
	if err != nil {
		h.internal(w, r, err)
		return
	}
	items := make([]movieDTO, 0, len(movies))
	for _, m := range movies {
		items = append(items, toMovieDTO(m))
	}
	writeJSON(w, http.StatusOK, movieListBody{Items: items})
}

func (h *Handler) getMovie(w http.ResponseWriter, r *http.Request) {
	movie, err := h.store.GetMovie(r.Context(), r.PathValue("slug"))
	if errors.Is(err, store.ErrNotFound) {
		h.fail(w, r, http.StatusNotFound, codeMovieNotFound, "This film is not available.", nil)
		return
	}
	if err != nil {
		h.internal(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, toMovieDTO(movie))
}

// video streams a film. The route is wrapped in withSession before it runs.
func (h *Handler) video(w http.ResponseWriter, r *http.Request) {
	movie, err := h.store.GetMovie(r.Context(), r.PathValue("slug"))
	if errors.Is(err, store.ErrNotFound) {
		h.fail(w, r, http.StatusNotFound, codeMovieNotFound, "This film is not available.", nil)
		return
	}
	if err != nil {
		h.internal(w, r, err)
		return
	}
	if err := h.media.Video(w, r, movie.VideoFile); err != nil {
		if errors.Is(err, media.ErrUnavailable) {
			h.fail(w, r, http.StatusNotFound, codeVideoUnavailable, "The film could not be loaded.", nil)
			return
		}
		h.internal(w, r, err)
	}
}

func (h *Handler) healthz(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()
	if err := h.store.Ping(ctx); err != nil {
		h.log.ErrorContext(r.Context(), "health check failed", "error", err)
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "unavailable"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// notFound answers any unknown path under /v1/.
func (h *Handler) notFound(w http.ResponseWriter, r *http.Request) {
	h.fail(w, r, http.StatusNotFound, codeNotFound, "Something went wrong. Please try again.", nil)
}
