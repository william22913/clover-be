package api_test

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/william22913/clover-be/internal/metrics"
)

// The security design's response headers are on every response, including
// errors.
func TestMiddleware_SecurityHeaders(t *testing.T) {
	env := newEnv(t)
	for _, path := range []string{"/", "/v1/movies", "/v1/movies/unlisted"} {
		res := env.do(http.MethodGet, path, "", nil)
		if got := res.Header.Get("Content-Security-Policy"); !strings.Contains(got, "default-src 'self'") ||
			!strings.Contains(got, "object-src 'none'") || !strings.Contains(got, "frame-ancestors 'none'") {
			t.Fatalf("%s CSP = %q", path, got)
		}
		if res.Header.Get("X-Content-Type-Options") != "nosniff" {
			t.Fatalf("%s missing nosniff", path)
		}
		if res.Header.Get("X-Frame-Options") != "DENY" {
			t.Fatalf("%s missing X-Frame-Options", path)
		}
		if res.Header.Get("Referrer-Policy") != "no-referrer" {
			t.Fatalf("%s missing Referrer-Policy", path)
		}
		if len(res.Header.Get("X-Request-Id")) != 32 {
			t.Fatalf("%s X-Request-Id = %q, want 32 hex characters", path, res.Header.Get("X-Request-Id"))
		}
	}
}

// A cross-site POST is refused and the session survives: the refusal is the
// boundary, not a sign-out.
func TestMiddleware_CSRFRefusesCrossSitePost(t *testing.T) {
	env := newEnv(t)
	cookie := signUpCookie(t, env, "viewer@example.com", "Passw0rd1")
	header := map[string]string{"Cookie": cookieHeader(cookie)}

	for name, headers := range map[string]map[string]string{
		"origin":   {"Origin": "https://evil.example"},
		"site":     {"Sec-Fetch-Site": "cross-site"},
		"same-ok?": {"Origin": "http://evil.example"},
	} {
		merge := map[string]string{"Cookie": header["Cookie"]}
		for k, v := range headers {
			merge[k] = v
		}
		res := env.do(http.MethodPost, "/v1/auth/sign-out", "", merge)
		if res.StatusCode != http.StatusForbidden {
			t.Fatalf("%s status = %d, want 403", name, res.StatusCode)
		}
		var body errorResponse
		decodeBody(t, res, &body)
		if body.Error.Code != "invalid_request" {
			t.Fatalf("%s code = %q, want invalid_request", name, body.Error.Code)
		}
	}

	me := env.do(http.MethodGet, "/v1/auth/me", "", header)
	if me.StatusCode != http.StatusOK {
		t.Fatalf("session after refused CSRF status = %d, want 200", me.StatusCode)
	}
}

// A same-origin Origin header is accepted.
func TestMiddleware_SameOriginPostAllowed(t *testing.T) {
	env := newEnv(t)
	cookie := signUpCookie(t, env, "viewer@example.com", "Passw0rd1")
	res := env.do(http.MethodPost, "/v1/auth/sign-out", "", map[string]string{
		"Cookie": cookieHeader(cookie),
		"Origin": env.server.URL,
	})
	if res.StatusCode != http.StatusNoContent {
		t.Fatalf("status = %d, want 204", res.StatusCode)
	}
}

// A wrong method on a known path is the JSON 405 with Allow.
func TestMiddleware_MethodNotAllowed(t *testing.T) {
	env := newEnv(t)
	res := env.do(http.MethodDelete, "/v1/movies", "", nil)
	if res.StatusCode != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want 405", res.StatusCode)
	}
	if allow := res.Header.Get("Allow"); allow != "GET, HEAD" {
		t.Fatalf("Allow = %q", allow)
	}
	var body errorResponse
	decodeBody(t, res, &body)
	if body.Error.Code != "method_not_allowed" {
		t.Fatalf("code = %q", body.Error.Code)
	}
}

// Every browser-called endpoint answers its own OPTIONS preflight, which
// matters for same-origin JSON POSTs and is required by the delivery guard.
func TestMiddleware_OptionsPreflight(t *testing.T) {
	env := newEnv(t)
	for path, want := range map[string]string{
		"/v1/movies":               "GET, HEAD, OPTIONS",
		"/v1/auth/sign-in":         "POST, OPTIONS",
		"/v1/auth/me":              "GET, HEAD, OPTIONS",
		"/v1/movies/halcyon":       "GET, HEAD, OPTIONS",
		"/play/halcyon":            "GET, HEAD, OPTIONS",
		"/v1/movies/halcyon/video": "GET, HEAD, OPTIONS",
	} {
		res := env.do(http.MethodOptions, path, "", nil)
		if res.StatusCode != http.StatusNoContent {
			t.Fatalf("%s status = %d, want 204", path, res.StatusCode)
		}
		if allow := res.Header.Get("Allow"); allow != want {
			t.Fatalf("%s Allow = %q, want %q", path, allow, want)
		}
	}
}

// Unknown API routes and the public /metrics path answer the JSON 404 envelope;
// metrics are never exposed on the public listener.
func TestMiddleware_NotFoundEnvelope(t *testing.T) {
	env := newEnv(t)
	for _, path := range []string{"/v1/unknown", "/metrics"} {
		res := env.do(http.MethodGet, path, "", nil)
		if res.StatusCode != http.StatusNotFound {
			t.Fatalf("%s status = %d, want 404", path, res.StatusCode)
		}
		var body errorResponse
		decodeBody(t, res, &body)
		if body.Error.Code != "not_found" {
			t.Fatalf("%s code = %q, want not_found", path, body.Error.Code)
		}
	}
}

// A traversal path is refused rather than redirected.
func TestMiddleware_TraversalRefused(t *testing.T) {
	env := newEnv(t)
	// The encoded slashes survive the client so the server really receives a
	// ".." segment; both paths must answer the ordinary JSON 404 envelope.
	for _, path := range []string{"/media/poster/../../web/index.html", "/v1/movies/..%2F..%2Fetc"} {
		res, err := env.server.Client().Get(env.server.URL + path)
		if err != nil {
			t.Fatal(err)
		}
		var body struct {
			Error struct {
				Code string `json:"code"`
			} `json:"error"`
		}
		decodeErr := json.NewDecoder(res.Body).Decode(&body)
		res.Body.Close()
		if res.StatusCode != http.StatusNotFound {
			t.Fatalf("%s: status = %d, want 404", path, res.StatusCode)
		}
		if decodeErr != nil || body.Error.Code != "not_found" {
			t.Fatalf("%s: body = %+v (decode err %v), want a not_found envelope", path, body, decodeErr)
		}
	}
}

// The UI shell is served for the client-side routes so a deep link or refresh
// works without the server knowing the route.
func TestMiddleware_ShellServesDeepLinks(t *testing.T) {
	env := newEnv(t)
	for _, path := range []string{"/", "/auth", "/play/halcyon"} {
		res := env.do(http.MethodGet, path, "", nil)
		if res.StatusCode != http.StatusOK {
			t.Fatalf("%s status = %d, want 200", path, res.StatusCode)
		}
		body, _ := io.ReadAll(res.Body)
		res.Body.Close()
		if !strings.Contains(string(body), "Clover") {
			t.Fatalf("%s did not serve the shell", path)
		}
	}
}

// Security design — PII in logs: a full sign-up/sign-in cycle leaves an email,
// a password, a hash or a token in no log line, while the trace fields are all
// there.
func TestMiddleware_LogsCarryNoPII(t *testing.T) {
	var sink bytes.Buffer
	env := newEnvWithLogger(t, slog.New(slog.NewTextHandler(&sink, nil)))

	cookie := signUpCookie(t, env, "viewer@example.com", "Passw0rd1")
	env.do(http.MethodPost, "/v1/auth/sign-in",
		jsonBody(map[string]any{"email": "viewer@example.com", "password": "Passw0rd1"}),
		map[string]string{"Cookie": cookieHeader(cookie)})
	env.do(http.MethodGet, "/v1/auth/me", "", map[string]string{"Cookie": cookieHeader(cookie)})

	logs := sink.String()
	if logs == "" {
		t.Fatal("no request log lines were written")
	}
	for _, secret := range []string{"viewer@example.com", "Passw0rd1", "$2a$", cookie.Value} {
		if strings.Contains(logs, secret) {
			t.Fatalf("a log line leaked %q", secret)
		}
	}
	for _, field := range []string{"request_id=", "route=", "status=", "duration_ms=", "user_id="} {
		if !strings.Contains(logs, field) {
			t.Fatalf("log lines are missing %s", field)
		}
	}
}

// Every public request is counted with its route template and status, and the
// error codes are counted too — this is the Observability decision.
func TestMiddleware_MetricsRecorded(t *testing.T) {
	env := newEnv(t)
	env.do(http.MethodGet, "/v1/movies", "", nil)  // 200
	env.do(http.MethodGet, "/v1/auth/me", "", nil) // 401 session_expired

	rec := httptest.NewRecorder()
	metrics.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("metrics status = %d", rec.Code)
	}
	text := rec.Body.String()
	for _, want := range []string{
		`clover_http_requests_total{method="GET",route="/v1/movies",status="200"}`,
		`clover_http_requests_total{method="GET",route="/v1/auth/me",status="401"}`,
		`clover_http_errors_total{code="session_expired"}`,
		`clover_http_request_duration_seconds_bucket`,
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("metrics output missing %s", want)
		}
	}
}
