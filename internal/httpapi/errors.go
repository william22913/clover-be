package httpapi

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/william22913/clover-be/internal/i18n"
)

// Stable error codes. They are the only codes the API returns; each has an
// entry in internal/i18n/en.json.
const (
	CodeAuthInvalidCredentials = "AUTH_INVALID_CREDENTIALS"
	CodeValidationError        = "VALIDATION_ERROR"
	CodeUnauthenticated        = "UNAUTHENTICATED"
	CodeCSRFRejected           = "CSRF_REJECTED"
	CodeMovieNotFound          = "MOVIE_NOT_FOUND"
	CodeNotFound               = "NOT_FOUND"
	CodeMethodNotAllowed       = "METHOD_NOT_ALLOWED"
	CodeRateLimited            = "RATE_LIMITED"
	CodeInternalError          = "INTERNAL_ERROR"
)

type contextKey int

const requestIDKey contextKey = iota

// withRequestID generates the id the error body, the X-Request-Id header and
// the log line share. An incoming X-Request-Id is deliberately ignored (S16).
func withRequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw := make([]byte, 8)
		if _, err := rand.Read(raw); err != nil {
			raw = []byte{0, 0, 0, 0, 0, 0, 0, 0}
		}
		id := hex.EncodeToString(raw)
		w.Header().Set("X-Request-Id", id)
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), requestIDKey, id)))
	})
}

func requestID(r *http.Request) string {
	id, _ := r.Context().Value(requestIDKey).(string)
	return id
}

type errorBody struct {
	Code      string            `json:"code"`
	Message   string            `json:"message"`
	RequestID string            `json:"requestId"`
	Fields    map[string]string `json:"fields,omitempty"`
}

type errorEnvelope struct {
	Error errorBody `json:"error"`
}

// writeError writes the single JSON error envelope used by every failure.
func (s *Server) writeError(w http.ResponseWriter, r *http.Request, status int, code string, fields map[string]string) {
	s.metrics.ObserveError(code)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(errorEnvelope{Error: errorBody{
		Code:      code,
		Message:   i18n.Message(code),
		RequestID: requestID(r),
		Fields:    fields,
	}})
}

// writeInternal logs an unexpected fault and answers with the generic error.
func (s *Server) writeInternal(w http.ResponseWriter, r *http.Request, err error) {
	s.logger.Error("request failed", "request_id", requestID(r), "route", r.Pattern, "error", err)
	s.writeError(w, r, http.StatusInternalServerError, CodeInternalError, nil)
}

// writeJSON writes a successful JSON response.
func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

// loggerOrDefault keeps a nil logger from panicking in tests.
func loggerOrDefault(logger *slog.Logger) *slog.Logger {
	if logger != nil {
		return logger
	}
	return slog.Default()
}
