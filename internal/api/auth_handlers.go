package api

import (
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"
	"strings"
	"time"

	"github.com/william22913/clover-be/internal/auth"
	"github.com/william22913/clover-be/internal/metrics"
	"github.com/william22913/clover-be/internal/store"
)

const (
	sessionCookieName = "clover_session"
	sessionLifetime   = 7 * 24 * time.Hour
	maxBodyBytes      = 1 << 20
)

func (h *Handler) signUp(w http.ResponseWriter, r *http.Request) {
	if !h.limiter.Allow(clientIP(r)) {
		h.fail(w, r, http.StatusTooManyRequests, codeRateLimited,
			"Too many attempts. Try again in a minute.", nil)
		return
	}

	var req credentialsRequest
	if status, ok := decodeJSON(w, r, &req); !ok {
		h.fail(w, r, status, codeInvalidRequest, "Something went wrong. Please try again.", nil)
		return
	}
	// REQ-022: surrounding whitespace is trimmed once at this boundary so two
	// spellings cannot become two accounts. The password is never trimmed.
	req.Email = strings.TrimSpace(req.Email)

	fields := map[string]string{}
	if msg := auth.ValidateEmail(req.Email); msg != "" {
		fields["email"] = msg
	}
	if msg := auth.ValidatePassword(req.Password); msg != "" {
		fields["password"] = msg
	}
	if len(fields) > 0 {
		metrics.SignUp("invalid")
		h.fail(w, r, http.StatusUnprocessableEntity, codeValidationFailed,
			"Check the highlighted fields.", fields)
		return
	}

	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		h.internal(w, r, err)
		return
	}
	token, tokenHash, err := auth.NewToken()
	if err != nil {
		h.internal(w, r, err)
		return
	}
	if _, err := h.store.CreateUserAndSession(r.Context(), req.Email, hash, tokenHash); err != nil {
		if errors.Is(err, store.ErrEmailTaken) {
			metrics.SignUp("duplicate")
			h.fail(w, r, http.StatusConflict, codeEmailTaken,
				"An account already uses this email. Sign in instead.", nil)
			return
		}
		h.internal(w, r, err)
		return
	}

	setSessionCookie(w, token)
	metrics.SignUp("ok")
	writeJSON(w, http.StatusCreated, userBody{User: userDTO{Email: req.Email}})
}

func (h *Handler) signIn(w http.ResponseWriter, r *http.Request) {
	if !h.limiter.Allow(clientIP(r)) {
		metrics.SignIn("rate_limited")
		h.fail(w, r, http.StatusTooManyRequests, codeRateLimited,
			"Too many attempts. Try again in a minute.", nil)
		return
	}

	var req credentialsRequest
	if status, ok := decodeJSON(w, r, &req); !ok {
		h.fail(w, r, status, codeInvalidRequest, "Something went wrong. Please try again.", nil)
		return
	}
	req.Email = strings.TrimSpace(req.Email)

	fields := map[string]string{}
	if msg := auth.ValidateEmail(req.Email); msg != "" {
		fields["email"] = msg
	}
	if msg := auth.ValidateSignInPassword(req.Password); msg != "" {
		fields["password"] = msg
	}
	if len(fields) > 0 {
		h.fail(w, r, http.StatusUnprocessableEntity, codeValidationFailed,
			"Check the highlighted fields.", fields)
		return
	}

	user, err := h.store.UserByEmail(r.Context(), req.Email)
	if errors.Is(err, store.ErrNotFound) {
		auth.CheckDummyPassword(req.Password)
		metrics.SignIn("failed")
		h.invalidCredentials(w, r)
		return
	}
	if err != nil {
		h.internal(w, r, err)
		return
	}
	if !auth.CheckPassword(user.PasswordHash, req.Password) {
		metrics.SignIn("failed")
		h.invalidCredentials(w, r)
		return
	}

	token, tokenHash, err := auth.NewToken()
	if err != nil {
		h.internal(w, r, err)
		return
	}
	var revokeTokenHash string
	if cookie, err := r.Cookie(sessionCookieName); err == nil && cookie.Value != "" {
		revokeTokenHash = auth.HashToken(cookie.Value)
	}
	if err := h.store.CreateSession(r.Context(), user.ID, tokenHash, revokeTokenHash); err != nil {
		h.internal(w, r, err)
		return
	}

	setSessionCookie(w, token)
	metrics.SignIn("ok")
	writeJSON(w, http.StatusOK, userBody{User: userDTO{Email: user.Email}})
}

func (h *Handler) signOut(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie(sessionCookieName); err == nil && cookie.Value != "" {
		revoked, err := h.store.RevokeSession(r.Context(), auth.HashToken(cookie.Value))
		if err != nil {
			h.internal(w, r, err)
			return
		}
		if revoked > 0 {
			metrics.SessionRevoked()
		}
	}
	clearSessionCookie(w)
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) me(w http.ResponseWriter, r *http.Request) {
	user, ok := userFrom(r.Context())
	if !ok {
		h.fail(w, r, http.StatusUnauthorized, codeSessionExpired,
			"Your session ended. Sign in again to keep watching.", nil)
		return
	}
	writeJSON(w, http.StatusOK, userBody{User: userDTO{Email: user.Email}})
}

// invalidCredentials is the one byte-identical 401 for both unknown email and
// wrong password.
func (h *Handler) invalidCredentials(w http.ResponseWriter, r *http.Request) {
	h.fail(w, r, http.StatusUnauthorized, codeInvalidCredentials,
		"That email and password do not match an account.", nil)
}

func setSessionCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    token,
		Path:     "/",
		MaxAge:   int(sessionLifetime.Seconds()),
		Expires:  time.Now().Add(sessionLifetime),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})
}

func clearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})
}

// decodeJSON enforces the 1 MiB limit, the JSON content type and no unknown
// fields. It returns the response status to use when decoding fails.
func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) (int, bool) {
	mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || mediaType != "application/json" {
		return http.StatusBadRequest, false
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		var tooLarge *http.MaxBytesError
		if errors.As(err, &tooLarge) {
			return http.StatusRequestEntityTooLarge, false
		}
		return http.StatusBadRequest, false
	}
	if err := dec.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		var tooLarge *http.MaxBytesError
		if errors.As(err, &tooLarge) {
			return http.StatusRequestEntityTooLarge, false
		}
		return http.StatusBadRequest, false
	}
	return http.StatusOK, true
}
