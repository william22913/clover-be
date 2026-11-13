package api

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/william22913/clover-be/internal/auth"
	"github.com/william22913/clover-be/internal/store"
)

// contentSecurityPolicy is the exact header from the security design.
const contentSecurityPolicy = "default-src 'self'; script-src 'self'; style-src 'self'; " +
	"img-src 'self' data:; media-src 'self'; font-src 'self'; connect-src 'self'; " +
	"object-src 'none'; base-uri 'none'; frame-ancestors 'none'; form-action 'self'"

// withRequestID is the outermost middleware: it issues X-Request-Id and puts
// the shared requestInfo in the context.
func (h *Handler) withRequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		info := &requestInfo{id: newRequestID()}
		w.Header().Set("X-Request-Id", info.id)
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), requestInfoKey, info)))
	})
}

// withLogging writes exactly one line per request.
func (h *Handler) withLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		lw := &logWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(lw, r)

		route := r.Pattern
		if route == "" || route == "/" {
			route = "unmatched"
		}
		var uid int64
		if info, ok := r.Context().Value(requestInfoKey).(*requestInfo); ok {
			uid = info.userID
		}
		h.log.InfoContext(r.Context(), "request",
			"request_id", requestIDFrom(r.Context()),
			"method", r.Method,
			"route", route,
			"status", lw.status,
			"duration_ms", time.Since(start).Milliseconds(),
			"user_id", uid,
		)
	})
}

// withSecurityHeaders sets the fixed response headers from the security design.
func withSecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Security-Policy", contentSecurityPolicy)
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "no-referrer")
		next.ServeHTTP(w, r)
	})
}

// withCSRF rejects cross-site POSTs. SameSite=Lax already keeps the cookie off
// cross-site form posts; this closes the gap for browsers that send it.
func (h *Handler) withCSRF(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && !h.originAllowed(r) {
			h.fail(w, r, http.StatusForbidden, codeInvalidRequest,
				"Something went wrong. Please try again.", nil)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (h *Handler) originAllowed(r *http.Request) bool {
	if site := r.Header.Get("Sec-Fetch-Site"); site != "" && site != "same-origin" {
		return false
	}
	origin := r.Header.Get("Origin")
	if origin == "" {
		return true
	}
	want := h.publicOrigin
	if want == "" {
		scheme := "http"
		if r.TLS != nil {
			scheme = "https"
		}
		want = scheme + "://" + r.Host
	}
	return strings.TrimSuffix(origin, "/") == strings.TrimSuffix(want, "/")
}

// withSession resolves the cookie to a live session or answers 401 first.
func (h *Handler) withSession(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(sessionCookieName)
		if err != nil || cookie.Value == "" {
			h.fail(w, r, http.StatusUnauthorized, codeSessionExpired,
				"Your session ended. Sign in again to keep watching.", nil)
			return
		}
		user, err := h.store.ResolveSession(r.Context(), auth.HashToken(cookie.Value))
		if errors.Is(err, store.ErrNotFound) {
			h.fail(w, r, http.StatusUnauthorized, codeSessionExpired,
				"Your session ended. Sign in again to keep watching.", nil)
			return
		}
		if err != nil {
			h.internal(w, r, err)
			return
		}
		if info, ok := r.Context().Value(requestInfoKey).(*requestInfo); ok {
			info.userID = user.UserID
		}
		next(w, r.WithContext(context.WithValue(r.Context(), sessionUserKey, user)))
	}
}

// rejectDirtyPaths refuses any request path containing a traversal segment
// before the mux can redirect it.
func (h *Handler) rejectDirtyPaths(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "..") || strings.ContainsRune(r.URL.Path, 0) {
			h.notFound(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (h *Handler) internal(w http.ResponseWriter, r *http.Request, cause error) {
	h.log.ErrorContext(r.Context(), "request failed",
		"request_id", requestIDFrom(r.Context()), "error", cause)
	h.fail(w, r, http.StatusInternalServerError, codeInternalError,
		"Something went wrong. Please try again.", nil)
}

func newRequestID() string {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return time.Now().UTC().Format("20060102150405.000000000")
	}
	return hex.EncodeToString(buf)
}

// clientIP is the TCP peer address only; X-Forwarded-For is never read.
func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// logWriter captures the status for the request line.
type logWriter struct {
	http.ResponseWriter
	status int
	wrote  bool
}

func (w *logWriter) WriteHeader(code int) {
	if !w.wrote {
		w.status = code
		w.wrote = true
	}
	w.ResponseWriter.WriteHeader(code)
}

func (w *logWriter) Write(b []byte) (int, error) {
	w.wrote = true
	return w.ResponseWriter.Write(b)
}

func (w *logWriter) ReadFrom(r io.Reader) (int64, error) {
	if rf, ok := w.ResponseWriter.(io.ReaderFrom); ok {
		return rf.ReadFrom(r)
	}
	return io.Copy(w.ResponseWriter, r)
}

func (w *logWriter) Flush() {
	if f, ok := w.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

func (w *logWriter) Unwrap() http.ResponseWriter { return w.ResponseWriter }
