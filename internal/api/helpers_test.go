package api_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"golang.org/x/crypto/bcrypt"

	"github.com/william22913/clover-be/internal/api"
	"github.com/william22913/clover-be/internal/media"
	"github.com/william22913/clover-be/internal/store"
)

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

// fakeStore is the persistence surface the handlers see, without PostgreSQL.
type fakeStore struct {
	mu       sync.Mutex
	movies   []store.Movie
	users    map[string]store.User // key: lower(email)
	sessions map[string]store.SessionUser
	nextID   int64
}

func newFakeStore() *fakeStore {
	return &fakeStore{
		movies: []store.Movie{
			{
				Slug: "halcyon", Title: "Halcyon Bay",
				ShortDescription: "A marine biologist comes back to the town that took her brother, and finds the tide still keeping his secrets.",
				ReleaseYear:      2023, RuntimeSeconds: 7560,
				PosterFile: "p-halcyon.jpg", StillFile: strptr("still-halcyon.jpg"), VideoFile: "clover-sample.mp4",
			},
			{
				Slug: "vesper", Title: "Vesper Light",
				ShortDescription: "Two strangers miss the same last ferry and spend one night deciding whether to miss it on purpose.",
				ReleaseYear:      2022, RuntimeSeconds: 6660,
				PosterFile: "p-vesper.jpg", StillFile: strptr("still-vesper.jpg"), VideoFile: "clover-sample.mp4",
			},
		},
		users:    map[string]store.User{},
		sessions: map[string]store.SessionUser{},
		nextID:   1,
	}
}

func strptr(s string) *string { return &s }

func (f *fakeStore) addUser(email, password string) store.User {
	f.mu.Lock()
	defer f.mu.Unlock()
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	if err != nil {
		panic(err)
	}
	u := store.User{ID: f.nextID, Email: email, PasswordHash: string(hash)}
	f.nextID++
	f.users[strings.ToLower(email)] = u
	return u
}

func (f *fakeStore) ListMovies(context.Context) ([]store.Movie, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([]store.Movie, len(f.movies))
	copy(out, f.movies)
	return out, nil
}

func (f *fakeStore) GetMovie(_ context.Context, slug string) (store.Movie, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, m := range f.movies {
		if m.Slug == slug {
			return m, nil
		}
	}
	return store.Movie{}, store.ErrNotFound
}

func (f *fakeStore) CreateUserAndSession(_ context.Context, email, passwordHash, tokenHash string) (int64, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	key := strings.ToLower(email)
	if _, ok := f.users[key]; ok {
		return 0, store.ErrEmailTaken
	}
	u := store.User{ID: f.nextID, Email: email, PasswordHash: passwordHash}
	f.nextID++
	f.users[key] = u
	f.sessions[tokenHash] = store.SessionUser{UserID: u.ID, Email: u.Email}
	return u.ID, nil
}

func (f *fakeStore) UserByEmail(_ context.Context, email string) (store.User, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	u, ok := f.users[strings.ToLower(email)]
	if !ok {
		return store.User{}, store.ErrNotFound
	}
	return u, nil
}

func (f *fakeStore) CreateSession(_ context.Context, userID int64, tokenHash, revokeTokenHash string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if revokeTokenHash != "" {
		delete(f.sessions, revokeTokenHash)
	}
	for _, u := range f.users {
		if u.ID == userID {
			f.sessions[tokenHash] = store.SessionUser{UserID: u.ID, Email: u.Email}
			return nil
		}
	}
	return fmt.Errorf("user %d not found", userID)
}

func (f *fakeStore) ResolveSession(_ context.Context, tokenHash string) (store.SessionUser, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	u, ok := f.sessions[tokenHash]
	if !ok {
		return store.SessionUser{}, store.ErrNotFound
	}
	return u, nil
}

func (f *fakeStore) RevokeSession(_ context.Context, tokenHash string) (int64, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if _, ok := f.sessions[tokenHash]; ok {
		delete(f.sessions, tokenHash)
		return 1, nil
	}
	return 0, nil
}

func (f *fakeStore) Ping(context.Context) error { return nil }

// addMovie appends one fixture row, used for the unavailable-video case.
func (f *fakeStore) addMovie(m store.Movie) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.movies = append(f.movies, m)
}

// testEnv is one handler wired to a fake store and a temporary media tree.
type testEnv struct {
	t      *testing.T
	store  *fakeStore
	server *httptest.Server
}

func newEnv(t *testing.T) *testEnv { return newEnvWithLogger(t, testLogger()) }

func newEnvWithLogger(t *testing.T, logger *slog.Logger) *testEnv {
	t.Helper()
	root := t.TempDir()
	for _, dir := range []string{"web", "web/assets", "media/poster", "media/still", "media/video"} {
		if err := os.MkdirAll(filepath.Join(root, dir), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(root, "web", "index.html"), []byte("<!doctype html><title>Clover</title>"), 0o644); err != nil {
		t.Fatal(err)
	}
	clip := make([]byte, 4096)
	for i := range clip {
		clip[i] = byte(i)
	}
	if err := os.WriteFile(filepath.Join(root, "media", "video", "clover-sample.mp4"), clip, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "media", "poster", "p-halcyon.jpg"), []byte("jpg"), 0o644); err != nil {
		t.Fatal(err)
	}

	fs := newFakeStore()
	handler := api.New(api.Options{
		Store:  fs,
		Media:  media.New(filepath.Join(root, "web"), filepath.Join(root, "media")),
		Logger: logger,
	})
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return &testEnv{t: t, store: fs, server: srv}
}

func (e *testEnv) do(method, path string, body string, headers map[string]string) *http.Response {
	e.t.Helper()
	var reader *strings.Reader
	if body == "" {
		reader = strings.NewReader("")
	} else {
		reader = strings.NewReader(body)
	}
	req, err := http.NewRequest(method, e.server.URL+path, reader)
	if err != nil {
		e.t.Fatal(err)
	}
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	for k, v := range headers {
		if k == "Content-Type" {
			req.Header.Set("Content-Type", v)
			continue
		}
		req.Header.Set(k, v)
	}
	res, err := e.server.Client().Do(req)
	if err != nil {
		e.t.Fatal(err)
	}
	e.t.Cleanup(func() {
		res.Body.Close()
	})
	return res
}

func decodeBody(t *testing.T, res *http.Response, dst any) {
	t.Helper()
	defer res.Body.Close()
	if err := json.NewDecoder(res.Body).Decode(dst); err != nil {
		t.Fatalf("decode body: %v", err)
	}
}

type errorResponse struct {
	Error struct {
		Code      string            `json:"code"`
		Message   string            `json:"message"`
		RequestID string            `json:"request_id"`
		Fields    map[string]string `json:"fields"`
	} `json:"error"`
}

func sessionCookie(res *http.Response) *http.Cookie {
	for _, c := range res.Cookies() {
		if c.Name == "clover_session" {
			return c
		}
	}
	return nil
}

// cookieHeader is the request Cookie header for one parsed session cookie.
func cookieHeader(c *http.Cookie) string { return c.Name + "=" + c.Value }

func jsonBody(fields map[string]any) string {
	b, _ := json.Marshal(fields)
	return string(b)
}
