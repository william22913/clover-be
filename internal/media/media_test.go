package media

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestResolve_S11_RejectsPathsOutsideTheRoot is the path-containment guard
// (S11): a request may never name a file outside the media root.
func TestResolve_S11_RejectsPathsOutsideTheRoot(t *testing.T) {
	root := t.TempDir()
	inside, err := Resolve(root, filepath.Join("posters", "amber-drift.webp"))
	if err != nil {
		t.Fatalf("inside path rejected: %v", err)
	}
	if !strings.HasPrefix(inside, root+string(os.PathSeparator)) {
		t.Fatalf("resolved path %q escaped %q", inside, root)
	}
	for _, rel := range []string{
		"../etc/passwd",
		"posters/../../etc/passwd",
		"/etc/passwd",
		"",
	} {
		if _, err := Resolve(root, rel); !errors.Is(err, ErrUnsafePath) {
			t.Fatalf("Resolve(%q) = %v, want ErrUnsafePath", rel, err)
		}
	}
}

// TestValidSlug_S11 covers the media-id rule the handlers enforce before any
// database or filesystem access.
func TestValidSlug_S11(t *testing.T) {
	for _, slug := range []string{"amber-drift", "a", strings.Repeat("a", 64)} {
		if !ValidSlug(slug) {
			t.Fatalf("%q rejected", slug)
		}
	}
	for _, slug := range []string{"", "Amber-Drift", "amber drift", "amber/drift", "../amber", strings.Repeat("a", 65)} {
		if ValidSlug(slug) {
			t.Fatalf("%q accepted", slug)
		}
	}
}

// TestServeVideo_TC052_TC055_RangeAndHead covers the transport behaviour the
// player depends on: a ranged request returns 206 with exactly the requested
// bytes, HEAD returns the length with no body, and no response is cacheable.
func TestServeVideo_TC052_TC055_RangeAndHead(t *testing.T) {
	path := writeFile(t, bytes.Repeat([]byte("v"), 1000))

	t.Run("range", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/media/movies/x/video", nil)
		req.Header.Set("Range", "bytes=100-199")
		rec := httptest.NewRecorder()
		if err := ServeVideo(rec, req, path); err != nil {
			t.Fatalf("ServeVideo: %v", err)
		}
		if rec.Code != http.StatusPartialContent {
			t.Fatalf("status = %d, want 206", rec.Code)
		}
		if got := rec.Header().Get("Content-Range"); got != "bytes 100-199/1000" {
			t.Fatalf("Content-Range = %q", got)
		}
		if rec.Body.Len() != 100 {
			t.Fatalf("body = %d bytes, want 100", rec.Body.Len())
		}
		if got := rec.Header().Get("Accept-Ranges"); got != "bytes" {
			t.Fatalf("Accept-Ranges = %q", got)
		}
		if got := rec.Header().Get("Cache-Control"); got != "private, no-store" {
			t.Fatalf("Cache-Control = %q", got)
		}
		if got := rec.Header().Get("Vary"); got != "Cookie" {
			t.Fatalf("Vary = %q", got)
		}
		if got := rec.Header().Get("Content-Type"); got != "video/mp4" {
			t.Fatalf("Content-Type = %q", got)
		}
	})

	t.Run("head", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodHead, "/v1/media/movies/x/video", nil)
		rec := httptest.NewRecorder()
		if err := ServeVideo(rec, req, path); err != nil {
			t.Fatalf("ServeVideo: %v", err)
		}
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200", rec.Code)
		}
		if got := rec.Header().Get("Content-Length"); got != "1000" {
			t.Fatalf("Content-Length = %q", got)
		}
		if rec.Body.Len() != 0 {
			t.Fatalf("HEAD returned %d body bytes", rec.Body.Len())
		}
	})
}

// TestServePoster_IsPubliclyCacheable covers the poster cache policy.
func TestServePoster_IsPubliclyCacheable(t *testing.T) {
	path := writeFile(t, []byte("RIFFxxxxWEBP"))
	req := httptest.NewRequest(http.MethodGet, "/v1/media/movies/x/poster", nil)
	rec := httptest.NewRecorder()
	if err := ServePoster(rec, req, path); err != nil {
		t.Fatalf("ServePoster: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	if got := rec.Header().Get("Content-Type"); got != "image/webp" {
		t.Fatalf("Content-Type = %q", got)
	}
	if got := rec.Header().Get("Cache-Control"); got != "public, max-age=86400" {
		t.Fatalf("Cache-Control = %q", got)
	}
}

// TestServeVideo_RejectsADirectory keeps a misconfigured row from streaming a
// listing.
func TestServeVideo_RejectsADirectory(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/v1/media/movies/x/video", nil)
	rec := httptest.NewRecorder()
	if err := ServeVideo(rec, req, t.TempDir()); !errors.Is(err, ErrUnsafePath) {
		t.Fatalf("err = %v, want ErrUnsafePath", err)
	}
}

func writeFile(t *testing.T, body []byte) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "clip.mp4")
	if err := os.WriteFile(path, body, 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}
	return path
}
