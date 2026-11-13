package api_test

import (
	"net/http"
	"strings"
	"testing"
)

func signUp(t *testing.T, env *testEnv, email, password string) *http.Response {
	t.Helper()
	return env.do(http.MethodPost, "/v1/auth/sign-up",
		jsonBody(map[string]any{"email": email, "password": password}), nil)
}

func signIn(t *testing.T, env *testEnv, email, password string) *http.Response {
	t.Helper()
	return env.do(http.MethodPost, "/v1/auth/sign-in",
		jsonBody(map[string]any{"email": email, "password": password}), nil)
}

func signUpCookie(t *testing.T, env *testEnv, email, password string) *http.Cookie {
	t.Helper()
	res := signUp(t, env, email, password)
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("sign-up status = %d, want 201", res.StatusCode)
	}
	cookie := sessionCookie(res)
	if cookie == nil {
		t.Fatal("sign-up must set the session cookie")
	}
	return cookie
}

func signInCookie(t *testing.T, env *testEnv, email, password string) *http.Cookie {
	t.Helper()
	res := signIn(t, env, email, password)
	if res.StatusCode != http.StatusOK {
		t.Fatalf("sign-in status = %d, want 200", res.StatusCode)
	}
	cookie := sessionCookie(res)
	if cookie == nil {
		t.Fatal("sign-in must set the session cookie")
	}
	return cookie
}

func fieldErrors(t *testing.T, res *http.Response) map[string]string {
	t.Helper()
	if res.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", res.StatusCode)
	}
	var body errorResponse
	decodeBody(t, res, &body)
	if body.Error.Code != "validation_failed" {
		t.Fatalf("code = %q, want validation_failed", body.Error.Code)
	}
	if body.Error.Message == "" {
		t.Fatal("validation error needs a message")
	}
	return body.Error.Fields
}

// TC-007 — a new visitor creates an account and is signed in immediately.
func TestAuth_TC007_SignUpCreatesAccountAndSession(t *testing.T) {
	env := newEnv(t)
	res := signUp(t, env, "newvisitor@example.com", "Clover123")
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, want 201", res.StatusCode)
	}
	var body struct {
		User struct {
			Email string `json:"email"`
		} `json:"user"`
	}
	decodeBody(t, res, &body)
	if body.User.Email != "newvisitor@example.com" {
		t.Fatalf("user = %+v", body.User)
	}
	cookie := sessionCookie(res)
	if cookie == nil {
		t.Fatal("sign-up must sign the visitor in")
	}
	me := env.do(http.MethodGet, "/v1/auth/me", "", map[string]string{"Cookie": cookieHeader(cookie)})
	if me.StatusCode != http.StatusOK {
		t.Fatalf("me status = %d, want 200", me.StatusCode)
	}
}

// TC-008 — no verification step: the fresh account signs in straight away.
func TestAuth_TC008_NewAccountWorksImmediately(t *testing.T) {
	env := newEnv(t)
	signUpCookie(t, env, "fresh@example.com", "Clover123")

	res := signIn(t, env, "fresh@example.com", "Clover123")
	if res.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", res.StatusCode)
	}
}

// TC-009 — a returning user signs in with the registered email and password.
func TestAuth_TC009_ReturningUserSignsIn(t *testing.T) {
	env := newEnv(t)
	env.store.addUser("viewer@example.com", "Passw0rd1")

	res := signIn(t, env, "viewer@example.com", "Passw0rd1")
	if res.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", res.StatusCode)
	}
	if sessionCookie(res) == nil {
		t.Fatal("sign-in must set the session cookie")
	}
}

// TC-024 — a correct email with a wrong password gets the generic message and
// no session.
func TestAuth_TC024_WrongPasswordGenericError(t *testing.T) {
	env := newEnv(t)
	env.store.addUser("viewer@example.com", "Passw0rd1")

	res := signIn(t, env, "viewer@example.com", "NotThePassword1")
	if res.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", res.StatusCode)
	}
	var body errorResponse
	decodeBody(t, res, &body)
	if body.Error.Code != "invalid_credentials" {
		t.Fatalf("code = %q, want invalid_credentials", body.Error.Code)
	}
	if sessionCookie(res) != nil {
		t.Fatal("a failed sign-in must not issue a session")
	}
}

// TC-025 — an unknown email is indistinguishable from a wrong password: same
// code, same message, same status.
func TestAuth_TC025_UnknownEmailSameGenericError(t *testing.T) {
	env := newEnv(t)
	env.store.addUser("viewer@example.com", "Passw0rd1")

	wrong := signIn(t, env, "viewer@example.com", "NotThePassword1")
	var wrongBody errorResponse
	decodeBody(t, wrong, &wrongBody)

	unknown := signIn(t, env, "nobody@example.com", "NotThePassword1")
	var unknownBody errorResponse
	decodeBody(t, unknown, &unknownBody)

	if wrong.StatusCode != unknown.StatusCode {
		t.Fatalf("statuses differ: %d vs %d", wrong.StatusCode, unknown.StatusCode)
	}
	if wrongBody.Error.Code != unknownBody.Error.Code ||
		wrongBody.Error.Message != unknownBody.Error.Message {
		t.Fatalf("messages differ: %+v vs %+v", wrongBody.Error, unknownBody.Error)
	}
}

// TC-026 — a duplicate email is refused, no second account exists, and the
// visitor is not signed into the existing one.
func TestAuth_TC026_DuplicateEmailRejected(t *testing.T) {
	env := newEnv(t)
	signUpCookie(t, env, "viewer@example.com", "Passw0rd1")

	res := signUp(t, env, "viewer@example.com", "Different123")
	if res.StatusCode != http.StatusConflict {
		t.Fatalf("status = %d, want 409", res.StatusCode)
	}
	var body errorResponse
	decodeBody(t, res, &body)
	if body.Error.Code != "email_taken" {
		t.Fatalf("code = %q, want email_taken", body.Error.Code)
	}
	if sessionCookie(res) != nil {
		t.Fatal("a refused duplicate must not sign anyone in")
	}
	if len(env.store.users) != 1 {
		t.Fatalf("accounts = %d, want exactly 1", len(env.store.users))
	}

	again := signIn(t, env, "viewer@example.com", "Passw0rd1")
	if again.StatusCode != http.StatusOK {
		t.Fatalf("original credentials status = %d, want 200", again.StatusCode)
	}
}

// TC-032 — an empty email is rejected on the email field alone.
func TestAuth_TC032_EmptyEmailRejected(t *testing.T) {
	env := newEnv(t)
	fields := fieldErrors(t, signUp(t, env, "", "Passw0rd1"))
	if fields["email"] == "" {
		t.Fatalf("fields = %+v, want an email message", fields)
	}
	if _, hasPassword := fields["password"]; hasPassword {
		t.Fatalf("fields = %+v, the valid password must not be flagged", fields)
	}
	if len(env.store.users) != 0 {
		t.Fatal("no account may be created")
	}
}

// TC-033 — malformed email addresses are rejected field-level.
func TestAuth_TC033_MalformedEmailRejected(t *testing.T) {
	env := newEnv(t)
	for _, email := range []string{
		"ann.example.com", "ann@", "@example.com", "ann@@example.com", "ann example@example.com",
	} {
		res := signUp(t, env, email, "Passw0rd1")
		if res.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("%q status = %d, want 422", email, res.StatusCode)
		}
		var body errorResponse
		decodeBody(t, res, &body)
		if body.Error.Fields["email"] == "" {
			t.Fatalf("%q has no email message: %+v", email, body.Error.Fields)
		}
	}
	if len(env.store.users) != 0 {
		t.Fatalf("accounts = %d, want 0", len(env.store.users))
	}
}

// TC-034 — an empty password is rejected on the password field.
func TestAuth_TC034_EmptyPasswordRejected(t *testing.T) {
	env := newEnv(t)
	fields := fieldErrors(t, signUp(t, env, "fresh@example.com", ""))
	if fields["password"] == "" {
		t.Fatalf("fields = %+v, want a password message", fields)
	}
	if _, hasEmail := fields["email"]; hasEmail {
		t.Fatalf("fields = %+v, the valid email must not be flagged", fields)
	}
}

// TC-035 — a 7-character password is below the inclusive minimum.
func TestAuth_TC035_SevenCharacterPasswordRejected(t *testing.T) {
	env := newEnv(t)
	fields := fieldErrors(t, signUp(t, env, "fresh@example.com", "Abc1234"))
	if !strings.Contains(fields["password"], "8 characters") {
		t.Fatalf("fields = %+v, want the minimum-length message", fields)
	}
}

// TC-036 — exactly 8 characters with a letter and a number is accepted.
func TestAuth_TC036_EightCharacterPasswordAccepted(t *testing.T) {
	env := newEnv(t)
	res := signUp(t, env, "fresh@example.com", "Abc12345")
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, want 201", res.StatusCode)
	}
}

// TC-037 — a password with no digit is rejected and says why.
func TestAuth_TC037_PasswordWithoutNumberRejected(t *testing.T) {
	env := newEnv(t)
	fields := fieldErrors(t, signUp(t, env, "fresh@example.com", "Abcdefgh"))
	if !strings.Contains(fields["password"], "number") {
		t.Fatalf("fields = %+v, want the missing-number message", fields)
	}
}

// TC-038 — a password with no letter is rejected and says why.
func TestAuth_TC038_PasswordWithoutLetterRejected(t *testing.T) {
	env := newEnv(t)
	fields := fieldErrors(t, signUp(t, env, "fresh@example.com", "12345678"))
	if !strings.Contains(fields["password"], "letter") {
		t.Fatalf("fields = %+v, want the missing-letter message", fields)
	}
}

// TC-039 — a longer password with a symbol is accepted.
func TestAuth_TC039_SymbolPasswordAccepted(t *testing.T) {
	env := newEnv(t)
	res := signUp(t, env, "fresh@example.com", "Clover2026!")
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, want 201", res.StatusCode)
	}
}

// TC-040 — an address differing only by letter case is a duplicate.
func TestAuth_TC040_CaseVariantEmailIsDuplicate(t *testing.T) {
	env := newEnv(t)
	env.store.addUser("ann@example.com", "Passw0rd1")

	res := signUp(t, env, "Ann@Example.com", "Passw0rd1")
	if res.StatusCode != http.StatusConflict {
		t.Fatalf("status = %d, want 409", res.StatusCode)
	}
	var body errorResponse
	decodeBody(t, res, &body)
	if body.Error.Code != "email_taken" {
		t.Fatalf("code = %q, want email_taken", body.Error.Code)
	}
	if len(env.store.users) != 1 {
		t.Fatalf("accounts = %d, want exactly 1", len(env.store.users))
	}
}

// TC-041 — sign-in matches the email case-insensitively.
func TestAuth_TC041_SignInEmailCaseInsensitive(t *testing.T) {
	env := newEnv(t)
	env.store.addUser("ann@example.com", "Passw0rd1")

	res := signIn(t, env, "ANN@EXAMPLE.COM", "Passw0rd1")
	if res.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", res.StatusCode)
	}
}

// TC-042 — the password is compared case-sensitively.
func TestAuth_TC042_SignInPasswordCaseSensitive(t *testing.T) {
	env := newEnv(t)
	env.store.addUser("ann@example.com", "Passw0rd1")

	res := signIn(t, env, "ann@example.com", "passw0rd1")
	if res.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", res.StatusCode)
	}
}

// TC-043 — the password is not trimmed.
func TestAuth_TC043_SignInPasswordNotTrimmed(t *testing.T) {
	env := newEnv(t)
	env.store.addUser("ann@example.com", "Passw0rd1")

	for _, attempt := range []string{" Passw0rd1", "Passw0rd1 "} {
		res := signIn(t, env, "ann@example.com", attempt)
		if res.StatusCode != http.StatusUnauthorized {
			t.Fatalf("%q status = %d, want 401", attempt, res.StatusCode)
		}
	}
}

// TC-044 — a password with an internal space is taken exactly as typed.
func TestAuth_TC044_InternalSpacePasswordExact(t *testing.T) {
	env := newEnv(t)
	cookie := signUpCookie(t, env, "fresh@example.com", "Abc 1234")
	env.do(http.MethodPost, "/v1/auth/sign-out", "", map[string]string{"Cookie": cookieHeader(cookie)})

	if res := signIn(t, env, "fresh@example.com", "Abc 1234"); res.StatusCode != http.StatusOK {
		t.Fatalf("exact password status = %d, want 200", res.StatusCode)
	}
	if res := signIn(t, env, "fresh@example.com", "Abc1234"); res.StatusCode != http.StatusUnauthorized {
		t.Fatalf("space-stripped password status = %d, want 401", res.StatusCode)
	}
}

// TC-045 — empty sign-in fields are rejected inline on both fields.
func TestAuth_TC045_EmptySignInFieldsRejected(t *testing.T) {
	env := newEnv(t)

	both := fieldErrors(t, signIn(t, env, "", ""))
	if both["email"] == "" || both["password"] == "" {
		t.Fatalf("fields = %+v, want both messages", both)
	}

	passwordOnly := fieldErrors(t, signIn(t, env, "viewer@example.com", ""))
	if passwordOnly["password"] == "" {
		t.Fatalf("fields = %+v, want a password message", passwordOnly)
	}
	if _, hasEmail := passwordOnly["email"]; hasEmail {
		t.Fatalf("fields = %+v, a filled email must not be flagged", passwordOnly)
	}
}

// Error contract: a body without the JSON content type is a 400, not a parse
// error page.
func TestAuth_ErrorContractWrongContentType(t *testing.T) {
	env := newEnv(t)
	res := env.do(http.MethodPost, "/v1/auth/sign-in", `{"email":"a@example.com","password":"Passw0rd1"}`,
		map[string]string{"Content-Type": "text/plain"})
	if res.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", res.StatusCode)
	}
	var body errorResponse
	decodeBody(t, res, &body)
	if body.Error.Code != "invalid_request" {
		t.Fatalf("code = %q, want invalid_request", body.Error.Code)
	}
}

// Error contract: a body over 1 MiB is a 413, and unknown fields are refused.
func TestAuth_ErrorContractBodyLimits(t *testing.T) {
	env := newEnv(t)
	huge := `{"email":"a@example.com","password":"` + strings.Repeat("a", 1<<20) + `"}`
	res := env.do(http.MethodPost, "/v1/auth/sign-in", huge, nil)
	if res.StatusCode != http.StatusRequestEntityTooLarge {
		t.Fatalf("oversized status = %d, want 413", res.StatusCode)
	}

	unknown := env.do(http.MethodPost, "/v1/auth/sign-in",
		`{"email":"a@example.com","password":"Passw0rd1","role":"admin"}`, nil)
	if unknown.StatusCode != http.StatusBadRequest {
		t.Fatalf("unknown-field status = %d, want 400", unknown.StatusCode)
	}
}

// Security design — session fixation: a new token is issued on every sign-in,
// and a session presented with that sign-in is revoked in the same transaction.
func TestAuth_SessionRotatesOnSignIn(t *testing.T) {
	env := newEnv(t)
	first := signUpCookie(t, env, "viewer@example.com", "Passw0rd1")

	res := env.do(http.MethodPost, "/v1/auth/sign-in",
		jsonBody(map[string]any{"email": "viewer@example.com", "password": "Passw0rd1"}),
		map[string]string{"Cookie": cookieHeader(first)})
	if res.StatusCode != http.StatusOK {
		t.Fatalf("sign-in status = %d, want 200", res.StatusCode)
	}
	second := sessionCookie(res)
	if second == nil || second.Value == first.Value {
		t.Fatal("sign-in must issue a new token")
	}

	old := env.do(http.MethodGet, "/v1/auth/me", "", map[string]string{"Cookie": cookieHeader(first)})
	if old.StatusCode != http.StatusUnauthorized {
		t.Fatalf("old cookie status = %d, want 401", old.StatusCode)
	}
	fresh := env.do(http.MethodGet, "/v1/auth/me", "", map[string]string{"Cookie": cookieHeader(second)})
	if fresh.StatusCode != http.StatusOK {
		t.Fatalf("new cookie status = %d, want 200", fresh.StatusCode)
	}
}

// REQ-022 / architecture — surrounding whitespace on the email is trimmed once
// at the validator boundary; internal whitespace is still invalid.
func TestAuth_EmailWhitespaceTrimmedAtBoundary(t *testing.T) {
	env := newEnv(t)
	res := signUp(t, env, "  spaced@example.com  ", "Passw0rd1")
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, want 201", res.StatusCode)
	}
	var body struct {
		User struct {
			Email string `json:"email"`
		} `json:"user"`
	}
	decodeBody(t, res, &body)
	if body.User.Email != "spaced@example.com" {
		t.Fatalf("stored email = %q, want the trimmed address", body.User.Email)
	}
	if s := signIn(t, env, " spaced@example.com ", "Passw0rd1"); s.StatusCode != http.StatusOK {
		t.Fatalf("padded sign-in status = %d, want 200", s.StatusCode)
	}
	if fields := fieldErrors(t, signUp(t, env, "two words@example.com", "Passw0rd1")); fields["email"] == "" {
		t.Fatalf("internal whitespace must still be refused: %+v", fields)
	}
}

// Rate limit: the eleventh attempt inside one window is refused before any
// password work happens.
func TestAuth_RateLimitAfterTenAttempts(t *testing.T) {
	env := newEnv(t)
	for i := 0; i < 10; i++ {
		res := signIn(t, env, "", "")
		if res.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("attempt %d status = %d, want 422", i+1, res.StatusCode)
		}
	}
	res := signIn(t, env, "viewer@example.com", "Passw0rd1")
	if res.StatusCode != http.StatusTooManyRequests {
		t.Fatalf("status = %d, want 429", res.StatusCode)
	}
	var body errorResponse
	decodeBody(t, res, &body)
	if body.Error.Code != "rate_limited" {
		t.Fatalf("code = %q, want rate_limited", body.Error.Code)
	}
}
