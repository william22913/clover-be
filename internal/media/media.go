// Package media is the only place that touches the filesystem for user-facing
// assets: the session-gated video stream and the public static files.
package media

import (
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// ErrUnavailable means the row points at a file that is missing or unreadable.
var ErrUnavailable = errors.New("media unavailable")

// Server serves files from the web directory and the media directory.
type Server struct {
	webDir   string
	mediaDir string
}

// New builds a server. webDir holds the UI; mediaDir holds poster/, still/ and
// video/.
func New(webDir, mediaDir string) *Server {
	return &Server{webDir: webDir, mediaDir: mediaDir}
}

// IndexPath is the shell document every UI route serves.
func (s *Server) IndexPath() string { return filepath.Join(s.webDir, "index.html") }

// Video streams one film file with Range support. The filename comes from a
// database row, never from the request URL.
func (s *Server) Video(w http.ResponseWriter, r *http.Request, filename string) error {
	if filename == "" || filename == "." || filename == ".." ||
		strings.ContainsAny(filename, "/\\") || strings.ContainsRune(filename, 0) {
		return ErrUnavailable
	}
	f, err := os.Open(filepath.Join(s.mediaDir, "video", filename))
	if err != nil {
		return ErrUnavailable
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil || info.IsDir() {
		return ErrUnavailable
	}

	w.Header().Set("Content-Type", "video/mp4")
	w.Header().Set("Content-Disposition", "inline")
	w.Header().Set("Cache-Control", "private, no-store")
	http.ServeContent(w, r, filename, info.ModTime(), f)
	return nil
}

// StaticHandler serves the UI files under /static/.
func (s *Server) StaticHandler() http.Handler {
	return http.StripPrefix("/static/", noIndex(http.FileServer(http.Dir(s.webDir))))
}

// PosterHandler serves posters under /media/poster/.
func (s *Server) PosterHandler() http.Handler {
	return http.StripPrefix("/media/poster/", noIndex(http.FileServer(http.Dir(filepath.Join(s.mediaDir, "poster")))))
}

// StillHandler serves stills under /media/still/.
func (s *Server) StillHandler() http.Handler {
	return http.StripPrefix("/media/still/", noIndex(http.FileServer(http.Dir(filepath.Join(s.mediaDir, "still")))))
}

// noIndex refuses directory requests so no listing is ever produced.
func noIndex(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/") || r.URL.Path == "" {
			http.NotFound(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
}
