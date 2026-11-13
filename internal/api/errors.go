package api

import (
	"encoding/json"
	"net/http"

	"github.com/william22913/clover-be/internal/metrics"
)

// Stable error codes from the contract.
const (
	codeInvalidRequest     = "invalid_request"
	codeInvalidCredentials = "invalid_credentials"
	codeSessionExpired     = "session_expired"
	codeMovieNotFound      = "movie_not_found"
	codeVideoUnavailable   = "video_unavailable"
	codeNotFound           = "not_found"
	codeMethodNotAllowed   = "method_not_allowed"
	codeEmailTaken         = "email_taken"
	codeValidationFailed   = "validation_failed"
	codeRateLimited        = "rate_limited"
	codeInternalError      = "internal_error"
)

type errorEnvelope struct {
	Code      string            `json:"code"`
	Message   string            `json:"message"`
	RequestID string            `json:"request_id"`
	Fields    map[string]string `json:"fields,omitempty"`
}

type errorBody struct {
	Error errorEnvelope `json:"error"`
}

// fail writes the one error envelope and counts the code.
func (h *Handler) fail(w http.ResponseWriter, r *http.Request, status int, code, message string, fields map[string]string) {
	metrics.Error(code)
	writeJSON(w, status, errorBody{Error: errorEnvelope{
		Code:      code,
		Message:   message,
		RequestID: requestIDFrom(r.Context()),
		Fields:    fields,
	}})
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
