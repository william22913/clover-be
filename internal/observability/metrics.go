// Package observability owns the Prometheus registry, the HTTP metrics
// middleware and the internal metrics listener. The design records one scrape
// endpoint and two business counters (architecture.md, Observability).
package observability

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Metrics holds every collector the service registers.
type Metrics struct {
	registry     *prometheus.Registry
	requests     *prometheus.CounterVec
	duration     *prometheus.HistogramVec
	errors       *prometheus.CounterVec
	logins       *prometheus.CounterVec
	videoRequest *prometheus.CounterVec
}

// New builds the registry and registers every collector once per process.
func New() *Metrics {
	m := &Metrics{
		registry: prometheus.NewRegistry(),
		requests: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "HTTP requests by registered route, method and response status.",
		}, []string{"route", "method", "status"}),
		duration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration by registered route and method.",
			Buckets: prometheus.DefBuckets,
		}, []string{"route", "method"}),
		errors: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "http_errors_total",
			Help: "Error responses by stable error code.",
		}, []string{"code"}),
		logins: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "clover_logins_total",
			Help: "Sign-in attempts by result.",
		}, []string{"result"}),
		videoRequest: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "clover_video_requests_total",
			Help: "Video stream requests by result.",
		}, []string{"result"}),
	}
	m.registry.MustRegister(m.requests, m.duration, m.errors, m.logins, m.videoRequest)
	return m
}

// Handler returns the Prometheus exposition handler.
func (m *Metrics) Handler() http.Handler {
	return promhttp.HandlerFor(m.registry, promhttp.HandlerOpts{})
}

// ObserveRequest records one finished request. route must be the registered
// pattern, never the raw path, so ids never become label values.
func (m *Metrics) ObserveRequest(route, method string, status int, elapsed time.Duration) {
	m.requests.WithLabelValues(route, method, strconv.Itoa(status)).Inc()
	m.duration.WithLabelValues(route, method).Observe(elapsed.Seconds())
}

// ObserveError records one error response by its stable code.
func (m *Metrics) ObserveError(code string) {
	m.errors.WithLabelValues(code).Inc()
}

// ObserveLogin records a sign-in result ("ok" or "failed").
func (m *Metrics) ObserveLogin(result string) {
	m.logins.WithLabelValues(result).Inc()
}

// ObserveVideo records a video route result ("ok", "unauthorized" or
// "not_found").
func (m *Metrics) ObserveVideo(result string) {
	m.videoRequest.WithLabelValues(result).Inc()
}

// statusRecorder captures the status code written by a handler.
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (s *statusRecorder) WriteHeader(code int) {
	if s.status == 0 {
		s.status = code
	}
	s.ResponseWriter.WriteHeader(code)
}

func (s *statusRecorder) Write(b []byte) (int, error) {
	if s.status == 0 {
		s.status = http.StatusOK
	}
	return s.ResponseWriter.Write(b)
}

// Middleware records status, route label and duration for every request. It is
// mounted inside the request-id middleware so the log line and the sample agree.
func (m *Metrics) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w}
		next.ServeHTTP(rec, r)
		status := rec.status
		if status == 0 {
			status = http.StatusOK
		}
		route := r.Pattern
		if route == "" {
			route = "unmatched"
		}
		m.ObserveRequest(route, r.Method, status, time.Since(start))
	})
}
