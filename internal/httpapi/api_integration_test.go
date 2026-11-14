package httpapi

// These tests need a real PostgreSQL database and the committed media files.
// They run when TEST_DATABASE_URL is set (CI provides a Postgres service) and
// are skipped otherwise. Each run starts from a truncated, freshly seeded
// catalogue so the fixture counts are exact.

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/william22913/clover-be/internal/config"
	"github.com/william22913/clover-be/internal/db"
	"github.com/william22913/clover-be/internal/observability"
	"github.com/william22913/clover-be/internal/seed"
	"github.com/william22913/clover-be/internal/sessions"
	"github.com/william22913/clover-be/migrations"
)

var (
	poolOnce sync.Once
	sharedDB *pgxpool.Pool
	poolFail error
)

// database returns the migrated, seeded test database, or skips the test.
func database(t *testing.T) *pgxpool.Pool {
	t.Helper()
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL is not set: database-backed cases run in CI")
	}
	poolOnce.Do(func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()

		pool, err := db.Connect(ctx, url)
		if err != nil {
			poolFail = err
			return
		}
		if err := db.Migrate(ctx, pool, migrations.FS); err != nil {
			poolFail = err
			return
		}
		if _, err := pool.Exec(ctx, `TRUNCATE sessions, movies, accounts CASCADE`); err != nil {
			poolFail = err
			return
		}
		mediaDir, err := filepath.Abs(filepath.Join("..", "..", "media"))
		if err != nil {
			poolFail = err
			return
		}
		if err := seed.Run(ctx, pool, mediaDir); err != nil {
			poolFail = err
			return
		}
		sharedDB = pool
	})
	if poolFail != nil {
		t.Fatalf("test database: %v", poolFail)
	}
	return sharedDB
}

// apiClient is one browser: it keeps the session cookie like a real client and
// never sends an Origin header unless a test asks for one.
type apiClient struct {
	t      *testing.T
	base   string
	client *http.Client
}

// newAPI starts the API against the test database. relaxedRateLimitLift lets the
// bulk sign-in matrix run without tripping the per-IP bucket; the rate-limit
// test uses newAPIStrict so the real bucket is exercised.
func newAPI(t *testing.T) *apiClient {
	t.Helper()
	return startAPI(t, true)
}

func newAPIStrict(t *testing.T) *apiClient {
	t.Helper()
	return startAPI(t, false)
}

func startAPI(t *testing.T, relaxLimit bool) *apiClient {
	t.Helper()
	pool := database(t)
	mediaDir, err := filepath.Abs(filepath.Join("..", "..", "media"))
	if err != nil {
		t.Fatalf("media dir: %v", err)
	}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	server := New(Deps{
		Pool: pool,
		Config: config.Config{
			PublicOrigin: "http://localhost:3000",
			MediaDir:     mediaDir,
		},
		Metrics: observability.New(),
		Logger:  logger,
	})
	if relaxLimit {
		server.loginLimiter = newRateLimiter(100000, 100000)
	}
	ts := httptest.NewServer(server.Handler())
	t.Cleanup(ts.Close)

	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatalf("cookie jar: %v", err)
	}
	return &apiClient{t: t, base: ts.URL, client: &http.Client{Jar: jar, Timeout: 30 * time.Second}}
}

func (c *apiClient) do(method, path string, body io.Reader, headers map[string]string) *http.Response {
	c.t.Helper()
	req, err := http.NewRequest(method, c.base+path, body)
	if err != nil {
		c.t.Fatalf("request: %v", err)
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := c.client.Do(req)
	if err != nil {
		c.t.Fatalf("%s %s: %v", method, path, err)
	}
	return resp
}

func (c *apiClient) login(identifier, password string) *http.Response {
	c.t.Helper()
	payload, err := json.Marshal(map[string]string{"identifier": identifier, "password": password})
	if err != nil {
		c.t.Fatalf("marshal: %v", err)
	}
	return c.do(http.MethodPost, "/v1/auth/login", bytes.NewReader(payload), map[string]string{
		"Content-Type": "application/json",
	})
}

func (c *apiClient) mustLogin(identifier, password string) {
	c.t.Helper()
	resp := c.login(identifier, password)
	defer drain(resp)
	if resp.StatusCode != http.StatusOK {
		c.t.Fatalf("login %s: status %d", identifier, resp.StatusCode)
	}
}

func decode[T any](t *testing.T, resp *http.Response) T {
	t.Helper()
	defer drain(resp)
	var out T
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	return out
}

func decodeError(t *testing.T, resp *http.Response) errorEnvelope {
	t.Helper()
	return decode[errorEnvelope](t, resp)
}

func drain(resp *http.Response) {
	_, _ = io.Copy(io.Discard, resp.Body)
	_ = resp.Body.Close()
}

// TestCatalogue_TC003_AnonymousReadsTwentyFilms covers TC-003: the catalogue is
// public, complete and carries every field the page needs.
func TestCatalogue_TC003_AnonymousReadsTwentyFilms(t *testing.T) {
	api := newAPI(t)
	resp := api.do(http.MethodGet, "/v1/movies", nil, nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d", resp.StatusCode)
	}
	if got := resp.Header.Get("Cache-Control"); got != "public, max-age=60" {
		t.Fatalf("Cache-Control = %q", got)
	}
	body := decode[listMoviesResponse](t, resp)
	if len(body.Movies) != 20 {
		t.Fatalf("films = %d, want 20", len(body.Movies))
	}
	for i, m := range body.Movies {
		if m.ID == "" || m.Title == "" || m.Description == "" ||
			m.PosterURL == "" || m.VideoURL == "" {
			t.Fatalf("film %d has a missing field: %+v", i, m)
		}
		if m.Position != i+1 {
			t.Fatalf("%s position = %d, want %d", m.ID, m.Position, i+1)
		}
		if m.PosterURL != "/v1/media/movies/"+m.ID+"/poster" ||
			m.VideoURL != "/v1/media/movies/"+m.ID+"/video" {
			t.Fatalf("%s media urls = %q %q", m.ID, m.PosterURL, m.VideoURL)
		}
	}
	if body.Movies[0].ID != "amber-drift" || body.Movies[19].ID != "descent" {
		t.Fatalf("unexpected editorial order: %s ... %s", body.Movies[0].ID, body.Movies[19].ID)
	}
}

// TestCatalogue_TC007_AcceptLanguageIsIgnored covers TC-007: no request header
// can change the language of the response.
func TestCatalogue_TC007_AcceptLanguageIsIgnored(t *testing.T) {
	api := newAPI(t)
	var first string
	for i, lang := range []string{"fr-FR", "", "xx-XX"} {
		headers := map[string]string{}
		if lang != "" {
			headers["Accept-Language"] = lang
		}
		resp := api.do(http.MethodGet, "/v1/movies", nil, headers)
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Accept-Language %q: status %d", lang, resp.StatusCode)
		}
		raw, err := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if err != nil {
			t.Fatalf("read: %v", err)
		}
		body := string(raw)
		if !strings.Contains(body, "Amber Drift") || strings.Contains(body, "\\u") {
			t.Fatalf("Accept-Language %q: catalogue is not plain English: %s", lang, body)
		}
		if i == 0 {
			first = body
		} else if body != first {
			t.Fatalf("Accept-Language %q changed the response body", lang)
		}
	}

	// The sign-in error copy is English too, whatever the header says.
	resp := api.do(http.MethodPost, "/v1/auth/login",
		strings.NewReader(`{"identifier":"alex","password":"Wrong1234"}`),
		map[string]string{"Content-Type": "application/json", "Accept-Language": "fr-FR"})
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("login status = %d", resp.StatusCode)
	}
	body := decodeError(t, resp)
	if body.Error.Message != "Incorrect username/email or password." {
		t.Fatalf("message = %q", body.Error.Message)
	}
}

// TestCatalogue_TC062_TC063_SeedData covers TC-062 and TC-063: every film has a
// working poster and video, durations never exceed 30:00, and the database
// refuses a 30:01 row.
func TestCatalogue_TC062_TC063_SeedData(t *testing.T) {
	api := newAPI(t)
	pool := database(t)

	body := decode[listMoviesResponse](t, api.do(http.MethodGet, "/v1/movies", nil, nil))
	var longest int
	for _, m := range body.Movies {
		if m.DurationSeconds < 1 || m.DurationSeconds > 1800 {
			t.Fatalf("%s duration = %ds, outside 1..1800", m.ID, m.DurationSeconds)
		}
		if m.DurationSeconds > longest {
			longest = m.DurationSeconds
		}
		poster := api.do(http.MethodGet, m.PosterURL, nil, nil)
		if poster.StatusCode != http.StatusOK {
			t.Fatalf("%s poster: status %d", m.ID, poster.StatusCode)
		}
		if ct := poster.Header.Get("Content-Type"); ct != "image/webp" {
			t.Fatalf("%s poster Content-Type = %q", m.ID, ct)
		}
		drain(poster)
	}
	// 960 s is the seeded maximum; TC-052 needs a film longer than 15 minutes.
	if longest != 960 {
		t.Fatalf("longest film = %ds, want 960", longest)
	}

	ctx := context.Background()
	if _, err := pool.Exec(ctx,
		`INSERT INTO movies (id, slug, title, description, poster_path, video_path, duration_seconds, catalogue_order)
		 VALUES (gen_random_uuid(), 'too-long', 'Too Long', 'x', 'posters/too-long.webp', 'videos/too-long.mp4', 1800, 900)`); err != nil {
		t.Fatalf("a 30:00 film must be valid: %v", err)
	}
	defer func() {
		if _, err := pool.Exec(ctx, `DELETE FROM movies WHERE slug = 'too-long'`); err != nil {
			t.Fatalf("cleanup: %v", err)
		}
	}()
	if _, err := pool.Exec(ctx,
		`INSERT INTO movies (id, slug, title, description, poster_path, video_path, duration_seconds, catalogue_order)
		 VALUES (gen_random_uuid(), 'over-long', 'Over Long', 'x', 'posters/over-long.webp', 'videos/over-long.mp4', 1801, 901)`); err == nil {
		t.Fatal("the database accepted a 30:01 film")
	}
}

// TestMovie_TC058_UnknownFilmIsCleanNotFound covers TC-058.
func TestMovie_TC058_UnknownFilmIsCleanNotFound(t *testing.T) {
	api := newAPI(t)
	for _, path := range []string{
		"/v1/movies/does-not-exist",
		"/v1/movies/Amber-Drift",
	} {
		resp := api.do(http.MethodGet, path, nil, nil)
		if resp.StatusCode != http.StatusNotFound {
			t.Fatalf("%s: status = %d, want 404", path, resp.StatusCode)
		}
		body := decodeError(t, resp)
		if body.Error.Code != CodeMovieNotFound {
			t.Fatalf("%s: code = %q", path, body.Error.Code)
		}
		if body.Error.RequestID == "" {
			t.Fatalf("%s: missing requestId", path)
		}
		if strings.Contains(strings.ToLower(body.Error.Message), "sql") ||
			strings.Contains(body.Error.Message, "pgx") {
			t.Fatalf("%s: message leaks internal detail: %q", path, body.Error.Message)
		}
	}
}

// TestVideo_TC061_NoSessionIsRefused covers TC-061 for a normal and a ranged
// request, and proves the refusal body is JSON, not video bytes.
func TestVideo_TC061_NoSessionIsRefused(t *testing.T) {
	api := newAPI(t)
	for _, headers := range []map[string]string{nil, {"Range": "bytes=0-99"}} {
		resp := api.do(http.MethodGet, "/v1/media/movies/amber-drift/video", nil, headers)
		if resp.StatusCode != http.StatusUnauthorized {
			t.Fatalf("status = %d, want 401", resp.StatusCode)
		}
		if ct := resp.Header.Get("Content-Type"); !strings.HasPrefix(ct, "application/json") {
			t.Fatalf("Content-Type = %q; a refusal must not be media", ct)
		}
		body := decodeError(t, resp)
		if body.Error.Code != CodeUnauthenticated {
			t.Fatalf("code = %q", body.Error.Code)
		}
	}
}

// TestAuth_TC023_TC024_OneGenericRefusal covers TC-023 and TC-024 at the API:
// a wrong password and an unknown account are indistinguishable.
func TestAuth_TC023_TC024_OneGenericRefusal(t *testing.T) {
	api := newAPI(t)
	cases := [][2]string{
		{"alex", "Wrong1234"},
		{"nobody_here", "Watch1234"},
		{"nobody_here@example.com", "Watch1234"},
	}
	var first errorEnvelope
	for i, creds := range cases {
		resp := api.login(creds[0], creds[1])
		if resp.StatusCode != http.StatusUnauthorized {
			t.Fatalf("%s: status = %d", creds[0], resp.StatusCode)
		}
		body := decodeError(t, resp)
		if body.Error.Code != CodeAuthInvalidCredentials || body.Error.Message != "Incorrect username/email or password." {
			t.Fatalf("%s: %+v", creds[0], body.Error)
		}
		if len(body.Error.Fields) != 0 {
			t.Fatalf("%s: refusal names a field: %+v", creds[0], body.Error.Fields)
		}
		if i == 0 {
			first = body
		} else if body.Error.Code != first.Error.Code || body.Error.Message != first.Error.Message {
			t.Fatalf("refusals differ: %+v vs %+v", body.Error, first.Error)
		}
	}

	// TC-023, second half: the form is still usable and the correct password
	// signs in immediately.
	api.mustLogin("alex", "Watch1234")
}

// TestAuth_IdentifierMatrix covers the identifier boundary cases TC-028 to
// TC-038 through the real endpoint.
func TestAuth_IdentifierMatrix(t *testing.T) {
	api := newAPI(t)
	cases := []struct {
		tc         string
		identifier string
		password   string
		want       int
	}{
		{"TC-028", "  alex  ", "Watch1234", http.StatusOK},
		{"TC-029", "abc", "Abcdefg1", http.StatusOK},
		{"TC-030", "ab", "Abcdefg1", http.StatusUnauthorized},
		{"TC-031", strings.Repeat("a", 50), "Abcdefg1", http.StatusOK},
		{"TC-032", strings.Repeat("a", 51), "Abcdefg1", http.StatusUnauthorized},
		{"TC-033", "mira_k9", "Play2025x", http.StatusOK},
		{"TC-034", "mira-k", "Play2025x", http.StatusUnauthorized},
		{"TC-034", "mira k", "Play2025x", http.StatusUnauthorized},
		{"TC-034", "mira.k", "Play2025x", http.StatusUnauthorized},
		{"TC-034", "mira@k", "Play2025x", http.StatusUnauthorized},
		{"TC-035", "mira.k@example.com", "Play2025x", http.StatusOK},
		{"TC-036", "alex@", "Watch1234", http.StatusUnauthorized},
		{"TC-036", "@example.com", "Watch1234", http.StatusUnauthorized},
		{"TC-036", "alex@@example.com", "Watch1234", http.StatusUnauthorized},
		{"TC-036", "alex example.com", "Watch1234", http.StatusUnauthorized},
		{"TC-037", "ALEX", "Watch1234", http.StatusOK},
		{"TC-038", "ALEX.RIVERA@EXAMPLE.COM", "Watch1234", http.StatusOK},
	}
	for _, tc := range cases {
		resp := api.login(tc.identifier, tc.password)
		if resp.StatusCode != tc.want {
			drain(resp)
			t.Fatalf("%s: login(%q) = %d, want %d", tc.tc, tc.identifier, resp.StatusCode, tc.want)
		}
		if tc.want == http.StatusUnauthorized {
			body := decodeError(t, resp)
			if body.Error.Code != CodeAuthInvalidCredentials {
				t.Fatalf("%s: code = %q", tc.tc, body.Error.Code)
			}
		} else {
			drain(resp)
		}
	}
}

// TestAuth_PasswordMatrix covers the password boundary cases TC-039 to TC-046.
func TestAuth_PasswordMatrix(t *testing.T) {
	api := newAPI(t)
	cases := []struct {
		tc       string
		username string
		password string
		want     int
	}{
		{"TC-039", "minpw", "Abcdefg1", http.StatusOK},
		{"TC-040", "alex", "Abcdef1", http.StatusUnauthorized},
		{"TC-041", "maxpw", "A1" + strings.Repeat("x", 126), http.StatusOK},
		{"TC-042", "maxpw", "A1" + strings.Repeat("x", 127), http.StatusUnauthorized},
		{"TC-043", "alex", "Abcdefgh", http.StatusUnauthorized},
		{"TC-044", "alex", "12345678", http.StatusUnauthorized},
		{"TC-045", "alex", "watch1234", http.StatusUnauthorized},
		{"TC-046", "spacey", "Watch1234 ", http.StatusOK},
		{"TC-046", "spacey", "Watch1234", http.StatusUnauthorized},
	}
	for _, tc := range cases {
		resp := api.login(tc.username, tc.password)
		if resp.StatusCode != tc.want {
			drain(resp)
			t.Fatalf("%s: login(%q) = %d, want %d", tc.tc, tc.password, resp.StatusCode, tc.want)
		}
		if tc.want == http.StatusUnauthorized {
			body := decodeError(t, resp)
			if body.Error.Code != CodeAuthInvalidCredentials || body.Error.Message != "Incorrect username/email or password." {
				t.Fatalf("%s: %+v", tc.tc, body.Error)
			}
		} else {
			drain(resp)
		}
	}
}

// TestAuth_TC047_NoAccountLockout covers REQ-028: ten failures then the correct
// password, with no delay and no lock.
func TestAuth_TC047_NoAccountLockout(t *testing.T) {
	api := newAPI(t)
	for i := 0; i < 10; i++ {
		resp := api.login("alex", fmt.Sprintf("Wrong%04d", i))
		if resp.StatusCode != http.StatusUnauthorized {
			drain(resp)
			t.Fatalf("failure %d: status %d", i+1, resp.StatusCode)
		}
		body := decodeError(t, resp)
		if body.Error.Code != CodeAuthInvalidCredentials {
			t.Fatalf("failure %d: code = %q", i+1, body.Error.Code)
		}
	}
	start := time.Now()
	resp := api.login("alex", "Watch1234")
	if resp.StatusCode != http.StatusOK {
		drain(resp)
		t.Fatalf("correct password after failures: status %d", resp.StatusCode)
	}
	drain(resp)
	// A lockout would take at least a second; a normal argon2 verify does not.
	// (Leaving a generous ceiling keeps the test about behaviour, not speed.)
	if elapsed := time.Since(start); elapsed > 5*time.Second {
		t.Fatalf("correct sign-in took %v after failures", elapsed)
	}
}

// TestAuth_RateLimit covers S9: the per-IP bucket answers 429 with Retry-After
// once exhausted, and no sign-in attempt is processed.
func TestAuth_RateLimit(t *testing.T) {
	api := newAPIStrict(t)
	var last *http.Response
	for i := 0; i < 31; i++ {
		resp := api.login("nobody_here", "Watch1234")
		if i < 30 {
			if resp.StatusCode != http.StatusUnauthorized {
				drain(resp)
				t.Fatalf("attempt %d: status %d", i+1, resp.StatusCode)
			}
			drain(resp)
			continue
		}
		last = resp
	}
	if last.StatusCode != http.StatusTooManyRequests {
		t.Fatalf("attempt 31: status = %d, want 429", last.StatusCode)
	}
	if last.Header.Get("Retry-After") == "" {
		t.Fatal("429 without Retry-After")
	}
	body := decodeError(t, last)
	if body.Error.Code != CodeRateLimited {
		t.Fatalf("code = %q", body.Error.Code)
	}
}

// TestAuth_TC025_TC026_FieldMessages covers the server half of TC-025 and
// TC-026: an empty field is a 422 naming exactly that field, in English.
func TestAuth_TC025_TC026_FieldMessages(t *testing.T) {
	api := newAPI(t)
	cases := []struct {
		name   string
		body   string
		field  string
		status int
	}{
		{"identifier empty", `{"identifier":"","password":"Watch1234"}`, "identifier", http.StatusUnprocessableEntity},
		{"identifier spaces", `{"identifier":"    ","password":"Watch1234"}`, "identifier", http.StatusUnprocessableEntity},
		{"password empty", `{"identifier":"alex","password":""}`, "password", http.StatusUnprocessableEntity},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			resp := api.do(http.MethodPost, "/v1/auth/login", strings.NewReader(tc.body),
				map[string]string{"Content-Type": "application/json"})
			if resp.StatusCode != tc.status {
				drain(resp)
				t.Fatalf("status = %d, want %d", resp.StatusCode, tc.status)
			}
			body := decodeError(t, resp)
			if body.Error.Code != CodeValidationError {
				t.Fatalf("code = %q", body.Error.Code)
			}
			if len(body.Error.Fields) != 1 || body.Error.Fields[tc.field] == "" {
				t.Fatalf("fields = %+v, want only %q", body.Error.Fields, tc.field)
			}
			if body.Error.Fields[tc.field] != map[string]string{
				"identifier": "Enter your username or email.",
				"password":   "Enter your password.",
			}[tc.field] {
				t.Fatalf("field message = %q", body.Error.Fields[tc.field])
			}
		})
	}
}

// TestAuth_MalformedRequests covers S10: a non-JSON body, an unknown field and
// an oversized body are refused with the VALIDATION_ERROR code and no sign-in
// attempt.
func TestAuth_MalformedRequests(t *testing.T) {
	api := newAPI(t)
	cases := []struct {
		name        string
		contentType string
		body        string
		status      int
	}{
		{"not json", "application/json", `identifier=alex&password=Watch1234`, http.StatusBadRequest},
		{"wrong content type", "text/plain", `{"identifier":"alex","password":"Watch1234"}`, http.StatusBadRequest},
		{"unknown field", "application/json", `{"identifier":"alex","password":"Watch1234","admin":true}`, http.StatusBadRequest},
		{"two objects", "application/json", `{"identifier":"alex","password":"Watch1234"}{"identifier":"alex"}`, http.StatusBadRequest},
		{"oversized", "application/json", `{"identifier":"` + strings.Repeat("a", 8<<10) + `","password":"Watch1234"}`, http.StatusRequestEntityTooLarge},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			resp := api.do(http.MethodPost, "/v1/auth/login", strings.NewReader(tc.body),
				map[string]string{"Content-Type": tc.contentType})
			if resp.StatusCode != tc.status {
				drain(resp)
				t.Fatalf("status = %d, want %d", resp.StatusCode, tc.status)
			}
			body := decodeError(t, resp)
			if body.Error.Code != CodeValidationError {
				t.Fatalf("code = %q", body.Error.Code)
			}
			if body.Error.RequestID == "" {
				t.Fatal("missing requestId")
			}
		})
	}
}

// TestAuth_CSRF_CrossOriginIsRefused covers S8 through the endpoint.
func TestAuth_CSRF_CrossOriginIsRefused(t *testing.T) {
	api := newAPI(t)
	resp := api.do(http.MethodPost, "/v1/auth/login",
		strings.NewReader(`{"identifier":"alex","password":"Watch1234"}`),
		map[string]string{"Content-Type": "application/json", "Origin": "https://evil.example"})
	if resp.StatusCode != http.StatusForbidden {
		drain(resp)
		t.Fatalf("status = %d, want 403", resp.StatusCode)
	}
	body := decodeError(t, resp)
	if body.Error.Code != CodeCSRFRejected {
		t.Fatalf("code = %q", body.Error.Code)
	}

	// The same-origin browser POST is accepted.
	resp = api.do(http.MethodPost, "/v1/auth/login",
		strings.NewReader(`{"identifier":"alex","password":"Watch1234"}`),
		map[string]string{"Content-Type": "application/json", "Origin": "http://localhost:3000"})
	if resp.StatusCode != http.StatusOK {
		drain(resp)
		t.Fatalf("same-origin login status = %d", resp.StatusCode)
	}
	drain(resp)
}

// TestMe_TC017_TC021_SessionLifecycle covers TC-017 (the cookie alone keeps the
// viewer signed in across requests) and TC-021 (a revoked session cannot read
// account data while a fresh one can).
func TestMe_TC017_TC021_SessionLifecycle(t *testing.T) {
	api := newAPI(t)
	api.mustLogin("alex", "Watch1234")

	for i := 0; i < 2; i++ {
		resp := api.do(http.MethodGet, "/v1/auth/me", nil, nil)
		if resp.StatusCode != http.StatusOK {
			drain(resp)
			t.Fatalf("me %d: status %d", i+1, resp.StatusCode)
		}
		body := decode[userResponse](t, resp)
		if body.User.Username != "alex" || body.User.ID == "" {
			t.Fatalf("me = %+v", body.User)
		}
	}

	resp := api.do(http.MethodPost, "/v1/auth/logout", nil, nil)
	if resp.StatusCode != http.StatusNoContent {
		drain(resp)
		t.Fatalf("logout status = %d", resp.StatusCode)
	}
	drain(resp)

	resp = api.do(http.MethodGet, "/v1/auth/me", nil, nil)
	if resp.StatusCode != http.StatusUnauthorized {
		drain(resp)
		t.Fatalf("me after logout: status %d", resp.StatusCode)
	}
	body := decodeError(t, resp)
	if body.Error.Code != CodeUnauthenticated {
		t.Fatalf("code = %q", body.Error.Code)
	}

	// A fresh session still works, proving only the old one was revoked.
	api.mustLogin("alex", "Watch1234")
	if resp := api.do(http.MethodGet, "/v1/auth/me", nil, nil); resp.StatusCode != http.StatusOK {
		drain(resp)
		t.Fatalf("fresh session: status %d", resp.StatusCode)
	} else {
		drain(resp)
	}
}

// TestVideo_TC020_ReplayedCookieAfterLogout covers TC-020: the captured cookie
// cannot fetch video, plain or ranged, after sign-out.
func TestVideo_TC020_ReplayedCookieAfterLogout(t *testing.T) {
	api := newAPI(t)
	api.mustLogin("alex", "Watch1234")

	// Capture the cookie the browser holds.
	var captured string
	for _, cookie := range api.client.Jar.Cookies(mustURL(t, api.base)) {
		if cookie.Name == sessionCookieName {
			captured = cookie.Value
		}
	}
	if captured == "" {
		t.Fatal("login set no session cookie")
	}

	resp := api.do(http.MethodPost, "/v1/auth/logout", nil, nil)
	drain(resp)

	for _, headers := range []map[string]string{
		{"Cookie": sessionCookieName + "=" + captured},
		{"Cookie": sessionCookieName + "=" + captured, "Range": "bytes=0-99"},
	} {
		resp := api.do(http.MethodGet, "/v1/media/movies/amber-drift/video", nil, headers)
		if resp.StatusCode != http.StatusUnauthorized {
			drain(resp)
			t.Fatalf("replayed cookie: status = %d, want 401", resp.StatusCode)
		}
		if ct := resp.Header.Get("Content-Type"); !strings.HasPrefix(ct, "application/json") {
			t.Fatalf("replayed cookie returned %q", ct)
		}
		drain(resp)
	}
}

// TestAuth_TC016_SessionExpiryRevokesAccess covers TC-016's server half: a
// session expired server-side cannot play video or read the account.
func TestAuth_TC016_SessionExpiryRevokesAccess(t *testing.T) {
	api := newAPI(t)
	api.mustLogin("alex", "Watch1234")

	me := decode[userResponse](t, api.do(http.MethodGet, "/v1/auth/me", nil, nil))
	// Move the whole session into the past: the row's CHECK keeps
	// expires_at after created_at, so both are shifted.
	if _, err := database(t).Exec(context.Background(),
		`UPDATE sessions
		    SET created_at = now() - interval '2 minutes',
		        updated_at = now() - interval '2 minutes',
		        last_seen_at = now() - interval '2 minutes',
		        expires_at = now() - interval '1 minute'
		  WHERE account_id = $1`, me.User.ID); err != nil {
		t.Fatalf("expire session: %v", err)
	}

	for _, path := range []string{"/v1/auth/me", "/v1/media/movies/amber-drift/video"} {
		resp := api.do(http.MethodGet, path, nil, nil)
		if resp.StatusCode != http.StatusUnauthorized {
			drain(resp)
			t.Fatalf("%s with an expired session: status %d", path, resp.StatusCode)
		}
		drain(resp)
	}
}

// TestSeed_Reorder_ReappliesCleanly covers a re-run after the catalogue order
// changed: UNIQUE(catalogue_order) must not block the swap halfway through.
func TestSeed_Reorder_ReappliesCleanly(t *testing.T) {
	pool := database(t)
	ctx := context.Background()

	if _, err := pool.Exec(ctx, `UPDATE movies SET catalogue_order = 21 - catalogue_order`); err != nil {
		t.Fatalf("reverse catalogue order: %v", err)
	}
	mediaDir, err := filepath.Abs(filepath.Join("..", "..", "media"))
	if err != nil {
		t.Fatalf("media dir: %v", err)
	}
	if err := seed.Run(ctx, pool, mediaDir); err != nil {
		t.Fatalf("re-seed after a reorder: %v", err)
	}

	rows, err := pool.Query(ctx, `SELECT catalogue_order, title FROM movies ORDER BY catalogue_order`)
	if err != nil {
		t.Fatalf("read catalogue: %v", err)
	}
	defer rows.Close()
	position := 0
	for rows.Next() {
		var order int
		var title string
		if err := rows.Scan(&order, &title); err != nil {
			t.Fatalf("scan: %v", err)
		}
		position++
		if order != position {
			t.Fatalf("position %d holds order %d", position, order)
		}
		if title != seed.Films[position-1].Title {
			t.Fatalf("position %d holds %q, want %q", position, title, seed.Films[position-1].Title)
		}
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("rows: %v", err)
	}
	if position != len(seed.Films) {
		t.Fatalf("re-seed left %d films, want %d", position, len(seed.Films))
	}
}

// TestAuth_TC017_SlidingSessionReissuesTheCookie covers the browser half of the
// sliding lifetime: when the server extends a day-old session, the response
// carries a fresh cookie instead of letting the browser drop it at day 30.
func TestAuth_TC017_SlidingSessionReissuesTheCookie(t *testing.T) {
	api := newAPI(t)
	api.mustLogin("alex", "Watch1234")

	me := decode[userResponse](t, api.do(http.MethodGet, "/v1/auth/me", nil, nil))
	// Idle for two days: past the 24h slide threshold, well inside the TTL.
	if _, err := database(t).Exec(context.Background(),
		`UPDATE sessions SET last_seen_at = now() - interval '2 days' WHERE account_id = $1`,
		me.User.ID); err != nil {
		t.Fatalf("age session: %v", err)
	}

	resp := api.do(http.MethodGet, "/v1/auth/me", nil, nil)
	if resp.StatusCode != http.StatusOK {
		drain(resp)
		t.Fatalf("status %d", resp.StatusCode)
	}
	var refreshed *http.Cookie
	for _, c := range resp.Cookies() {
		if c.Name == "clover_session" {
			refreshed = c
		}
	}
	drain(resp)
	if refreshed == nil {
		t.Fatal("an extended session did not re-issue the cookie")
	}
	if refreshed.MaxAge != int(sessions.TTL/time.Second) {
		t.Fatalf("refreshed Max-Age = %d", refreshed.MaxAge)
	}
	if !refreshed.HttpOnly || refreshed.SameSite != http.SameSiteLaxMode {
		t.Fatalf("refreshed cookie lost its attributes: %+v", refreshed)
	}
}

// TestAuth_TC022_TwoDevices covers TC-022: two independent sign-ins for the
// same account are both live, and signing out of one leaves the other working.
func TestAuth_TC022_TwoDevices(t *testing.T) {
	deviceOne := newAPI(t)
	deviceTwo := newAPI(t)

	deviceOne.mustLogin("mira_k9", "Play2025x")
	deviceTwo.mustLogin("mira_k9", "Play2025x")

	for name, device := range map[string]*apiClient{"one": deviceOne, "two": deviceTwo} {
		resp := device.do(http.MethodGet, "/v1/auth/me", nil, nil)
		if resp.StatusCode != http.StatusOK {
			drain(resp)
			t.Fatalf("device %s: status %d", name, resp.StatusCode)
		}
		if body := decode[userResponse](t, resp); body.User.Username != "mira_k9" {
			t.Fatalf("device %s: user = %+v", name, body.User)
		}
	}

	drain(deviceOne.do(http.MethodPost, "/v1/auth/logout", nil, nil))

	resp := deviceTwo.do(http.MethodGet, "/v1/auth/me", nil, nil)
	if resp.StatusCode != http.StatusOK {
		drain(resp)
		t.Fatalf("device two after device one signed out: status %d", resp.StatusCode)
	}
	drain(resp)
}

// TestVideo_TC052_TC055_TC056_RangeStreaming covers the server half of TC-052,
// TC-055 and TC-056: a signed-in viewer can Range-seek every film (206 with the
// requested slice), and the transport supports progressive playback.
func TestVideo_TC052_TC055_TC056_RangeStreaming(t *testing.T) {
	api := newAPI(t)
	api.mustLogin("alex", "Watch1234")

	// TC-052/TC-055: a seek into the middle of the 16-minute film.
	resp := api.do(http.MethodGet, "/v1/media/movies/long-evening/video", nil,
		map[string]string{"Range": "bytes=100000-"})
	if resp.StatusCode != http.StatusPartialContent {
		drain(resp)
		t.Fatalf("long-evening seek: status = %d, want 206", resp.StatusCode)
	}
	if got := resp.Header.Get("Content-Range"); !strings.HasPrefix(got, "bytes 100000-") {
		t.Fatalf("Content-Range = %q", got)
	}
	if resp.Header.Get("Accept-Ranges") != "bytes" {
		t.Fatalf("Accept-Ranges = %q", resp.Header.Get("Accept-Ranges"))
	}
	if got := resp.Header.Get("Cache-Control"); got != "private, no-store" {
		t.Fatalf("Cache-Control = %q", got)
	}
	drain(resp)

	// TC-056: every seeded film streams its own bytes.
	films := decode[listMoviesResponse](t, api.do(http.MethodGet, "/v1/movies", nil, nil))
	for _, m := range films.Movies {
		resp := api.do(http.MethodGet, m.VideoURL, nil, map[string]string{"Range": "bytes=0-3"})
		if resp.StatusCode != http.StatusPartialContent && resp.StatusCode != http.StatusOK {
			drain(resp)
			t.Fatalf("%s video: status %d", m.ID, resp.StatusCode)
		}
		if ct := resp.Header.Get("Content-Type"); ct != "video/mp4" {
			drain(resp)
			t.Fatalf("%s video Content-Type = %q", m.ID, ct)
		}
		raw, err := io.ReadAll(io.LimitReader(resp.Body, 4))
		_ = resp.Body.Close()
		if err != nil || len(raw) == 0 {
			t.Fatalf("%s video returned no bytes", m.ID)
		}
	}
}

// TestSeed_TC048_IdentifiersAreUnique covers TC-048 against the real rows.
func TestSeed_TC048_IdentifiersAreUnique(t *testing.T) {
	pool := database(t)
	queries := []struct{ name, sql string }{
		{"username", `SELECT count(*), count(DISTINCT lower(username)) FROM accounts`},
		{"email", `SELECT count(*), count(DISTINCT lower(email)) FROM accounts`},
	}
	for _, q := range queries {
		var total, distinct int
		if err := pool.QueryRow(context.Background(), q.sql).Scan(&total, &distinct); err != nil {
			t.Fatalf("count %s: %v", q.name, err)
		}
		if total == 0 || total != distinct {
			t.Fatalf("%s: %d rows but %d distinct values", q.name, total, distinct)
		}
	}
}

// TestSeed_TC014_NoRegistrationRoute covers the API side of TC-014: there is no
// account-creation endpoint, only the unknown-route answer.
func TestSeed_TC014_NoRegistrationRoute(t *testing.T) {
	api := newAPI(t)
	for _, path := range []string{"/v1/auth/register", "/v1/auth/sign-up", "/v1/accounts"} {
		resp := api.do(http.MethodPost, path, strings.NewReader(`{"username":"new"}`),
			map[string]string{"Content-Type": "application/json"})
		if resp.StatusCode != http.StatusNotFound && resp.StatusCode != http.StatusMethodNotAllowed {
			drain(resp)
			t.Fatalf("%s: status = %d, want 404 or 405", path, resp.StatusCode)
		}
		drain(resp)
	}
}

func mustURL(t *testing.T, raw string) *url.URL {
	t.Helper()
	parsed, err := url.Parse(raw)
	if err != nil {
		t.Fatalf("parse %q: %v", raw, err)
	}
	return parsed
}
