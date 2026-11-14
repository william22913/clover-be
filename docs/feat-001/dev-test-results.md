# feat-001 — Clover v1 movie streaming: implementation and test results

Spec version: 1
Testcase version: 1
Part PRs: PART-1=pending
Part PRs: PART-2=pending
Part PRs: PART-3=pending
Part PRs: PART-4=pending

Repository: `clover-be` (Go API). Branch: `feat/feat-001-dev-movie-streaming-v1`,
created from the `feat/feat-001` handoff branch. The browser app is the separate
`clover-fe` repository; nothing here serves HTML.

## What was built

| Part | Content | Where |
|---|---|---|
| PART-1 | Schema migrations, runtime config, the 20-film catalogue + 7 fixture accounts, committed posters and sample videos | `migrations/`, `internal/{config,db,seed}`, `cmd/{server,seed}`, `media/` |
| PART-2 | Argon2id sign-in, opaque session cookie, sign-out, signed-in account read | `internal/{accounts,sessions}`, `internal/httpapi/{auth,middleware}.go` |
| PART-3 | Public catalogue, film detail, poster stream, session-gated Range video | `internal/{movies,media}`, `internal/httpapi/movies.go` |
| PART-4 | Prometheus registry on `:9090`, one structured log line per request, probes | `internal/observability`, `internal/httpapi/{router,errors,middleware}.go` |

## Verification runs

| Run | Command / tool | Result |
|---|---|---|
| Unit tests (local, no database) | `go test ./...` | PASS — every package `ok`; the database-backed cases skip with a clear reason when `TEST_DATABASE_URL` is unset |
| Static checks | `gofmt -l .`, `go vet ./...` | PASS — no output |
| Layout detector | `detect_go_project_layout.py .` | General-Go project: no `nexcommon` dependency, therefore the stdlib layout in `architecture.md` applies (`cmd/`, `internal/`, `migrations/`) |
| Layout verifier | `verify_generated_go_layout.py . <all .go files>` | `OK - no misplaced endpoint files, no leaked template artifacts` (28 files) |
| Full Go suite incl. database cases | `verify_stack` runner, `tests` compose service (`go test ./...` against PostgreSQL 16 in the stack) | PASS — `tests exited (0)` |
| Live API + transport | `verify_stack` runner with the Postman collection | PASS — stack healthy; **31 requests, 79 assertions, 0 failed**, including one `OPTIONS` per browser-called route and a `/metrics` scrape |
| SQL security | `nexcraft:sql-security-validator` on every SQL site | PASS — no blockers; all values bound with `$N`; the only raw execution is the embedded migration body from `//go:embed` |
| Container build | `docker build` inside the verification runner | PASS — both binaries built, image ran as non-root |

`verify_bundle_coverage.py` does not apply: it derives nexcommon's error codes,
and this service has no nexcommon dependency. Its intent is covered by
`TestI18nCatalogueCoversEveryErrorCode`, which asserts an English entry for
every code the API can return.

## Decisions recorded (no question was raised; these were mine to make)

1. **400 and 413 for malformed bodies.** `architecture.md` S10 requires
   `MaxBytesReader` (4 KiB), rejected unknown fields and a strict content type;
   its error table lists only 422 for validation. Implementation: malformed
   JSON, unknown field, duplicate value or wrong content type → **400**; a body
   over 4 KiB → **413**; both use the documented `VALIDATION_ERROR` code and
   message, so the frontend's handling is unchanged. Empty fields stay **422**
   with `fields`.
2. **Sign-out is not permission-gated.** `POST /v1/auth/logout` always answers
   **204**, with the same-origin check and an explicit cookie clear, so a
   repeated or stale sign-out never errors (idempotent by design).
3. **CSRF with no browser headers.** A POST carrying neither `Origin` nor
   `Sec-Fetch-Site` is allowed (S8 names those headers); one carrying a
   mismatched value is refused with **403 `CSRF_REJECTED`**. This keeps
   non-browser clients working without weakening the browser case.
4. **Unauthorized is 401, never 403.** The contract has no `FORBIDDEN` code;
   `403` is reserved for `CSRF_REJECTED`. Every permission failure, including a
   visitor hitting the video route, answers `401 UNAUTHENTICATED` before any
   media byte is written.
5. **Media is committed, not generated at runtime.** Posters are the approved
   mock assets; the videos are small faststart MP4s (1 fps, 320×180, silent AAC)
   committed under `media/`. Durations match the seed rows exactly: 960 s for
   `long-evening` (the >15-minute film TC-052 needs), 600 s for `amber-drift` and
   `orbit` (the ~10-minute film TC-055 names), 90 s for the other 17. The seed
   verifies every file exists and is non-empty and fails loudly otherwise.
6. **One image, two binaries.** The image contains `/app/server` and `/app/seed`;
   compose runs the seed once and gates the API on its successful exit.
7. **`OPTIONS` on an API route answers 405.** The SPA is same-origin through its
   own proxy, so no CORS preflight is needed; every browser-called route still
   has a defined, JSON-enveloped answer to `OPTIONS`, which the collection
   asserts.
8. **CI definition lives in `ci/`.** The credential available to this run has no
   GitHub `workflow` scope, so pushing `.github/workflows/ci.yml` is rejected.
   The pipeline is delivered verbatim as `ci/github-actions-ci.yml` with the
   one-line install step in the README.

## Security design compliance

| # | Control as designed | Implementation | Proof |
|---|---|---|---|
| S1 | Video needs `playback:watch`; 401 before bytes; `private, no-store`, `Vary: Cookie` | `getVideo` checks the session first; `media.ServeVideo` sets the headers | `TestVideo_TC061_NoSessionIsRefused`, `TestVideo_TC052_TC055_TC056_RangeStreaming`, `TestMedia_ServeVideo_TC052_TC055_RangeAndHead` |
| S2 | Sign-out deletes the row; every request re-hashes and re-checks | `sessions.Revoke`, `sessions.Resolve` | `TestVideo_TC020_ReplayedCookieAfterLogout`, `TestMe_TC017_TC021_SessionLifecycle` |
| S3 | One generic 401; dummy verify for unknown identifiers | `login` `refuse()` path + `accounts.VerifyDummy` | `TestAuth_TC023_TC024_OneGenericRefusal`, identifier matrix |
| S4 | Argon2id m=19456, t=2, p=1, 16-byte salt, PHC | `internal/accounts/password.go` | `TestPassword_Argon2idPHC` |
| S5 | No truncation; rune-count 8–128 | `ValidatePassword` | `TestValidatePassword_TC039_TC042_Length` |
| S6 | Only the token's SHA-256 is stored | `sessions.HashToken` / `token_hash bytea` | `TestSessions_TokenIsOpaqueAndHashed`, schema CHECK on 32 bytes |
| S7 | HttpOnly, SameSite=Lax, Path=/, Secure per config, 30-day Max-Age | `login` cookie | Postman `session cookie is HttpOnly` assertion |
| S8 | Same-origin check on every POST | `sameOrigin` | `TestSameOrigin_S8`, `TestAuth_CSRF_CrossOriginIsRefused` |
| S9 | 30/min per-IP bucket, no account lockout | `rateLimiter`, `limitLogin` | `TestRateLimiter_TC047_NoAccountLockout`, `TestAuth_TC047_NoAccountLockout`, `TestAuth_RateLimit` |
| S10 | 4 KiB limit, unknown fields rejected, strict content type | `login` decoder | `TestAuth_MalformedRequests` (8 KiB → 413, unknown field → 400) |
| S11 | Slug rule + row-sourced paths contained under the media root | `media.ValidSlug`, `media.Resolve` | `TestResolve_S11_RejectsPathsOutsideTheRoot`, `TestValidSlug_S11`, TC-058 |
| S12 | FE-only (`next` validation) | — | Not in this repository |
| S13 | FE serves the CSP; API sets nosniff/frame/referrer headers | `securityHeaders` | `TestUnknownRouteAndWrongVerb_AreJSON` (headers asserted via the router) |
| S14 | Secrets only from the environment; `.env.example` placeholders | `internal/config`, `.env.example`, `.gitignore` | Repository scan |
| S15 | Parameterized SQL only | all queries | SQL validator PASS |
| S16 | Server-generated request id in body + header + log line | `withRequestID`, `writeError`, `logRequests` | `TestRequestID_IsGeneratedServerSide`, `TestErrorEnvelope_Shape` |

`X-Content-Type-Options: nosniff`, `X-Frame-Options: DENY` and
`Referrer-Policy: strict-origin-when-cross-origin` are set on every API and
media response; the CSP itself belongs to the frontend's nginx (S13).

## Error contract compliance

| HTTP | Code | Implemented | Verified by |
|---|---|---|---|
| 401 | `AUTH_INVALID_CREDENTIALS` | unknown identifier, wrong password, malformed identifier, out-of-policy password | `TestAuth_TC023_TC024_OneGenericRefusal`, both matrices |
| 422 | `VALIDATION_ERROR` + `fields` | empty identifier / empty password | `TestAuth_TC025_TC026_FieldMessages` |
| 401 | `UNAUTHENTICATED` | no/expired/revoked session on `/v1/auth/me` and the video route | TC-016, TC-020, TC-021 tests |
| 403 | `CSRF_REJECTED` | mismatched `Origin` / `Sec-Fetch-Site` on POST | `TestSameOrigin_S8`, `TestAuth_CSRF_CrossOriginIsRefused` |
| 404 | `MOVIE_NOT_FOUND` | unknown or malformed slug on film, poster, video | `TestMovie_TC058_UnknownFilmIsCleanNotFound` |
| 404 | `NOT_FOUND` | unknown route | `TestUnknownRouteAndWrongVerb_AreJSON` |
| 405 | `METHOD_NOT_ALLOWED` | known route, wrong verb (including OPTIONS) | `TestUnknownRouteAndWrongVerb_AreJSON`, Preflight folder |
| 429 | `RATE_LIMITED` + `Retry-After` | sign-in bucket exhausted | `TestAuth_RateLimit` |
| 500 | `INTERNAL_ERROR` | unexpected fault; details only in the log | `recoverPanic`, `writeInternal` |
| 400 / 413 | `VALIDATION_ERROR` | malformed/oversized sign-in body (decision 1) | `TestAuth_MalformedRequests` |

Every error body carries `requestId`, matching the `X-Request-Id` header and the
log line; no error message contains SQL, a path or a stack trace.

## Test case results

Legend: **PASS** = executed and observed; **PARTIAL** = the server half of the
case is executed and passed, the browser half belongs to `clover-fe`;
**BLOCKED** = the case can only be executed in a browser against the frontend.

| Case ID | Result | Evidence | Notes |
|---|---|---|---|
| TC-001 | BLOCKED | — | Catalogue rendering is the frontend; data verified by TC-003 |
| TC-002 | BLOCKED | — | Page layout/count copy is the frontend |
| TC-003 | PASS | `TestCatalogue_TC003_AnonymousReadsTwentyFilms`; Newman "List the catalogue" | 20 films, all fields, editorial order, public |
| TC-004 | BLOCKED | — | Empty-state copy and rendering are the frontend; the API shape is `{"movies":[]}` |
| TC-005 | BLOCKED | — | Poster-failure placeholder is the frontend |
| TC-006 | PARTIAL | `TestI18n_TC006_EveryMessageIsEnglish`; catalogue and error copy asserted English | All API copy is English with no locale switch; UI copy lives in `clover-fe` |
| TC-007 | PASS | `TestCatalogue_TC007_AcceptLanguageIsIgnored`; Newman "English whatever Accept-Language says" | `fr-FR`, absent, `xx-XX` → identical English bodies |
| TC-008 | BLOCKED | — | Sign-in gate and target chip are UI |
| TC-009 | PARTIAL | Newman "Sign in with the username"; `TestAuth_IdentifierMatrix` | Username sign-in succeeds; landing on the player is UI |
| TC-010 | PARTIAL | `TestAuth_IdentifierMatrix` (email row); Newman uppercase email | Email sign-in succeeds; landing is UI |
| TC-011 | PARTIAL | Newman "The account comes from the session" | Signed-in state is maintained; direct-open player is UI |
| TC-012 | BLOCKED | — | Cancel control is UI |
| TC-013 | BLOCKED | — | Header sign-in flow is UI |
| TC-014 | PARTIAL | `TestSeed_TC014_NoRegistrationRoute` | No registration route exists (404/405); UI inspection is blocked |
| TC-015 | PARTIAL | Newman "Signed out cannot fetch video" | Video refuses without a session; the player-address flow is UI |
| TC-016 | PARTIAL | `TestAuth_TC016_SessionExpiryRevokesAccess` | Server-expired session → 401 on `/me` and video; browser redirect is UI |
| TC-017 | PARTIAL | `TestMe_TC017_TC021_SessionLifecycle`; Newman | The cookie alone keeps the session across requests; page reload is UI |
| TC-018 | BLOCKED | — | Needs a persistent browser profile (browser restart) |
| TC-019 | PARTIAL | `TestMe_TC017_TC021_SessionLifecycle`; Newman sign-out then 401s | Sign-out revokes; catalogue/confirmation copy is UI |
| TC-020 | PASS | `TestVideo_TC020_ReplayedCookieAfterLogout` | Captured cookie replayed after sign-out: 401 for GET and Range, no bytes |
| TC-021 | PASS | `TestMe_TC017_TC021_SessionLifecycle` | Revoked session refused; a fresh session works |
| TC-022 | PASS | `TestAuth_TC022_TwoDevices` | Two live sessions for one account; signing out of one leaves the other |
| TC-023 | PASS | Newman "Wrong password is one generic refusal"; `TestAuth_TC023_TC024_OneGenericRefusal` | Same generic refusal, then the correct password succeeds |
| TC-024 | PASS | `TestAuth_TC023_TC024_OneGenericRefusal`; Newman | Unknown username and unknown email give the identical body |
| TC-025 | PARTIAL | `TestAuth_TC025_TC026_FieldMessages`; Newman "Empty identifier names the field" | 422 with `identifier` message; the pre-request UI block is frontend |
| TC-026 | PARTIAL | `TestAuth_TC025_TC026_FieldMessages` | 422 with `password` message; UI block is frontend |
| TC-027 | PARTIAL | `TestAuth_TC025_TC026_FieldMessages` (spaces row) | Whitespace-only identifier → 422; no browser request is sent by the UI |
| TC-028 | PASS | `TestAuth_IdentifierMatrix` | `"  alex  "` signs in |
| TC-029 | PASS | `TestAuth_IdentifierMatrix` | 3-character username accepted |
| TC-030 | PASS | `TestAuth_IdentifierMatrix` | 2-character username refused generically |
| TC-031 | PASS | `TestAuth_IdentifierMatrix` | 50-character username accepted |
| TC-032 | PASS | `TestAuth_IdentifierMatrix` | 51-character username refused |
| TC-033 | PASS | `TestAuth_IdentifierMatrix` | `mira_k9` accepted |
| TC-034 | PASS | `TestAuth_IdentifierMatrix` | hyphen, space, dot, at-sign all refused generically |
| TC-035 | PASS | `TestAuth_IdentifierMatrix` | `mira.k@example.com` accepted |
| TC-036 | PASS | `TestAuth_IdentifierMatrix` | all malformed emails refused with the same message |
| TC-037 | PASS | `TestAuth_IdentifierMatrix` | `ALEX` accepted |
| TC-038 | PASS | `TestAuth_IdentifierMatrix`; Newman uppercase email | `ALEX.RIVERA@EXAMPLE.COM` accepted |
| TC-039 | PASS | `TestAuth_PasswordMatrix` | 8-character password accepted |
| TC-040 | PASS | `TestAuth_PasswordMatrix` | 7-character password refused |
| TC-041 | PASS | `TestAuth_PasswordMatrix` | 128-character password accepted |
| TC-042 | PASS | `TestAuth_PasswordMatrix` | 129-character password refused |
| TC-043 | PASS | `TestAuth_PasswordMatrix` | letters without a digit refused |
| TC-044 | PASS | `TestAuth_PasswordMatrix` | digits without a letter refused |
| TC-045 | PASS | `TestAuth_PasswordMatrix`, `TestPassword_TC045_CaseSensitive` | passwords are case-sensitive |
| TC-046 | PASS | `TestAuth_PasswordMatrix`, `TestPassword_TC046_TrailingSpaceIsSignificant` | trailing space is significant both ways |
| TC-047 | PASS | `TestAuth_TC047_NoAccountLockout`, `TestRateLimiter_TC047_NoAccountLockout` | 10 failures, then the correct password at once; no lock |
| TC-048 | PASS | `TestSeed_TC048_IdentifiersAreUnique` | 7 accounts, unique usernames and emails case-insensitively |
| TC-049 | BLOCKED | — | Player layout and title placement are UI |
| TC-050 | BLOCKED | — | Play control is UI |
| TC-051 | BLOCKED | — | Pause control is UI |
| TC-052 | PARTIAL | `TestVideo_TC052_TC055_TC056_RangeStreaming`; Newman "Seek deep into the 16-minute film" | Server honours a byte-range seek (206, exact slice) on the 16-minute film; dragging the timeline is UI |
| TC-053 | BLOCKED | — | Volume/mute controls are UI |
| TC-054 | BLOCKED | — | Fullscreen is UI |
| TC-055 | PARTIAL | `TestMedia_ServeVideo_TC052_TC055_RangeAndHead`; Newman Range request | `Accept-Ranges: bytes` and partial responses let playback start before download; the throttled-browser observation is UI |
| TC-056 | PARTIAL | `TestVideo_TC052_TC055_TC056_RangeStreaming` | All 20 films stream their own bytes with the right content type; playing each is UI |
| TC-057 | PARTIAL | `TestMovie_TC058_UnknownFilmIsCleanNotFound` | 404 `MOVIE_NOT_FOUND` with a clean body; the "Film not found" page is UI |
| TC-058 | PASS | `TestMovie_TC058_UnknownFilmIsCleanNotFound`; Newman | 404, machine-readable, no stack or database detail |
| TC-059 | BLOCKED | — | Player error state and retry are UI |
| TC-060 | BLOCKED | — | Retry behaviour is UI |
| TC-061 | PASS | `TestVideo_TC061_NoSessionIsRefused`; Newman (plain and ranged) | 401 and no media bytes; catalogue stays public |
| TC-062 | PASS | `TestCatalogue_TC062_TC063_SeedData`, `TestVideo_TC052_TC055_TC056_RangeStreaming` | All 20 films have a title, description, working poster and working video |
| TC-063 | PASS | `TestCatalogue_TC062_TC063_SeedData`, `TestSeed_TC063_DurationCeiling` | Longest seeded film is 960 s; 1800 s is accepted by the schema, 1801 s is rejected at both seed validation and the database CHECK |

Totals: 63 cases — 32 PASS, 16 PARTIAL, 15 BLOCKED. Every API-level acceptance
case in the test design is executed; the PARTIAL cases are the end-to-end cases
whose browser half runs in `clover-fe`, and the BLOCKED ones are pure player/UI
behaviour that has no server side.

## Known gaps and follow-ups

1. **Frontend cases.** The 15 BLOCKED cases and the UI half of the 16 PARTIAL
   cases are executed in the `clover-fe` repository against this API; nothing in
   this repository can render them.
2. **CI workflow placement.** `ci/github-actions-ci.yml` must be copied to
   `.github/workflows/ci.yml` (README gives the line) because this run's
   credential lacks the GitHub `workflow` scope. Until then CI does not run on
   push; the same suite was executed by the verification runner instead.
3. **Database roles (deployment note from the SQL review).** The compose stack
   uses one role that is both migrator and runtime user, which is fine for a
   local/demo stack. Before any shared deployment, split it into a migrator role
   owning the schema and a runtime role with only DML on the three tables, and
   use `sslmode=verify-full` for the connection string.
4. **Stale predecessor on the remote.** The remote still carries merged PRs #1
   and #2 into a deleted `feat/FEAT-001` branch (bcrypt, a static `web/` UI).
   They are not ancestors of this branch and were not reused; they are history
   noise on the remote, not part of this delivery.
