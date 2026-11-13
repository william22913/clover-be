// Package metrics owns the Prometheus collectors and the middleware that
// records every public request.
package metrics

import (
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	requestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "clover_http_requests_total",
		Help: "Public HTTP requests by method, route template and status.",
	}, []string{"method", "route", "status"})

	requestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "clover_http_request_duration_seconds",
		Help:    "Public HTTP request duration by method, route template and status.",
		Buckets: prometheus.DefBuckets,
	}, []string{"method", "route", "status"})

	errorsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "clover_http_errors_total",
		Help: "Error responses by stable error code.",
	}, []string{"code"})

	signInsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "clover_signins_total",
		Help: "Sign-in attempts by result.",
	}, []string{"result"})

	signUpsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "clover_signups_total",
		Help: "Sign-up attempts by result.",
	}, []string{"result"})

	sessionsRevokedTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "clover_sessions_revoked_total",
		Help: "Sessions revoked by sign-out.",
	})
)

// Handler serves the Prometheus text format. It is mounted on the internal
// listener only.
func Handler() http.Handler { return promhttp.Handler() }

// Error records one error response by its stable code.
func Error(code string) { errorsTotal.WithLabelValues(code).Inc() }

// SignIn records a sign-in attempt result: ok, failed or rate_limited.
func SignIn(result string) { signInsTotal.WithLabelValues(result).Inc() }

// SignUp records a sign-up attempt result: ok, duplicate or invalid.
func SignUp(result string) { signUpsTotal.WithLabelValues(result).Inc() }

// SessionRevoked records one revoked session.
func SessionRevoked() { sessionsRevokedTotal.Inc() }

// Middleware records the status actually written and labels by route template.
func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(sw, r)

		route := r.Pattern
		if route == "" || route == "/" {
			route = "unmatched"
		}
		status := strconv.Itoa(sw.status)
		requestsTotal.WithLabelValues(r.Method, route, status).Inc()
		requestDuration.WithLabelValues(r.Method, route, status).Observe(time.Since(start).Seconds())
	})
}

// statusWriter captures the first status written and keeps the optional
// interfaces http.ServeContent relies on, so range streaming is untouched.
type statusWriter struct {
	http.ResponseWriter
	status int
	wrote  bool
}

func (w *statusWriter) WriteHeader(code int) {
	if !w.wrote {
		w.status = code
		w.wrote = true
	}
	w.ResponseWriter.WriteHeader(code)
}

func (w *statusWriter) Write(b []byte) (int, error) {
	w.wrote = true
	return w.ResponseWriter.Write(b)
}

func (w *statusWriter) ReadFrom(r io.Reader) (int64, error) {
	if rf, ok := w.ResponseWriter.(io.ReaderFrom); ok {
		return rf.ReadFrom(r)
	}
	return io.Copy(w.ResponseWriter, r)
}

func (w *statusWriter) Flush() {
	if f, ok := w.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

func (w *statusWriter) Unwrap() http.ResponseWriter { return w.ResponseWriter }
