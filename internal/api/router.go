// Package api holds every route, the middleware chain, the error envelope and
// the handlers.
package api

import (
	"context"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/william22913/clover-be/internal/auth"
	"github.com/william22913/clover-be/internal/media"
	"github.com/william22913/clover-be/internal/metrics"
	"github.com/william22913/clover-be/internal/store"
)

// Store is the persistence surface the handlers need.
type Store interface {
	ListMovies(ctx context.Context) ([]store.Movie, error)
	GetMovie(ctx context.Context, slug string) (store.Movie, error)
	CreateUserAndSession(ctx context.Context, email, passwordHash, tokenHash string) (int64, error)
	UserByEmail(ctx context.Context, email string) (store.User, error)
	CreateSession(ctx context.Context, userID int64, tokenHash, revokeTokenHash string) error
	ResolveSession(ctx context.Context, tokenHash string) (store.SessionUser, error)
	RevokeSession(ctx context.Context, tokenHash string) (int64, error)
	Ping(ctx context.Context) error
}

// Handler is the public HTTP surface.
type Handler struct {
	store        Store
	media        *media.Server
	limiter      *auth.Limiter
	publicOrigin string
	log          *slog.Logger
}

// Options configures the router.
type Options struct {
	Store        Store
	Media        *media.Server
	PublicOrigin string
	Logger       *slog.Logger
}

// New builds the whole public handler with the middleware chain in the fixed
// order: request id, log, metrics, security headers, CSRF, route, session.
func New(opts Options) http.Handler {
	h := &Handler{
		store:        opts.Store,
		media:        opts.Media,
		limiter:      auth.NewLimiter(10, int64(time.Minute)),
		publicOrigin: opts.PublicOrigin,
		log:          opts.Logger,
	}

	mux := http.NewServeMux()
	mux.Handle("/v1/movies", h.methods([]string{http.MethodGet, http.MethodHead}, h.listMovies))
	mux.Handle("/v1/movies/{slug}", h.methods([]string{http.MethodGet, http.MethodHead}, h.getMovie))
	mux.Handle("/v1/movies/{slug}/video",
		h.methods([]string{http.MethodGet, http.MethodHead}, h.withSession(h.video)))
	mux.Handle("/v1/auth/sign-up", h.methods([]string{http.MethodPost}, h.signUp))
	mux.Handle("/v1/auth/sign-in", h.methods([]string{http.MethodPost}, h.signIn))
	mux.Handle("/v1/auth/sign-out", h.methods([]string{http.MethodPost}, h.signOut))
	mux.Handle("/v1/auth/me", h.methods([]string{http.MethodGet, http.MethodHead}, h.withSession(h.me)))
	mux.Handle("/healthz", h.methods([]string{http.MethodGet, http.MethodHead}, h.healthz))
	mux.Handle("/v1/", http.HandlerFunc(h.notFound))
	// Metrics live on the internal listener only; the public port answers the
	// ordinary JSON 404 envelope.
	mux.Handle("/metrics", http.HandlerFunc(h.notFound))
	mux.Handle("/static/", h.media.StaticHandler())
	mux.Handle("/media/poster/", h.media.PosterHandler())
	mux.Handle("/media/still/", h.media.StillHandler())
	mux.Handle("/", http.HandlerFunc(h.shell))

	var handler http.Handler = mux
	handler = h.withCSRF(handler)
	handler = h.rejectDirtyPaths(handler)
	handler = withSecurityHeaders(handler)
	handler = metrics.Middleware(handler)
	handler = h.withLogging(handler)
	handler = h.withRequestID(handler)
	return handler
}

// methods answers the CORS preflight itself and turns a wrong method on a
// known path into the JSON 405 with an Allow header.
func (h *Handler) methods(allowed []string, fn http.HandlerFunc) http.HandlerFunc {
	allow := strings.Join(append(append([]string{}, allowed...), http.MethodOptions), ", ")
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			w.Header().Set("Allow", allow)
			w.WriteHeader(http.StatusNoContent)
			return
		}
		for _, m := range allowed {
			if r.Method == m {
				fn(w, r)
				return
			}
		}
		w.Header().Set("Allow", strings.Join(allowed, ", "))
		h.fail(w, r, http.StatusMethodNotAllowed, codeMethodNotAllowed,
			"That method is not allowed here.", nil)
	}
}

// shell serves the UI for /, /auth, /play/{slug} and any other unknown GET.
func (h *Handler) shell(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.Header().Set("Allow", "GET, HEAD, OPTIONS")
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		w.Header().Set("Allow", "GET, HEAD")
		h.fail(w, r, http.StatusMethodNotAllowed, codeMethodNotAllowed,
			"That method is not allowed here.", nil)
		return
	}
	w.Header().Set("Cache-Control", "no-cache")
	http.ServeFile(w, r, h.media.IndexPath())
}
