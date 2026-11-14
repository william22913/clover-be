// Package media serves the seeded poster images and video files. File paths
// come from the movie row, never from the request, and every resolved path is
// checked to stay inside the media root.
package media

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// slugPattern is the media id rule (S11). Movie slugs come from the database,
// but the pattern is still enforced before a path is built.
var slugPattern = regexp.MustCompile(`^[a-z0-9-]{1,64}$`)

// ErrUnsafePath means a resolved path escaped the media root or the id is
// malformed.
var ErrUnsafePath = errors.New("media: unsafe path")

// ValidSlug reports whether slug is a well-formed media id.
func ValidSlug(slug string) bool {
	return slugPattern.MatchString(slug)
}

// Resolve joins rel under mediaDir and guarantees the result stays inside it.
func Resolve(mediaDir, rel string) (string, error) {
	if rel == "" || filepath.IsAbs(rel) {
		return "", ErrUnsafePath
	}
	root, err := filepath.Abs(mediaDir)
	if err != nil {
		return "", fmt.Errorf("media: resolve root: %w", err)
	}
	target, err := filepath.Abs(filepath.Join(root, rel))
	if err != nil {
		return "", fmt.Errorf("media: resolve path: %w", err)
	}
	if target != root && !strings.HasPrefix(target, root+string(os.PathSeparator)) {
		return "", ErrUnsafePath
	}
	return target, nil
}

// ServePoster writes a poster image with its public cache headers.
func ServePoster(w http.ResponseWriter, r *http.Request, path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return err
	}
	if info.IsDir() {
		return ErrUnsafePath
	}

	w.Header().Set("Content-Type", "image/webp")
	w.Header().Set("Cache-Control", "public, max-age=86400")
	http.ServeContent(w, r, "", info.ModTime(), f)
	return nil
}

// ServeVideo streams a video file with Range support. Responses are private
// and never cached (S1, REQ-021); http.ServeContent answers HEAD, 206 and
// conditional requests with constant memory.
func ServeVideo(w http.ResponseWriter, r *http.Request, path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return err
	}
	if info.IsDir() {
		return ErrUnsafePath
	}

	w.Header().Set("Content-Type", "video/mp4")
	w.Header().Set("Accept-Ranges", "bytes")
	w.Header().Set("Cache-Control", "private, no-store")
	w.Header().Add("Vary", "Cookie")
	http.ServeContent(w, r, "", info.ModTime(), f)
	return nil
}
