package media

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func newTestServer(t *testing.T) (*Server, string) {
	t.Helper()
	root := t.TempDir()
	for _, dir := range []string{"web", "media/video", "media/poster"} {
		if err := os.MkdirAll(filepath.Join(root, dir), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	clip := make([]byte, 4096)
	for i := range clip {
		clip[i] = byte(i)
	}
	if err := os.WriteFile(filepath.Join(root, "media", "video", "alpha.mp4"), clip, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "web", "app.js"), []byte("console.log(1);"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "media", "poster", "p-alpha.jpg"), []byte("jpg"), 0o644); err != nil {
		t.Fatal(err)
	}
	return New(filepath.Join(root, "web"), filepath.Join(root, "media")), root
}

// TC-021 — a range request is answered with 206 from the asked offset, which is
// what lets playback start before the file has fully arrived and lets a seek
// resume rather than restart.
func TestVideo_RangeReturns206(t *testing.T) {
	s, _ := newTestServer(t)
	req := httptest.NewRequest(http.MethodGet, "/v1/movies/alpha/video", nil)
	req.Header.Set("Range", "bytes=1024-2047")
	rec := httptest.NewRecorder()

	if err := s.Video(rec, req, "alpha.mp4"); err != nil {
		t.Fatal(err)
	}
	res := rec.Result()
	defer res.Body.Close()
	if res.StatusCode != http.StatusPartialContent {
		t.Fatalf("status = %d, want 206", res.StatusCode)
	}
	if got := res.Header.Get("Content-Range"); got != "bytes 1024-2047/4096" {
		t.Fatalf("Content-Range = %q", got)
	}
	if got := res.Header.Get("Accept-Ranges"); got != "bytes" {
		t.Fatalf("Accept-Ranges = %q", got)
	}
	if got := res.Header.Get("Content-Type"); got != "video/mp4" {
		t.Fatalf("Content-Type = %q", got)
	}
	if got := res.Header.Get("Cache-Control"); got != "private, no-store" {
		t.Fatalf("Cache-Control = %q", got)
	}
	if rec.Body.Len() != 1024 {
		t.Fatalf("body = %d bytes, want 1024", rec.Body.Len())
	}
}

// HEAD reports the size without sending the body.
func TestVideo_HeadSendsNoBody(t *testing.T) {
	s, _ := newTestServer(t)
	req := httptest.NewRequest(http.MethodHead, "/v1/movies/alpha/video", nil)
	rec := httptest.NewRecorder()

	if err := s.Video(rec, req, "alpha.mp4"); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if got := rec.Header().Get("Content-Length"); got != "4096" {
		t.Fatalf("Content-Length = %q, want 4096", got)
	}
	if rec.Body.Len() != 0 {
		t.Fatalf("body = %d bytes, want none", rec.Body.Len())
	}
}

// TC-029 support — a row whose file is gone is ErrUnavailable, never a crash.
func TestVideo_MissingFileUnavailable(t *testing.T) {
	s, _ := newTestServer(t)
	rec := httptest.NewRecorder()
	err := s.Video(rec, httptest.NewRequest(http.MethodGet, "/v1/movies/gone/video", nil), "gone.mp4")
	if !errors.Is(err, ErrUnavailable) {
		t.Fatalf("err = %v, want ErrUnavailable", err)
	}
}

// A database filename can never escape the video directory.
func TestVideo_RejectsUnsafeFilename(t *testing.T) {
	s, root := newTestServer(t)
	if err := os.WriteFile(filepath.Join(root, "secret.mp4"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"", ".", "..", "../secret.mp4", "video/alpha.mp4", "a\\b.mp4", "a\x00.mp4"} {
		rec := httptest.NewRecorder()
		err := s.Video(rec, httptest.NewRequest(http.MethodGet, "/x", nil), name)
		if !errors.Is(err, ErrUnavailable) {
			t.Fatalf("%q err = %v, want ErrUnavailable", name, err)
		}
		if rec.Body.Len() != 0 {
			t.Fatalf("%q wrote %d bytes", name, rec.Body.Len())
		}
	}
}

// The public asset handlers serve files but never a directory listing.
func TestAssetHandlers_ServeFilesButNotDirectories(t *testing.T) {
	s, _ := newTestServer(t)

	for _, tc := range []struct {
		name    string
		handler http.Handler
		file    string
	}{
		{"static", s.StaticHandler(), "/static/app.js"},
		{"poster", s.PosterHandler(), "/media/poster/p-alpha.jpg"},
	} {
		rec := httptest.NewRecorder()
		tc.handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, tc.file, nil))
		if rec.Code != http.StatusOK || rec.Body.Len() == 0 {
			t.Fatalf("%s file: status = %d, body = %d", tc.name, rec.Code, rec.Body.Len())
		}

		dir := strings.TrimSuffix(tc.file, filepath.Base(tc.file))
		rec = httptest.NewRecorder()
		tc.handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, dir, nil))
		if rec.Code != http.StatusNotFound {
			t.Fatalf("%s directory: status = %d, want 404", tc.name, rec.Code)
		}
	}
}
