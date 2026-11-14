package httpapi

import (
	"context"
	"net"
	"net/http"
	"net/netip"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"github.com/william22913/clover-be/internal/sessions"
)

// identity is what the session cookie resolved to. The zero value is a
// visitor: no valid session.
type identity struct {
	Role     Role
	Account  string
	Username string
	Valid    bool
}

const identityKey contextKey = 100

func identityFrom(ctx context.Context) identity {
	id, _ := ctx.Value(identityKey).(identity)
	return id
}

// withSession resolves the cookie once per request and stores the identity in
// the context. An unknown or expired cookie is simply anonymous; the routes
// that need a session answer 401 themselves.
func (s *Server) withSession(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(sessionCookieName)
		if err == nil && cookie.Value != "" {
			sess, err := sessions.Resolve(r.Context(), s.pool, cookie.Value)
			if err == nil {
				if sess.Slid {
					// The row's expiry was extended; the browser's copy
					// must move with it.
					s.setSessionCookie(w, cookie.Value)
				}
				r = r.WithContext(context.WithValue(r.Context(), identityKey, identity{
					Role:     RoleMember,
					Account:  sess.Account.String(),
					Username: sess.Username,
					Valid:    true,
				}))
			} else if err != sessions.ErrInvalid {
				s.logger.Error("session lookup failed", "request_id", requestID(r), "error", err)
			}
		}
		next.ServeHTTP(w, r)
	})
}

// authorize enforces one route's permission. A request with no valid session
// holds the visitor role, which is how the public catalogue stays public; every
// route that needs more than the visitor role declares its Permission constant
// here.
func (s *Server) authorize(w http.ResponseWriter, r *http.Request, required Permission) bool {
	role := RoleVisitor
	if id := identityFrom(r.Context()); id.Valid {
		role = id.Role
	}
	if role.Allows(required) {
		return true
	}
	s.writeError(w, r, http.StatusUnauthorized, CodeUnauthenticated, nil)
	return false
}

// securityHeaders sets the response headers the API owns. The frontend serves
// the CSP for the pages; these cover API and media responses.
func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("X-Frame-Options", "DENY")
		h.Set("Referrer-Policy", "strict-origin-when-cross-origin")
		next.ServeHTTP(w, r)
	})
}

// logRequests writes one structured line per request carrying the request id.
func (s *Server) logRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusWriter{ResponseWriter: w}
		next.ServeHTTP(rec, r)
		status := rec.status
		if status == 0 {
			status = http.StatusOK
		}
		route := r.Pattern
		if route == "" {
			route = "unmatched"
		}
		s.logger.Info("http request",
			"request_id", requestID(r),
			"method", r.Method,
			"route", route,
			"status", status,
			"duration_ms", time.Since(start).Milliseconds(),
			"remote_ip", s.clientIP(r),
		)
	})
}

// recoverPanic turns a handler panic into the generic 500 envelope.
func (s *Server) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				s.logger.Error("panic recovered",
					"request_id", requestID(r),
					"route", r.Pattern,
					"panic", rec,
					"stack", string(debug.Stack()),
				)
				s.writeError(w, r, http.StatusInternalServerError, CodeInternalError, nil)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// sameOrigin rejects a POST whose Origin or Sec-Fetch-Site is present and not
// this origin (S8). A request carrying neither header is not a browser POST and
// is left to the permission checks.
func (s *Server) sameOrigin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			if origin := r.Header.Get("Origin"); origin != "" && origin != s.publicOrigin {
				s.writeError(w, r, http.StatusForbidden, CodeCSRFRejected, nil)
				return
			}
			if site := r.Header.Get("Sec-Fetch-Site"); site != "" && site != "same-origin" && site != "none" {
				s.writeError(w, r, http.StatusForbidden, CodeCSRFRejected, nil)
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}

// statusWriter captures the status code without changing the response body.
type statusWriter struct {
	http.ResponseWriter
	status int
}

func (s *statusWriter) WriteHeader(code int) {
	if s.status == 0 {
		s.status = code
	}
	s.ResponseWriter.WriteHeader(code)
}

func (s *statusWriter) Write(b []byte) (int, error) {
	if s.status == 0 {
		s.status = http.StatusOK
	}
	return s.ResponseWriter.Write(b)
}

// clientIP returns the address the sign-in bucket and the log line key on.
// Forwarding headers are believed only when the immediate peer is a configured
// trusted proxy, so a caller reaching the API directly cannot spoof them. The
// chain is read from the right: the proxy appends the address it observed, so
// anything a client prepends is ignored.
func (s *Server) clientIP(r *http.Request) string {
	peer := remoteIP(r.RemoteAddr)
	if !s.isTrustedProxy(peer) {
		return peer
	}
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		parts := strings.Split(forwarded, ",")
		for i := len(parts) - 1; i >= 0; i-- {
			addr, err := netip.ParseAddr(strings.TrimSpace(parts[i]))
			if err != nil {
				continue
			}
			if s.isTrustedProxy(addr.String()) {
				continue
			}
			return addr.String()
		}
	}
	return peer
}

func remoteIP(remoteAddr string) string {
	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		return remoteAddr
	}
	return host
}

func (s *Server) isTrustedProxy(ip string) bool {
	addr, err := netip.ParseAddr(ip)
	if err != nil {
		return false
	}
	for _, prefix := range s.trustedProxies {
		if prefix.Contains(addr) {
			return true
		}
	}
	return false
}

// rateLimiter is the per-IP token bucket on sign-in (S9). It is correct for
// one replica, which is what app_scale small gives v1.
type rateLimiter struct {
	mu      sync.Mutex
	buckets map[string]*bucket
	rate    float64
	burst   float64
	now     func() time.Time
}

type bucket struct {
	tokens float64
	last   time.Time
}

func newRateLimiter(perMinute float64, burst float64) *rateLimiter {
	return &rateLimiter{
		buckets: make(map[string]*bucket),
		rate:    perMinute / 60,
		burst:   burst,
		now:     time.Now,
	}
}

// allow consumes one token for key, reporting whether the request may go on
// and, when refused, how long the caller should wait.
func (l *rateLimiter) allow(key string) (bool, time.Duration) {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := l.now()
	b, ok := l.buckets[key]
	if !ok || now.Sub(b.last) > 10*time.Minute {
		b = &bucket{tokens: l.burst, last: now}
		l.buckets[key] = b
	}
	b.tokens += now.Sub(b.last).Seconds() * l.rate
	if b.tokens > l.burst {
		b.tokens = l.burst
	}
	b.last = now

	if b.tokens < 1 {
		wait := time.Duration((1 - b.tokens) / l.rate * float64(time.Second))
		return false, wait
	}
	b.tokens--
	return true, 0
}

// cleanup drops idle buckets occasionally so the map cannot grow without bound.
func (l *rateLimiter) cleanup() {
	l.mu.Lock()
	defer l.mu.Unlock()
	if len(l.buckets) < 4096 {
		return
	}
	cutoff := l.now().Add(-10 * time.Minute)
	for key, b := range l.buckets {
		if b.last.Before(cutoff) {
			delete(l.buckets, key)
		}
	}
}
