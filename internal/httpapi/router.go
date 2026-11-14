// Package httpapi is the API surface: router, middlewares, handlers and the
// error envelope. It returns JSON and media bytes only — never HTML.
package httpapi

import (
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/william22913/clover-be/internal/config"
	"github.com/william22913/clover-be/internal/observability"
)

// apiPrefix is fixed, not configuration: the frontend proxy and the OpenAPI
// contract both pin it.
const apiPrefix = "/v1"

// Deps is everything the API needs from the process.
type Deps struct {
	Pool    *pgxpool.Pool
	Config  config.Config
	Metrics *observability.Metrics
	Logger  *slog.Logger
}

// Server holds the dependencies shared by the handlers.
type Server struct {
	pool         *pgxpool.Pool
	cfg          config.Config
	publicOrigin string
	mediaDir     string
	metrics      *observability.Metrics
	logger       *slog.Logger
	loginLimiter *rateLimiter
}

// New builds the API server.
func New(deps Deps) *Server {
	return &Server{
		pool:         deps.Pool,
		cfg:          deps.Config,
		publicOrigin: deps.Config.PublicOrigin,
		mediaDir:     deps.Config.MediaDir,
		metrics:      deps.Metrics,
		logger:       loggerOrDefault(deps.Logger),
		loginLimiter: newRateLimiter(30, 30),
	}
}

// Handler returns the fully wired API handler.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()

	// Public catalogue and posters.
	mux.Handle("GET /v1/movies", http.HandlerFunc(s.listMovies))
	mux.Handle("GET /v1/movies/{movieId}", http.HandlerFunc(s.getMovie))
	mux.Handle("GET /v1/media/movies/{movieId}/poster", http.HandlerFunc(s.getPoster))

	// Sign-in, sign-out and the signed-in account.
	mux.Handle("POST /v1/auth/login", s.sameOrigin(s.limitLogin(http.HandlerFunc(s.login))))
	mux.Handle("POST /v1/auth/logout", s.sameOrigin(http.HandlerFunc(s.logout)))
	mux.Handle("GET /v1/auth/me", http.HandlerFunc(s.me))

	// Session-gated video. A GET pattern also matches HEAD, so http.ServeContent
	// answers HEAD without a second registration.
	mux.Handle("GET /v1/media/movies/{movieId}/video", http.HandlerFunc(s.getVideo))

	// Every known path answers a wrong verb with the JSON 405 envelope.
	for _, path := range []string{
		"/v1/movies",
		"/v1/movies/{movieId}",
		"/v1/auth/login",
		"/v1/auth/logout",
		"/v1/auth/me",
		"/v1/media/movies/{movieId}/video",
		"/v1/media/movies/{movieId}/poster",
	} {
		mux.Handle(path, http.HandlerFunc(s.methodNotAllowed))
	}

	// Operational routes stay at the root.
	mux.Handle("GET /healthz", http.HandlerFunc(s.healthz))
	mux.Handle("GET /readyz", http.HandlerFunc(s.readyz))

	// Any other path is the API's NOT_FOUND.
	mux.Handle("/", http.HandlerFunc(s.notFound))

	var h http.Handler = mux
	h = s.metrics.Middleware(h)
	h = s.withSession(h)
	h = s.logRequests(h)
	h = securityHeaders(h)
	h = s.recoverPanic(h)
	h = withRequestID(h)
	return h
}

// limitLogin applies the per-IP sign-in bucket and answers 429 when it is
// exhausted.
func (s *Server) limitLogin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ok, wait := s.loginLimiter.allow(clientIP(r))
		if !ok {
			w.Header().Set("Retry-After", retryAfterSeconds(wait))
			s.writeError(w, r, http.StatusTooManyRequests, CodeRateLimited, nil)
			return
		}
		s.loginLimiter.cleanup()
		next.ServeHTTP(w, r)
	})
}

func retryAfterSeconds(wait time.Duration) string {
	seconds := int(wait.Seconds())
	if seconds < 1 {
		seconds = 1
	}
	return strconv.Itoa(seconds)
}

// notFound answers an unknown API route with the envelope.
func (s *Server) notFound(w http.ResponseWriter, r *http.Request) {
	s.writeError(w, r, http.StatusNotFound, CodeNotFound, nil)
}

// methodNotAllowed answers a known path with an unsupported verb.
func (s *Server) methodNotAllowed(w http.ResponseWriter, r *http.Request) {
	s.writeError(w, r, http.StatusMethodNotAllowed, CodeMethodNotAllowed, nil)
}
