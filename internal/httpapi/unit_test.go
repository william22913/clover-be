package httpapi

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/william22913/clover-be/internal/config"
	"github.com/william22913/clover-be/internal/i18n"
	"github.com/william22913/clover-be/internal/observability"
)

// TestPermissions_Matrix covers the architecture's "<path>:<action>" rule with
// dot-prefix coverage and the "all" action.
func TestPermissions_Matrix(t *testing.T) {
	if !RoleMember.Allows(PermPlaybackWatch) {
		t.Fatal("member cannot watch playback")
	}
	if !RoleMember.Allows(PermAccountView) || !RoleMember.Allows(PermSessionDelete) {
		t.Fatal("member lost an account/session grant")
	}
	if !RoleVisitor.Allows(PermCatalogView) || !RoleVisitor.Allows(PermSessionCreate) {
		t.Fatal("visitor lost a public grant")
	}
	if RoleVisitor.Allows(PermPlaybackWatch) || RoleVisitor.Allows(PermAccountView) {
		t.Fatal("visitor holds a member-only grant")
	}
	if !AllowsGrants([]string{"catalog:all"}, "catalog:view") {
		t.Fatal("action 'all' did not cover a specific action")
	}
	if !AllowsGrants([]string{"catalog:all"}, "catalog.view:read") {
		t.Fatal("a covering path and the 'all' action did not grant a child path")
	}
	if AllowsGrants([]string{"catalog:view"}, "catalogx:view") {
		t.Fatal("a bare prefix covered an unrelated path")
	}
	if AllowsGrants([]string{"catalog:view"}, "catalog:delete") {
		t.Fatal("a grant covered a different action")
	}
	if AllowsGrants([]string{"malformed", ":", "a:"}, "catalog:view") {
		t.Fatal("a malformed grant matched")
	}
}

// TestRateLimiter_TC047_NoAccountLockout covers REQ-028: repeated failures are
// throttled per IP over time, never by disabling the account, and the bucket
// refills so a later correct attempt is served.
func TestRateLimiter_TC047_NoAccountLockout(t *testing.T) {
	now := time.Date(2026, 11, 14, 12, 0, 0, 0, time.UTC)
	limiter := newRateLimiter(30, 30)
	limiter.now = func() time.Time { return now }

	for i := 0; i < 10; i++ {
		if ok, _ := limiter.allow("203.0.113.7"); !ok {
			t.Fatalf("attempt %d was throttled before the bucket ran out", i+1)
		}
	}
	for i := 0; i < 20; i++ {
		if ok, _ := limiter.allow("203.0.113.7"); !ok {
			t.Fatalf("attempt %d was throttled inside the documented burst", i+11)
		}
	}
	ok, wait := limiter.allow("203.0.113.7")
	if ok || wait <= 0 {
		t.Fatalf("31st attempt in a minute: ok=%v wait=%v", ok, wait)
	}
	if ok, _ := limiter.allow("198.51.100.9"); !ok {
		t.Fatal("one exhausted IP throttled a different IP")
	}
	now = now.Add(2 * time.Minute)
	if ok, _ := limiter.allow("203.0.113.7"); !ok {
		t.Fatal("the bucket did not refill after waiting")
	}
}

// TestSameOrigin_S8 covers the CSRF check: a cross-origin or cross-site POST is
// refused, a same-origin POST and a request with neither header are allowed,
// and a GET is never affected.
func TestSameOrigin_S8(t *testing.T) {
	s := newTestServerObject(t, "https://clover.example")

	cases := []struct {
		name     string
		method   string
		header   map[string]string
		want     int
		wantCSRF bool
	}{
		{name: "same origin", method: http.MethodPost, header: map[string]string{"Origin": "https://clover.example"}, want: http.StatusOK},
		{name: "cross origin", method: http.MethodPost, header: map[string]string{"Origin": "https://evil.example"}, want: http.StatusForbidden, wantCSRF: true},
		{name: "cross site fetch", method: http.MethodPost, header: map[string]string{"Sec-Fetch-Site": "cross-site"}, want: http.StatusForbidden, wantCSRF: true},
		{name: "same origin fetch", method: http.MethodPost, header: map[string]string{"Sec-Fetch-Site": "same-origin"}, want: http.StatusOK},
		{name: "no headers", method: http.MethodPost, want: http.StatusOK},
		{name: "get unaffected", method: http.MethodGet, header: map[string]string{"Origin": "https://evil.example"}, want: http.StatusOK},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, "/v1/auth/login", nil)
			for k, v := range tc.header {
				req.Header.Set(k, v)
			}
			rec := httptest.NewRecorder()
			s.sameOrigin(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
			})).ServeHTTP(rec, req)

			if rec.Code != tc.want {
				t.Fatalf("status = %d, want %d", rec.Code, tc.want)
			}
			if tc.wantCSRF {
				var body errorEnvelope
				if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
					t.Fatalf("body is not the error envelope: %v", err)
				}
				if body.Error.Code != CodeCSRFRejected {
					t.Fatalf("code = %q, want %q", body.Error.Code, CodeCSRFRejected)
				}
			}
		})
	}
}

// TestRequestID_IsGeneratedServerSide covers S16: an incoming X-Request-Id is
// not trusted, and the generated id appears in the header, the error body and
// the context used by the log line.
func TestRequestID_IsGeneratedServerSide(t *testing.T) {
	var seen string
	handler := withRequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen = requestID(r)
		w.WriteHeader(http.StatusNoContent)
	}))
	req := httptest.NewRequest(http.MethodGet, "/v1/movies", nil)
	req.Header.Set("X-Request-Id", "attacker-chosen")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	header := rec.Header().Get("X-Request-Id")
	if header == "" || header == "attacker-chosen" {
		t.Fatalf("X-Request-Id = %q", header)
	}
	if len(header) != 16 {
		t.Fatalf("request id %q is not 8 random bytes in hex", header)
	}
	if seen != header {
		t.Fatalf("context id %q does not match the header %q", seen, header)
	}
}

// TestI18nCatalogueCoversEveryErrorCode is the test internal/i18n refers to: no
// code the API can return is missing English copy.
func TestI18nCatalogueCoversEveryErrorCode(t *testing.T) {
	for _, code := range []string{
		CodeAuthInvalidCredentials, CodeValidationError, CodeUnauthenticated,
		CodeCSRFRejected, CodeMovieNotFound, CodeNotFound,
		CodeMethodNotAllowed, CodeRateLimited, CodeInternalError,
		"FIELD_IDENTIFIER_REQUIRED", "FIELD_PASSWORD_REQUIRED",
	} {
		if msg := i18n.Message(code); msg == "" || msg == code {
			t.Fatalf("%s has no English message", code)
		}
	}
}

// TestErrorEnvelope_Shape checks the one error body every failure uses.
func TestErrorEnvelope_Shape(t *testing.T) {
	s := newTestServerObject(t, "https://clover.example")
	handler := withRequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.writeError(w, r, http.StatusUnprocessableEntity, CodeValidationError, map[string]string{
			"identifier": i18n.Message("FIELD_IDENTIFIER_REQUIRED"),
		})
	}))

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/v1/auth/login", nil))

	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d", rec.Code)
	}
	if got := rec.Header().Get("Content-Type"); !strings.HasPrefix(got, "application/json") {
		t.Fatalf("Content-Type = %q", got)
	}
	if got := rec.Header().Get("Cache-Control"); got != "no-store" {
		t.Fatalf("Cache-Control = %q", got)
	}
	if got := rec.Header().Get("X-Request-Id"); got == "" {
		t.Fatal("missing X-Request-Id")
	}

	var body errorEnvelope
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode: %v (%s)", err, rec.Body.String())
	}
	if body.Error.Code != CodeValidationError || body.Error.Message == "" {
		t.Fatalf("error body = %+v", body.Error)
	}
	if body.Error.RequestID == "" || body.Error.RequestID != rec.Header().Get("X-Request-Id") {
		t.Fatalf("requestId %q does not match the header", body.Error.RequestID)
	}
	if body.Error.Fields["identifier"] != "Enter your username or email." {
		t.Fatalf("field message = %q", body.Error.Fields["identifier"])
	}
}

// TestUnknownRouteAndWrongVerb_AreJSON checks the router never answers HTML:
// an unknown path is NOT_FOUND and a known path with the wrong verb is
// METHOD_NOT_ALLOWED.
func TestUnknownRouteAndWrongVerb_AreJSON(t *testing.T) {
	s := newTestServerObject(t, "https://clover.example")
	h := s.Handler()

	for _, tc := range []struct {
		method, path, code string
		status             int
	}{
		{http.MethodGet, "/nope", CodeNotFound, http.StatusNotFound},
		{http.MethodPost, "/v1/movies", CodeMethodNotAllowed, http.StatusMethodNotAllowed},
		{http.MethodDelete, "/v1/auth/login", CodeMethodNotAllowed, http.StatusMethodNotAllowed},
	} {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(tc.method, tc.path, nil))
		if rec.Code != tc.status {
			t.Fatalf("%s %s: status = %d, want %d", tc.method, tc.path, rec.Code, tc.status)
		}
		if ct := rec.Header().Get("Content-Type"); !strings.HasPrefix(ct, "application/json") {
			t.Fatalf("%s %s: Content-Type = %q", tc.method, tc.path, ct)
		}
		var body errorEnvelope
		if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
			t.Fatalf("%s %s: body is not JSON: %s", tc.method, tc.path, rec.Body.String())
		}
		if body.Error.Code != tc.code {
			t.Fatalf("%s %s: code = %q, want %q", tc.method, tc.path, body.Error.Code, tc.code)
		}
	}
}

// TestHealthz_NeedsNoDatabase covers liveness (architecture.md, Endpoints).
func TestHealthz_NeedsNoDatabase(t *testing.T) {
	s := newTestServerObject(t, "https://clover.example")
	rec := httptest.NewRecorder()
	s.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/healthz", nil))
	if rec.Code != http.StatusOK || rec.Body.String() != "ok\n" {
		t.Fatalf("healthz = %d %q", rec.Code, rec.Body.String())
	}
}

// TestClientIP covers the rate-limit key.
func TestClientIP(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/v1/auth/login", nil)
	req.RemoteAddr = "203.0.113.9:51721"
	if got := clientIP(req); got != "203.0.113.9" {
		t.Fatalf("clientIP = %q", got)
	}
	req.RemoteAddr = "not-an-address"
	if got := clientIP(req); got != "not-an-address" {
		t.Fatalf("clientIP fallback = %q", got)
	}
}

// newTestServerObject builds a Server with no database. Only handlers that
// never touch the pool may be called on it.
func newTestServerObject(t *testing.T, publicOrigin string) *Server {
	t.Helper()
	return New(Deps{
		Config:  config.Config{PublicOrigin: publicOrigin},
		Metrics: observability.New(),
		Logger:  slog.New(slog.NewTextHandler(io.Discard, nil)),
	})
}
