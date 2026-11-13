---
feature_code: FEAT-001
service: clover
version: 1
status: complete
spec_version: 1
open_questions: []
pr_url: https://github.com/william22913/sandbox-docs/pull/57
last_updated: 2026-11-13
---

# FEAT-001 — Movie Streaming (clover) — test design

Black-box test design derived from the merged `spec.md` v1 (PR #56), with the
merged `PRD.md` (spec_version 1) and `AppFlow.md` (spec_version 1) used as
supporting context only. The spec is authoritative; every expected result below
traces to a `REQ-*` in it. No case in this document has been executed. Cases
describe observable behaviour only and do not prescribe implementation.

There is no `UIMap.md` on the docs base branch for this feature, so screen names
are the spec's own words: **catalogue**, **login / sign-up screen**, **player
page**. Visual design, layout and motion are out of scope here (owned by
`/ui-preparation`); no case asserts colour, spacing or component choice.

## Scope, fixtures and assumptions

### Fixtures

All cases assume the seeded v1 environment required by `spec.md`: the fixed set
of sample movies is present at startup. Adapt the names below to the actual
seed, but keep the properties.

**Movies**

- Every movie in the seed has a poster image, a title, a short description and
  a playable MP4/H.264 video file.
- At least two distinct movies, so a case can prove selection opens *that*
  movie (e.g. "Halcyon" and "Vesper").
- One test-only movie whose video file is unavailable/corrupt, prepared in test
  data for REQ-018 (v1 has no screen where a person edits movie data).
- One movie identifier that does not exist in the seed, for REQ-017.

**Accounts**

| Account | Credentials | Purpose |
|---|---|---|
| Viewer A | `viewer@example.com` / `Passw0rd1` | Primary existing account; signed-in cases |
| Viewer B | `second@example.com` / `Second123` | Second independent account (multi-device, duplicate checks) |
| New visitor | e.g. `newvisitor@example.com` / `Clover123` | Created in sign-up cases; must not exist beforehand |

Both fixture passwords are at least 8 characters with at least one letter and
one number, as REQ-024 requires.

### Assumptions recorded for later reconciliation

- API-level cases assert the behaviour `spec.md` requires. Concrete route
  paths, request/response shapes and status codes are owned by
  `/code-preparation` and are not invented here. "Request the video stream"
  means an HTTP request against whatever route the implementation exposes,
  carrying the session cookie.
- Session lifetime is the 7-day server-side session recorded in the spec's
  hand-off notes. Cases that need expiry or the lifetime boundary prepare it
  with the test environment's clock/session data; the spec does not name a
  mechanism.
- The video-failure condition (REQ-018) is prepared in test data by pointing
  one movie at an unavailable or corrupt video file.
- "Every published movie" in v1 means every movie in the seed: there is no
  publishing or management screen in v1.
- Duplicate-account checks may read the test environment's own data store to
  confirm a single account row. That is verification of test data, not
  inspection of implementation code.
- The spec does not define what signing out on one device does to another
  device's active session. REQ-021 asserts only that signing in on one device
  does not sign the user out on another; no case asserts the sign-out
  direction.
- The spec sets no maximum password length and no ban on symbols, so the valid
  equivalence class includes longer and mixed-character passwords (TC-039).
- Error wording is not fixed by the spec; "clear error" is assessed as a
  user-facing English message that explains what to fix and shows no raw error,
  stack trace or internal detail.

## Requirement coverage

Every active `REQ-001`..`REQ-024` in `spec.md` v1 has at least one citing case.

| REQ | Requirement (short) | Cases | Notes |
|---|---|---|---|
| REQ-001 | Catalogue visible immediately without login, no account prompt in the way | TC-001 | |
| REQ-002 | Every published movie listed; no pagination, search, filtering or sorting | TC-002, TC-003 | |
| REQ-003 | Select any movie and open its player page | TC-004 | Verified via a signed-in viewer, because a signed-out selection is gated by REQ-004. |
| REQ-004 | Signed-out selection goes to the combined login/sign-up screen, then returns to that movie | TC-006, TC-007, TC-010 | |
| REQ-005 | Self-service email/password sign-up; no email verification; usable immediately | TC-007, TC-008 | |
| REQ-006 | Returning user signs in with the registered email and password | TC-009, TC-010 | |
| REQ-007 | Signed-in user stays signed in between visits | TC-011, TC-012 | The 7-day lifetime is the spec's recorded session decision; TC-012 tests its boundary. |
| REQ-008 | Player offers play, pause, timeline scrubber, volume and fullscreen | TC-016, TC-017, TC-018, TC-019 | |
| REQ-009 | Progressive playback; seeking resumes from the new point | TC-020, TC-021, TC-022 | |
| REQ-010 | Sign-out returns the visitor to anonymous browsing; player locked again | TC-014 | |
| REQ-011 | The entire interface is in English | TC-005 | |
| REQ-012 | Direct player URL while signed out behaves like REQ-004 | TC-023 | |
| REQ-013 | Unknown email or wrong password: the same generic error, stay on the form | TC-024, TC-025 | |
| REQ-014 | Duplicate email at sign-up rejected; no second account; not signed into the existing one | TC-026 | |
| REQ-015 | Invalid sign-up/sign-in input rejected inline; typed input is preserved | TC-032, TC-034, TC-035, TC-037, TC-038, TC-045 | |
| REQ-016 | Abandoning the login screen keeps the visitor anonymous; the movie stays locked | TC-027 | |
| REQ-017 | Non-existent movie URL shows a clear "movie not available" message and a way back | TC-028 | |
| REQ-018 | Video load failure shows a clear error with retry; never an endless spinner | TC-029, TC-030 | |
| REQ-019 | Expired session on the player sends the user to login and back to the same movie | TC-031 | |
| REQ-020 | Sign-out is complete server-side; the old cookie cannot fetch video or account data | TC-015 | |
| REQ-021 | Same account on several devices; signing in elsewhere does not sign out | TC-013 | Cross-device sign-out behaviour is unspecified, so no case asserts it. |
| REQ-022 | Email required, syntactically valid, unique case-insensitively | TC-032, TC-033, TC-040, TC-041 | |
| REQ-023 | Sign-in email is case-insensitive; password must match exactly (no trim, no case-fold) | TC-041, TC-042, TC-043, TC-044, TC-045 | |
| REQ-024 | Password required, minimum 8 characters, at least one letter and one number, taken exactly as typed | TC-034, TC-035, TC-036, TC-037, TC-038, TC-039, TC-044 | Applies at account creation only; the sign-in side is covered by REQ-023 cases. The forward-compatibility clause ("cannot lock out an account created under a different rule later") has no v1 counterpart: v1 has one creation rule, so there is nothing to exercise until a later rule exists. |

## Acceptance criteria coverage

| SC | Acceptance criterion (short) | Cases |
|---|---|---|
| SC-001 | Anonymous first page shows every published movie with poster, title and short description; no login needed | TC-001, TC-002 |
| SC-002 | Anonymous visitor selects a movie, creates an account, lands on that movie's player, signed in, video available | TC-007 |
| SC-003 | Anonymous visitor selects a movie, signs in with an existing account, lands on that movie's player | TC-010 |
| SC-004 | Play, pause, timeline scrub, volume and fullscreen take effect; seeking resumes from the new point | TC-016, TC-017, TC-018, TC-019, TC-020 |
| SC-005 | Correct email plus wrong password returns an error and does not sign the visitor in | TC-024 |
| SC-006 | Already-registered email at sign-up returns an error; exactly one account remains | TC-026 |
| SC-007 | A visitor who abandons the login screen remains anonymous and cannot play a movie | TC-027 |
| SC-008 | A user who returns later is still signed in and can play without re-entering credentials | TC-011 |
| SC-009 | Replaying the previously issued session cookie against the video or account endpoints is refused | TC-015 |
| SC-010 | A non-existent movie and an unloadable video each show a clear message with a recovery action | TC-028, TC-029 |
| SC-011 | Sign-up rejects a 7-character password, an 8-character password without a digit, and an 8-character password without a letter, each explaining the broken rule | TC-035, TC-037, TC-038 |
| SC-012 | Signing up with an address differing only by letter case is rejected as a duplicate | TC-040 |

## Non-goals — what is deliberately not tested

The spec's Non-Goals list absences, not behaviours, so most have no case: there
is no user-facing action to attempt and no stated rejection behaviour.

- Recommendations, subscriptions/payments, search, filtering/sorting,
  movie-management screens, email verification, password reset/account
  recovery, profiles, watchlists, viewing history, ratings, comments/sharing,
  non-English UI, and native apps: **no case** — each is out of scope with no
  v1 surface to exercise.
- The two absences the spec does turn into testable statements are covered:
  "no search, filtering, sorting or pagination" is part of REQ-002 and is
  checked by TC-003; "English only" is REQ-011 and is checked by TC-005.
- No case checks the behaviour of arbitrary unknown routes beyond the player
  URL for a non-existent movie (REQ-017), because the spec defines no rule for
  other unknown routes.

## Test cases

### TC-001 — An anonymous visitor sees the catalogue immediately, with no account prompt
- **Type:** positive
- **Spec:** REQ-001, SC-001
- **Level:** UI
- **Priority:** P1
- **Automation candidate:** yes
- **Preconditions / test data:** No active session (fresh browser, no cookies); the seeded catalogue is present.
- **Action / input:** Open the site's root URL as an anonymous visitor.
- **Expected result:** The first page shown is the movie catalogue — no redirect to a login screen and no sign-in prompt or overlay in the way. Each movie card shows a loaded poster image, a title and a short description. Browsing requires no account.

### TC-002 — The catalogue lists every published movie, with no pagination
- **Type:** positive
- **Spec:** REQ-002, SC-001
- **Level:** end-to-end
- **Priority:** P1
- **Automation candidate:** yes
- **Preconditions / test data:** No active session; the seed contains N movies (read N from the seed at test time).
- **Action / input:** Open the catalogue; compare every displayed title with the seed list; scroll to the end of the page.
- **Expected result:** All N seeded movies appear exactly once — no movie missing and no extra movie. There is no page 2, no "load more", and no infinite-scroll cut-off: the whole set is reachable in one unpaginated list without signing in.

### TC-003 — The catalogue offers no search, filter, sort or pagination controls
- **Type:** positive
- **Spec:** REQ-002
- **Level:** UI
- **Priority:** P2
- **Automation candidate:** yes
- **Preconditions / test data:** No active session; the catalogue has several movies.
- **Action / input:** Inspect the catalogue page for a search field, genre/category filters, a sort control, and pagination or "load more" controls.
- **Expected result:** None of these controls exists. The catalogue's only action is selecting a movie to open it.

### TC-004 — Selecting any movie opens that movie's player page
- **Type:** positive
- **Spec:** REQ-003
- **Level:** end-to-end
- **Priority:** P1
- **Automation candidate:** yes
- **Preconditions / test data:** Signed in as Viewer A; the catalogue shows at least two distinct movies.
- **Action / input:** Select movie X from the catalogue; return to the catalogue; select movie Y.
- **Expected result:** Each selection opens the player page for exactly that movie (the page's movie context matches the selected title), not the catalogue and not another movie. The player is ready with a play control.

### TC-005 — The entire interface is in English
- **Type:** positive
- **Spec:** REQ-011
- **Level:** UI
- **Priority:** P2
- **Automation candidate:** yes
- **Preconditions / test data:** Seeded catalogue; Viewer A credentials available.
- **Action / input:** Visit the catalogue, the login / sign-up screen, and the player page as a signed-in viewer; also trigger one error (wrong password) and open a non-existent movie URL; inspect all visible copy.
- **Expected result:** Every label, message, error and control text on these screens is in English. No other language appears, and there is no language or locale switcher.

### TC-006 — A signed-out visitor selecting a movie is taken to the combined login / sign-up screen
- **Type:** positive
- **Spec:** REQ-004
- **Level:** end-to-end
- **Priority:** P1
- **Automation candidate:** yes
- **Preconditions / test data:** No active session; the catalogue is present.
- **Action / input:** Select any movie from the catalogue.
- **Expected result:** The login / sign-up screen appears, offering both signing in and creating an account on the same screen. The selected movie is remembered for the return trip, and the video does not play while the visitor is signed out.

### TC-007 — A new visitor creates an account and lands on the selected movie's player
- **Type:** positive
- **Spec:** REQ-004, REQ-005, SC-002
- **Level:** end-to-end
- **Priority:** P1
- **Automation candidate:** yes
- **Preconditions / test data:** No active session; a fresh email address (e.g. `newvisitor@example.com`) that has no account; password `Clover123`.
- **Action / input:** Select movie X from the catalogue; choose sign-up on the login / sign-up screen; enter the fresh email and the valid password; submit.
- **Expected result:** The account is created and the visitor is signed in. The browser lands on movie X's player page — never the catalogue and never a generic home page — with the video ready to play. No email-verification step intervenes.

### TC-008 — Sign-up requires no email verification and the new account works immediately
- **Type:** positive
- **Spec:** REQ-005
- **Level:** end-to-end
- **Priority:** P2
- **Automation candidate:** yes
- **Preconditions / test data:** An account just created in this run (from TC-007 or equivalent); a separate, fresh browser context with no session.
- **Action / input:** In the fresh browser context, sign in with the newly created credentials, without clicking a verification link or entering a code.
- **Expected result:** Sign-in succeeds immediately. No verification step is required and no unverified state blocks playback: the signed-in viewer can open a movie and play it.

### TC-009 — A returning user signs in with the registered email and password
- **Type:** positive
- **Spec:** REQ-006
- **Level:** end-to-end
- **Priority:** P1
- **Automation candidate:** yes
- **Preconditions / test data:** Viewer A's account was registered earlier; the visitor is signed out.
- **Action / input:** On the login / sign-up screen, enter `viewer@example.com` and the exact password `Passw0rd1`; submit.
- **Expected result:** The visitor is signed in with no error and reaches the signed-in experience; playback is available.

### TC-010 — Signing in from the gated screen returns the visitor to the selected movie
- **Type:** positive
- **Spec:** REQ-004, REQ-006, SC-003
- **Level:** end-to-end
- **Priority:** P1
- **Automation candidate:** yes
- **Preconditions / test data:** No active session; Viewer A's account exists; the catalogue is present.
- **Action / input:** Select movie Y from the catalogue; on the login / sign-up screen sign in with Viewer A's email and password.
- **Expected result:** After sign-in the visitor lands on movie Y's player page with the video ready to play — never the catalogue and never a generic home page — and is signed in.

### TC-011 — A signed-in user stays signed in across a browser restart
- **Type:** positive
- **Spec:** REQ-007, SC-008
- **Level:** end-to-end
- **Priority:** P1
- **Automation candidate:** yes
- **Preconditions / test data:** Viewer A is signed in in the browser; the return happens within the session lifetime.
- **Action / input:** Close the browser completely, reopen it, and open the Clover site.
- **Expected result:** The visitor is still signed in without entering credentials; any movie can be opened and played directly.

### TC-012 — The session is valid just before the 7-day lifetime and refused after it
- **Type:** boundary
- **Spec:** REQ-007
- **Level:** API
- **Priority:** P2
- **Automation candidate:** yes (requires test-clock or session-expiry control)
- **Preconditions / test data:** A session created at time T for Viewer A; the environment can place that session at T + 6 days 23:59 and at T + 7 days 00:01.
- **Action / input:** At each prepared point, request the video stream (and open the player page) using that session cookie.
- **Expected result:** Before the 7-day lifetime the request is authorised and the video plays. After the lifetime the session is no longer accepted, no video data is served, and the visitor is asked to sign in.

### TC-013 — The same account can be signed in on two devices at once
- **Type:** positive
- **Spec:** REQ-021
- **Level:** API
- **Priority:** P2
- **Automation candidate:** yes
- **Preconditions / test data:** Viewer B's account exists; two independent browser contexts A and B with separate cookie stores.
- **Action / input:** Sign in on A; sign in with the same account on B; then use both sessions to request the video stream and open a player.
- **Expected result:** Both sign-ins succeed and remain valid at the same time. Signing in on B does not sign A out, and A can still play. (What signing out on one device does to the other is not specified, so it is not asserted.)

### TC-014 — Sign-out returns the visitor to anonymous browsing with the player locked
- **Type:** positive
- **Spec:** REQ-010
- **Level:** end-to-end
- **Priority:** P1
- **Automation candidate:** yes
- **Preconditions / test data:** Signed in as Viewer A on the player page of a playable movie.
- **Action / input:** Sign out; then look at the catalogue; then select a movie (and, separately, open a player URL directly).
- **Expected result:** Sign-out ends the signed-in state and the catalogue is visible again with no login prompt. Selecting a movie or opening a player URL now shows the login / sign-up screen, and no video plays until the visitor signs in again.

### TC-015 — A revoked session cookie is refused by the video and account endpoints
- **Type:** negative
- **Spec:** REQ-020, SC-009
- **Level:** API
- **Priority:** P1
- **Automation candidate:** yes
- **Preconditions / test data:** Signed in as Viewer A; capture the session cookie value, the video stream URL for a movie, and any account-data URL the app requests while signed in.
- **Action / input:** Sign out; then replay the captured cookie against the captured video stream URL and the account-data URL from a fresh client.
- **Expected result:** Both requests are refused: no video bytes are served and no account data is returned. The old session cannot be used again to watch or to fetch account data. The exact status code is the implementation's; the assertion is that the request is rejected (redirect to sign-in, 401/403, or equivalent).

### TC-016 — Play and pause take effect
- **Type:** positive
- **Spec:** REQ-008, SC-004
- **Level:** end-to-end
- **Priority:** P1
- **Automation candidate:** yes
- **Preconditions / test data:** Signed in as Viewer A on a movie's player page; the video is ready.
- **Action / input:** Press play; wait a few seconds; press pause; wait; press play again.
- **Expected result:** Play starts playback and the timeline advances. Pause stops playback and holds the position (it does not reset to the start). Playing again resumes from that position.

### TC-017 — The timeline scrubber reflects playback position
- **Type:** positive
- **Spec:** REQ-008, SC-004
- **Level:** end-to-end
- **Priority:** P1
- **Automation candidate:** yes
- **Preconditions / test data:** Signed in on the player; playback running.
- **Action / input:** Observe the timeline during playback; drag the scrub handle toward a different point and hold there.
- **Expected result:** A timeline control is present and shows the current position and the movie's total duration. Its indicator advances as the movie plays, and dragging moves the indicator to the dragged position.

### TC-018 — The volume control takes effect during playback
- **Type:** positive
- **Spec:** REQ-008, SC-004
- **Level:** end-to-end
- **Priority:** P2
- **Automation candidate:** yes
- **Preconditions / test data:** Signed in on the player; playback running at an audible volume.
- **Action / input:** Lower the volume; then raise it again.
- **Expected result:** The audible output changes with the control and the control reflects the new level. Playback continues without stopping.

### TC-019 — Fullscreen enters and exits without losing playback state
- **Type:** positive
- **Spec:** REQ-008, SC-004
- **Level:** end-to-end
- **Priority:** P2
- **Automation candidate:** yes
- **Preconditions / test data:** Signed in on the player; playback available.
- **Action / input:** Activate fullscreen; then exit fullscreen.
- **Expected result:** The player expands to fill the screen in fullscreen and returns to the normal page layout on exit. Playback state and position are preserved across both transitions.

### TC-020 — Seeking resumes from the new point rather than restarting
- **Type:** positive
- **Spec:** REQ-009, SC-004
- **Level:** end-to-end
- **Priority:** P1
- **Automation candidate:** yes
- **Preconditions / test data:** Signed in on the player; a movie several minutes long; playback at an early point.
- **Action / input:** Seek forward to a mid-movie point (e.g. about 5:00) and let playback continue; then seek backward to an earlier point (e.g. about 1:00).
- **Expected result:** After each seek, playback resumes at the sought point (within normal buffering tolerance) and continues forward from there. It never restarts from 0:00 or jumps to an unrelated point.

### TC-021 — Playback starts before the whole file has downloaded
- **Type:** positive
- **Spec:** REQ-009
- **Level:** end-to-end
- **Priority:** P2
- **Automation candidate:** yes (requires network throttling)
- **Preconditions / test data:** Signed in; the network is throttled so the movie cannot download instantly; a normal-length movie.
- **Action / input:** Open the player and press play.
- **Expected result:** Playback starts after a short initial buffer while the file is still transferring, and continues to stream as it plays. The viewer is not made to wait for the entire file to download before the video starts.

### TC-022 — Seeking to the very start and near the end behaves correctly
- **Type:** boundary
- **Spec:** REQ-009
- **Level:** end-to-end
- **Priority:** P2
- **Automation candidate:** yes
- **Preconditions / test data:** Signed in on the player; the movie's length is known.
- **Action / input:** Seek to 0:00 and play; then seek to the last few seconds and let playback run out.
- **Expected result:** From 0:00, playback plays from the start. From near the end, playback continues to the end and stops at the ended state (restarting only if the viewer chooses). No seek error, no jump to an unrelated point, no endless buffering.

### TC-023 — A direct player URL while signed out goes through login and back to the movie
- **Type:** negative
- **Spec:** REQ-012
- **Level:** end-to-end
- **Priority:** P1
- **Automation candidate:** yes
- **Preconditions / test data:** No session; a valid player URL for an existing movie; Viewer A's account exists.
- **Action / input:** Open the player URL directly in a browser with no session; sign in on the screen shown.
- **Expected result:** The login / sign-up screen is shown — not the player and not a blank page. After successful sign-in the visitor lands on exactly that movie's player with the video ready to play. The same happens when a player URL is refreshed after signing out.

### TC-024 — A wrong password shows a generic error and keeps the user on the form
- **Type:** negative
- **Spec:** REQ-013, SC-005
- **Level:** end-to-end
- **Priority:** P1
- **Automation candidate:** yes
- **Preconditions / test data:** Viewer A's account exists; the visitor is signed out.
- **Action / input:** Enter `viewer@example.com` with a wrong password; submit.
- **Expected result:** An error message is shown and the visitor is not signed in. They stay on the login form with the email they typed preserved so they can retry. The message does not reveal whether the email exists or whether the password was the wrong part.

### TC-025 — An unknown email produces the same generic error
- **Type:** negative
- **Spec:** REQ-013
- **Level:** end-to-end
- **Priority:** P1
- **Automation candidate:** yes
- **Preconditions / test data:** An email address with no account (e.g. `nobody@example.com`); Viewer A exists for comparison.
- **Action / input:** Enter the unknown email with any password; submit; compare the message with the one from TC-024.
- **Expected result:** An error is shown and the visitor is not signed in. The message uses the same generic wording as for a wrong password — nothing indicates that the email is unknown — and the user stays on the form.

### TC-026 — A duplicate email at sign-up is rejected without creating a second account
- **Type:** negative
- **Spec:** REQ-014, SC-006
- **Level:** end-to-end
- **Priority:** P1
- **Automation candidate:** yes
- **Preconditions / test data:** `viewer@example.com` is already registered with `Passw0rd1`.
- **Action / input:** On the sign-up form, enter `viewer@example.com` with a different valid password (e.g. `Different123`); submit.
- **Expected result:** A clear error states that the email already has an account. No second account is created and the visitor is not signed into the existing account — they remain signed out on the form. Signing in with `viewer@example.com` / `Passw0rd1` still works, and a test-data check confirms exactly one account exists for that address.

### TC-027 — Abandoning the login screen leaves the visitor anonymous and the movie locked
- **Type:** negative
- **Spec:** REQ-016, SC-007
- **Level:** end-to-end
- **Priority:** P1
- **Automation candidate:** yes
- **Preconditions / test data:** No session; the catalogue is present.
- **Action / input:** Select a movie; on the login / sign-up screen navigate back to the catalogue (or dismiss it) without signing in; then select a movie again, and separately open a player URL directly.
- **Expected result:** The visitor remains anonymous and the catalogue stays fully browsable. The movie never plays: every attempt to reach the player shows the login / sign-up screen again.

### TC-028 — A non-existent movie URL shows a clear message and a way back
- **Type:** negative
- **Spec:** REQ-017, SC-010
- **Level:** end-to-end
- **Priority:** P1
- **Automation candidate:** yes
- **Preconditions / test data:** Signed in (so the failure is the missing movie, not the login gate); a movie identifier that does not exist in the seed.
- **Action / input:** Open the player URL with that non-existent movie identifier; follow the recovery action shown.
- **Expected result:** A clear "movie not available" message in English is shown together with a visible way back to the catalogue. There is no blank page, no endless loading and no raw error or stack trace. Following the way back shows the catalogue.

### TC-029 — A video that cannot load shows an error with retry, not an endless spinner
- **Type:** negative
- **Spec:** REQ-018, SC-010
- **Level:** end-to-end
- **Priority:** P1
- **Automation candidate:** yes
- **Preconditions / test data:** Signed in; one movie's video file is unavailable or corrupt (test data).
- **Action / input:** Open that movie's player and attempt to play.
- **Expected result:** Within a finite time after the load attempt fails, the player shows a clear error message and a visible retry action. It does not sit on an endless loading spinner, and it does not present the movie as playable (play does not silently appear to succeed).

### TC-030 — The retry action re-attempts playback
- **Type:** positive
- **Spec:** REQ-018
- **Level:** end-to-end
- **Priority:** P2
- **Automation candidate:** yes
- **Preconditions / test data:** The TC-029 state: signed in, the movie's video file is unavailable.
- **Action / input:** Activate retry while the video is still unavailable; then restore the video file and activate retry again.
- **Expected result:** The first retry makes a fresh attempt and the error message remains with retry still available (no crash, no blank player). Once the video is available, retry loads it and the movie plays from the start.

### TC-031 — An expired session on the player sends the user to login and back to the movie
- **Type:** negative
- **Spec:** REQ-019
- **Level:** end-to-end
- **Priority:** P1
- **Automation candidate:** yes
- **Preconditions / test data:** Signed in on movie X's player page; that session has expired (clock advanced past the 7-day lifetime, or the session prepared as expired in the test environment); the credentials are still valid.
- **Action / input:** Attempt to play (or seek) on the open player page; sign in when prompted.
- **Expected result:** The next playback attempt sends the visitor to the login / sign-up screen. After signing in, they return to movie X's player and playback works. They are not dropped on the catalogue or a generic home page.

### TC-032 — An empty email at sign-up is rejected inline and nothing typed is lost
- **Type:** negative
- **Spec:** REQ-015, REQ-022
- **Level:** end-to-end
- **Priority:** P1
- **Automation candidate:** yes
- **Preconditions / test data:** The login / sign-up screen showing the sign-up option.
- **Action / input:** Leave the email empty, type a valid password, and submit.
- **Expected result:** An inline message on the email field explains that an email is required; no account is created. The password remains in the form, and the visitor can correct the email and submit without retyping the password.

### TC-033 — Malformed email addresses are rejected inline
- **Type:** negative
- **Spec:** REQ-022
- **Level:** end-to-end
- **Priority:** P1
- **Automation candidate:** yes
- **Preconditions / test data:** The sign-up form; none of the addresses is registered; each attempt uses a valid password.
- **Action / input:** Submit, in turn, with each of: `ann.example.com` (no `@`), `ann@` (no domain), `@example.com` (no local part), `ann@@example.com` (two `@`), `ann example@example.com` (space inside the address).
- **Expected result:** Each submission is rejected with an inline message explaining the email is not valid; no account is created for any of them. The typed address stays in the field so it can be corrected. No raw error page appears.

### TC-034 — An empty password at sign-up is rejected inline
- **Type:** negative
- **Spec:** REQ-015, REQ-024
- **Level:** end-to-end
- **Priority:** P1
- **Automation candidate:** yes
- **Preconditions / test data:** The sign-up form; a fresh valid email.
- **Action / input:** Enter the fresh email and an empty password; submit.
- **Expected result:** An inline message explains that a password is required; no account is created. The email stays typed, and the visitor can fill the password and submit.

### TC-035 — A 7-character password is rejected at sign-up
- **Type:** boundary
- **Spec:** REQ-024, SC-011
- **Level:** end-to-end
- **Priority:** P1
- **Automation candidate:** yes
- **Preconditions / test data:** The sign-up form; a fresh email.
- **Action / input:** Password `Abc1234` (7 characters: letters and a digit); submit.
- **Expected result:** Sign-up is rejected with a message explaining that the password must be at least 8 characters; no account is created and the typed values are preserved.

### TC-036 — An exactly 8-character password with a letter and a number is accepted
- **Type:** boundary
- **Spec:** REQ-024
- **Level:** end-to-end
- **Priority:** P1
- **Automation candidate:** yes
- **Preconditions / test data:** The sign-up form; a fresh email.
- **Action / input:** Password `Abc12345` (8 characters: letters and a digit); submit.
- **Expected result:** The account is created and the visitor is signed in — the minimum length boundary is inclusive. No validation error is shown.

### TC-037 — An 8-character password with no number is rejected
- **Type:** negative
- **Spec:** REQ-024, SC-011
- **Level:** end-to-end
- **Priority:** P1
- **Automation candidate:** yes
- **Preconditions / test data:** The sign-up form; a fresh email.
- **Action / input:** Password `Abcdefgh` (8 letters, no number); submit.
- **Expected result:** Sign-up is rejected with a message explaining that the password must contain at least one number; no account is created.

### TC-038 — An 8-character password with no letter is rejected
- **Type:** negative
- **Spec:** REQ-024, SC-011
- **Level:** end-to-end
- **Priority:** P1
- **Automation candidate:** yes
- **Preconditions / test data:** The sign-up form; a fresh email.
- **Action / input:** Password `12345678` (8 digits, no letter); submit.
- **Expected result:** Sign-up is rejected with a message explaining that the password must contain at least one letter; no account is created.

### TC-039 — A longer password with a letter, a number and a symbol is accepted
- **Type:** positive
- **Spec:** REQ-024
- **Level:** end-to-end
- **Priority:** P2
- **Automation candidate:** yes
- **Preconditions / test data:** The sign-up form; a fresh email.
- **Action / input:** Password `Clover2026!` (11 characters: letters, digits and a symbol); submit.
- **Expected result:** The account is created and the visitor is signed in. The spec sets no maximum length and does not ban symbols, so this password is valid.

### TC-040 — An email differing only by letter case is rejected as a duplicate
- **Type:** negative
- **Spec:** REQ-022, SC-012
- **Level:** end-to-end
- **Priority:** P1
- **Automation candidate:** yes
- **Preconditions / test data:** `ann@example.com` is registered with a known password.
- **Action / input:** Sign up with `Ann@Example.com` and a valid password.
- **Expected result:** The sign-up is rejected as an already-existing account — the same duplicate error as TC-026. No second account is created, and the visitor is not signed into the existing account. Signing in with the original address and password still works.

### TC-041 — Sign-in matches the email case-insensitively
- **Type:** positive
- **Spec:** REQ-022, REQ-023
- **Level:** end-to-end
- **Priority:** P1
- **Automation candidate:** yes
- **Preconditions / test data:** `ann@example.com` is registered with password `Passw0rd1`.
- **Action / input:** Sign in with `ANN@EXAMPLE.COM` and the exact password.
- **Expected result:** Sign-in succeeds and the visitor is signed in.

### TC-042 — Sign-in treats the password case-sensitively
- **Type:** negative
- **Spec:** REQ-023, REQ-013
- **Level:** end-to-end
- **Priority:** P1
- **Automation candidate:** yes
- **Preconditions / test data:** An account registered with password `Passw0rd1`.
- **Action / input:** Sign in with the registered email and `passw0rd1` (same letters, different case).
- **Expected result:** Sign-in is rejected with the generic invalid-credentials error; the visitor is not signed in and stays on the form.

### TC-043 — Sign-in does not trim the password
- **Type:** negative
- **Spec:** REQ-023, REQ-013
- **Level:** end-to-end
- **Priority:** P1
- **Automation candidate:** yes
- **Preconditions / test data:** An account registered with password `Passw0rd1`.
- **Action / input:** Sign in first with ` Passw0rd1` (leading space) and then with `Passw0rd1 ` (trailing space).
- **Expected result:** Both attempts are rejected with the generic invalid-credentials error; no silent trimming makes either one succeed.

### TC-044 — A password containing a space is taken exactly as typed
- **Type:** boundary
- **Spec:** REQ-024, REQ-023
- **Level:** end-to-end
- **Priority:** P2
- **Automation candidate:** yes
- **Preconditions / test data:** The sign-up form; a fresh email.
- **Action / input:** Sign up with password `Abc 1234` (8 characters including one internal space, with letters and digits); sign out; sign in with exactly `Abc 1234`; then sign in with `Abc1234`.
- **Expected result:** Sign-up accepts the password exactly as typed and the account is created. Signing in with `Abc 1234` succeeds; signing in with the space removed (`Abc1234`) is rejected. Nothing is trimmed or normalised.

### TC-045 — Empty sign-in fields are rejected inline and typed values are kept
- **Type:** negative
- **Spec:** REQ-015, REQ-023
- **Level:** end-to-end
- **Priority:** P2
- **Automation candidate:** yes
- **Preconditions / test data:** The login / sign-up screen showing the sign-in option; Viewer A's account exists.
- **Action / input:** Submit with both fields empty; then type the email and submit again with the password empty.
- **Expected result:** Inline message(s) explain the missing field(s) and nothing is signed in. Whatever was typed stays in its field, and no raw error page appears.
