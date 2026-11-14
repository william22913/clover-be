package httpapi

import (
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"
	"strings"
	"time"

	"github.com/william22913/clover-be/internal/accounts"
	"github.com/william22913/clover-be/internal/i18n"
	"github.com/william22913/clover-be/internal/sessions"
)

// sessionCookieName is the one cookie the API sets.
const sessionCookieName = "clover_session"

// maxLoginBody is the sign-in body ceiling (S10).
const maxLoginBody = 4 << 10

type loginRequest struct {
	Identifier string `json:"identifier"`
	Password   string `json:"password"`
}

type userPayload struct {
	ID       string `json:"id"`
	Username string `json:"username"`
}

type userResponse struct {
	User userPayload `json:"user"`
}

// login validates the credentials, creates a server-side session and sets the
// cookie. Every refused attempt answers the same generic error (S3).
func (s *Server) login(w http.ResponseWriter, r *http.Request) {
	mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || mediaType != "application/json" {
		s.writeError(w, r, http.StatusBadRequest, CodeValidationError, nil)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxLoginBody)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	var req loginRequest
	if err := decoder.Decode(&req); err != nil {
		var tooLarge *http.MaxBytesError
		if errors.As(err, &tooLarge) {
			s.writeError(w, r, http.StatusRequestEntityTooLarge, CodeValidationError, nil)
			return
		}
		s.writeError(w, r, http.StatusBadRequest, CodeValidationError, nil)
		return
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		s.writeError(w, r, http.StatusBadRequest, CodeValidationError, nil)
		return
	}

	identifier := strings.TrimSpace(req.Identifier)
	if identifier == "" {
		s.writeError(w, r, http.StatusUnprocessableEntity, CodeValidationError, map[string]string{
			"identifier": i18n.Message("FIELD_IDENTIFIER_REQUIRED"),
		})
		return
	}
	if req.Password == "" {
		s.writeError(w, r, http.StatusUnprocessableEntity, CodeValidationError, map[string]string{
			"password": i18n.Message("FIELD_PASSWORD_REQUIRED"),
		})
		return
	}

	refuse := func() {
		s.metrics.ObserveLogin("failed")
		s.writeError(w, r, http.StatusUnauthorized, CodeAuthInvalidCredentials, nil)
	}

	ident, err := accounts.ParseIdentifier(identifier)
	if err != nil {
		refuse()
		return
	}
	if err := accounts.ValidatePassword(req.Password); err != nil {
		refuse()
		return
	}

	account, err := accounts.ByIdentifier(r.Context(), s.pool, ident)
	if errors.Is(err, accounts.ErrNotFound) {
		accounts.VerifyDummy(req.Password)
		refuse()
		return
	}
	if err != nil {
		s.writeInternal(w, r, err)
		return
	}

	ok, err := accounts.VerifyPassword(req.Password, account.PasswordHash)
	if err != nil {
		s.writeInternal(w, r, err)
		return
	}
	if !ok {
		refuse()
		return
	}

	token, err := sessions.Create(r.Context(), s.pool, account.ID)
	if err != nil {
		s.writeInternal(w, r, err)
		return
	}
	if err := sessions.DeleteExpired(r.Context(), s.pool); err != nil {
		s.logger.Warn("expired session cleanup failed", "request_id", requestID(r), "error", err)
	}

	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    token,
		Path:     "/",
		MaxAge:   int(sessions.TTL / time.Second),
		HttpOnly: true,
		Secure:   s.cfg.CookieSecure,
		SameSite: http.SameSiteLaxMode,
	})
	s.metrics.ObserveLogin("ok")
	writeJSON(w, http.StatusOK, userResponse{User: userPayload{
		ID:       account.ID.String(),
		Username: account.Username,
	}})
}

// logout deletes the presented session and clears the cookie. It is
// idempotent: a request with no session still answers 204, so a repeated
// sign-out never fails.
func (s *Server) logout(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie(sessionCookieName); err == nil && cookie.Value != "" {
		if err := sessions.Revoke(r.Context(), s.pool, cookie.Value); err != nil {
			s.writeInternal(w, r, err)
			return
		}
	}
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   s.cfg.CookieSecure,
		SameSite: http.SameSiteLaxMode,
	})
	w.WriteHeader(http.StatusNoContent)
}

// me returns the signed-in account's own data. The account comes from the
// session, never from a request parameter (scope: own).
func (s *Server) me(w http.ResponseWriter, r *http.Request) {
	if !s.authorize(w, r, PermAccountView) {
		return
	}
	id := identityFrom(r.Context())
	writeJSON(w, http.StatusOK, userResponse{User: userPayload{
		ID:       id.Account,
		Username: id.Username,
	}})
}
