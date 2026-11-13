package api_test

import (
	"net/http"
	"strings"
	"testing"

	"github.com/william22913/clover-be/internal/store"
)

type movieJSON struct {
	Slug             string  `json:"slug"`
	Title            string  `json:"title"`
	ShortDescription string  `json:"short_description"`
	ReleaseYear      int     `json:"release_year"`
	RuntimeSeconds   int     `json:"runtime_seconds"`
	PosterURL        string  `json:"poster_url"`
	StillURL         *string `json:"still_url"`
}

// TC-001 — an anonymous visitor sees the catalogue with no login.
func TestAPI_TC001_CatalogueIsPublic(t *testing.T) {
	env := newEnv(t)
	res := env.do(http.MethodGet, "/v1/movies", "", nil)
	if res.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", res.StatusCode)
	}
	var body struct {
		Items []movieJSON `json:"items"`
	}
	decodeBody(t, res, &body)
	if len(body.Items) != 2 {
		t.Fatalf("items = %d, want 2", len(body.Items))
	}
	for _, item := range body.Items {
		if item.Title == "" || item.ShortDescription == "" || item.PosterURL == "" {
			t.Fatalf("card fields missing: %+v", item)
		}
	}
}

// TC-002 — every published movie is listed, in catalogue order, with no
// internal id in the payload.
func TestAPI_TC002_CatalogueListsEveryPublishedMovie(t *testing.T) {
	env := newEnv(t)
	res := env.do(http.MethodGet, "/v1/movies", "", nil)
	if res.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", res.StatusCode)
	}
	var raw struct {
		Items []map[string]any `json:"items"`
	}
	decodeBody(t, res, &raw)
	if len(raw.Items) != len(env.store.movies) {
		t.Fatalf("items = %d, want %d", len(raw.Items), len(env.store.movies))
	}
	want := []string{"halcyon", "vesper"}
	for i, item := range raw.Items {
		if item["slug"] != want[i] {
			t.Fatalf("item %d slug = %v, want %s", i, item["slug"], want[i])
		}
		if _, hasID := item["id"]; hasID {
			t.Fatal("payload must not carry an internal id")
		}
	}
}

// TC-003 — there is no pagination or query surface on the catalogue.
func TestAPI_TC003_CatalogueHasNoPagination(t *testing.T) {
	env := newEnv(t)
	res := env.do(http.MethodGet, "/v1/movies?page=2", "", nil)
	if res.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", res.StatusCode)
	}
	var body struct {
		Items    []movieJSON `json:"items"`
		NextPage any         `json:"next_page"`
	}
	decodeBody(t, res, &body)
	if len(body.Items) != 2 {
		t.Fatalf("items = %d, want the full unpaginated set", len(body.Items))
	}
	if body.NextPage != nil {
		t.Fatal("catalogue must not offer paging")
	}
}

// TC-004 — selecting a movie opens exactly that movie.
func TestAPI_TC004_GetMovieBySlug(t *testing.T) {
	env := newEnv(t)
	res := env.do(http.MethodGet, "/v1/movies/vesper", "", nil)
	if res.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", res.StatusCode)
	}
	var movie movieJSON
	decodeBody(t, res, &movie)
	if movie.Slug != "vesper" || movie.Title != "Vesper Light" {
		t.Fatalf("movie = %+v", movie)
	}
	if movie.StillURL == nil || !strings.HasPrefix(*movie.StillURL, "/media/still/") {
		t.Fatalf("still_url = %v", movie.StillURL)
	}
}

// TC-028 — a player URL for a movie that does not exist is a clear 404.
func TestAPI_TC028_UnknownMovieIsMovieNotFound(t *testing.T) {
	env := newEnv(t)
	res := env.do(http.MethodGet, "/v1/movies/unlisted", "", nil)
	if res.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", res.StatusCode)
	}
	var body errorResponse
	decodeBody(t, res, &body)
	if body.Error.Code != "movie_not_found" {
		t.Fatalf("code = %q, want movie_not_found", body.Error.Code)
	}
	if body.Error.RequestID == "" {
		t.Fatal("error body must carry the request id")
	}
	if got := res.Header.Get("X-Request-Id"); got != body.Error.RequestID {
		t.Fatalf("X-Request-Id = %q, body = %q", got, body.Error.RequestID)
	}
}

// TC-027 — a signed-out visitor cannot reach the video stream.
func TestAPI_TC027_VideoLockedWithoutSession(t *testing.T) {
	env := newEnv(t)
	res := env.do(http.MethodGet, "/v1/movies/halcyon/video", "", nil)
	if res.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", res.StatusCode)
	}
	var body errorResponse
	decodeBody(t, res, &body)
	if body.Error.Code != "session_expired" {
		t.Fatalf("code = %q, want session_expired", body.Error.Code)
	}
}

// TC-009 — a signed-in viewer can play: the stream is served, and HEAD works
// too.
func TestAPI_TC009_SignedInViewerStreams(t *testing.T) {
	env := newEnv(t)
	cookie := signUpCookie(t, env, "viewer@example.com", "Passw0rd1")

	res := env.do(http.MethodGet, "/v1/movies/halcyon/video", "", map[string]string{"Cookie": cookieHeader(cookie)})
	if res.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", res.StatusCode)
	}
	if ct := res.Header.Get("Content-Type"); ct != "video/mp4" {
		t.Fatalf("content-type = %q, want video/mp4", ct)
	}
	if res.Header.Get("Accept-Ranges") != "bytes" {
		t.Fatal("Accept-Ranges must advertise byte ranges")
	}

	head := env.do(http.MethodHead, "/v1/movies/halcyon/video", "", map[string]string{"Cookie": cookieHeader(cookie)})
	if head.StatusCode != http.StatusOK {
		t.Fatalf("HEAD status = %d, want 200", head.StatusCode)
	}
}

// TC-021 — playback starts before the whole file has arrived: a byte-range
// request is answered with 206 from that offset instead of re-sending the file.
func TestAPI_TC021_RangeRequestReturns206(t *testing.T) {
	env := newEnv(t)
	cookie := signUpCookie(t, env, "viewer@example.com", "Passw0rd1")

	res := env.do(http.MethodGet, "/v1/movies/halcyon/video", "",
		map[string]string{"Cookie": cookie.String(), "Range": "bytes=2048-4095"})
	if res.StatusCode != http.StatusPartialContent {
		t.Fatalf("status = %d, want 206", res.StatusCode)
	}
	if cr := res.Header.Get("Content-Range"); !strings.HasPrefix(cr, "bytes 2048-4095/4096") {
		t.Fatalf("content-range = %q", cr)
	}
}

// TC-015 — a revoked session cannot fetch video or account data.
func TestAPI_TC015_RevokedSessionRefused(t *testing.T) {
	env := newEnv(t)
	cookie := signUpCookie(t, env, "viewer@example.com", "Passw0rd1")

	out := env.do(http.MethodPost, "/v1/auth/sign-out", "", map[string]string{"Cookie": cookieHeader(cookie)})
	if out.StatusCode != http.StatusNoContent {
		t.Fatalf("sign-out status = %d, want 204", out.StatusCode)
	}

	for _, path := range []string{"/v1/movies/halcyon/video", "/v1/auth/me"} {
		res := env.do(http.MethodGet, path, "", map[string]string{"Cookie": cookieHeader(cookie)})
		if res.StatusCode != http.StatusUnauthorized {
			t.Fatalf("%s status = %d, want 401", path, res.StatusCode)
		}
	}
}

// TC-029 — a row whose file is missing is video_unavailable, not a 500.
func TestAPI_TC029_MissingVideoIsUnavailable(t *testing.T) {
	env := newEnv(t)
	env.store.addMovie(store.Movie{
		Slug: "fathoms", Title: "Nine Fathoms", ShortDescription: "A deep-sea station loses contact.",
		ReleaseYear: 2023, RuntimeSeconds: 6420,
		PosterFile: "p-fathoms.jpg", VideoFile: "missing.mp4",
	})
	cookie := signUpCookie(t, env, "viewer@example.com", "Passw0rd1")

	res := env.do(http.MethodGet, "/v1/movies/fathoms/video", "", map[string]string{"Cookie": cookieHeader(cookie)})
	if res.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", res.StatusCode)
	}
	var body errorResponse
	decodeBody(t, res, &body)
	if body.Error.Code != "video_unavailable" {
		t.Fatalf("code = %q, want video_unavailable", body.Error.Code)
	}
}

// TC-013 — two sign-ins for one account create two live sessions.
func TestAPI_TC013_TwoDevicesStaySignedIn(t *testing.T) {
	env := newEnv(t)
	env.store.addUser("second@example.com", "Second123")

	first := signInCookie(t, env, "second@example.com", "Second123")
	second := signInCookie(t, env, "second@example.com", "Second123")
	if first.Value == second.Value {
		t.Fatal("each sign-in must issue a new token")
	}
	for _, cookie := range []*http.Cookie{first, second} {
		res := env.do(http.MethodGet, "/v1/auth/me", "", map[string]string{"Cookie": cookieHeader(cookie)})
		if res.StatusCode != http.StatusOK {
			t.Fatalf("session status = %d, want 200", res.StatusCode)
		}
	}
}

// TC-014 — signing out clears the cookie and returns to anonymous browsing.
func TestAPI_TC014_SignOutClearsCookie(t *testing.T) {
	env := newEnv(t)
	cookie := signUpCookie(t, env, "viewer@example.com", "Passw0rd1")

	res := env.do(http.MethodPost, "/v1/auth/sign-out", "", map[string]string{"Cookie": cookieHeader(cookie)})
	if res.StatusCode != http.StatusNoContent {
		t.Fatalf("status = %d, want 204", res.StatusCode)
	}
	cleared := sessionCookie(res)
	if cleared == nil || cleared.MaxAge >= 0 {
		t.Fatalf("cookie was not cleared: %+v", cleared)
	}
}

// TC-012 — the cookie carries the 7-day lifetime and the security attributes.
func TestAPI_TC012_SessionCookieLifetime(t *testing.T) {
	env := newEnv(t)
	cookie := signUpCookie(t, env, "viewer@example.com", "Passw0rd1")

	if cookie.MaxAge != 7*24*60*60 {
		t.Fatalf("MaxAge = %d, want 604800", cookie.MaxAge)
	}
	if !cookie.HttpOnly || !cookie.Secure || cookie.SameSite != http.SameSiteLaxMode {
		t.Fatalf("cookie attributes = %+v", cookie)
	}
	if cookie.Path != "/" {
		t.Fatalf("Path = %q, want /", cookie.Path)
	}
}

// healthz answers ok while the database answers.
func TestAPI_HealthzReportsOK(t *testing.T) {
	env := newEnv(t)
	res := env.do(http.MethodGet, "/healthz", "", nil)
	if res.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", res.StatusCode)
	}
}
