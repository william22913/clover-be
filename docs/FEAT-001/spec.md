---
feature_code: FEAT-001
service: clover
ticket_url: null
has_ui: true
version: 1
status: complete
completeness:
  problem: covered
  target_user: covered
  success_criteria: covered
  non_goals: covered
  primary_flow: covered
  edge_cases: covered
  dependencies: covered
  priority: covered
  data_validation_rules: covered
  external_integrations: covered
open_questions: []
sources:
  - path: knowledge/design-reference.md
    source_version: unversioned
    note: Dribbble video the manager supplied; the Clover v1 interface must match it. Owned and read by /ui-preparation, recorded here only so later steps do not lose it.
pr_history: []
pr_url: https://github.com/william22913/sandbox-docs/pull/56
last_updated: 2026-11-13
---

# FEAT-001 — Movie Streaming (clover)

First release of Clover, a movie streaming web app. The whole product goal in
the manager's own words: "open the website, pick a movie, and start watching."

## Problem

There is no Clover product yet. Today a visitor has nowhere to browse a movie
catalogue and nothing to watch on. v1 must make the three-step journey
possible end to end — open the site, pick a movie, watch it — with real
streaming video, not a mockup. The catalogue also has to be usable by someone
who has not signed up, so a first-time visitor can see what is on offer before
being asked for an account.

## Target User

- **Browser visitors** arriving with no account: they can browse the full
  catalogue without any login step. They hit this feature on their very first
  page load, on desktop or mobile browser.
- **Signed-in viewers** who have picked a movie and want to watch it. They hit
  this feature after signing in, and again on return visits when they want to
  carry on watching.

No admin, operator or back-office role exists in v1.

## Primary Flow

- **REQ-001**: A visitor who opens the Clover site sees the movie catalogue
  immediately — a grid of movies, each showing a poster, a title and a short
  description — with no login required and no account prompt in the way.
- **REQ-002**: The catalogue lists every published movie in v1. It is not
  paginated and has no search, filtering or sorting (see Non-Goals).
- **REQ-003**: A visitor can select any movie from the catalogue and open that
  movie's player page.
- **REQ-004**: If the visitor is not signed in when they select a movie, they
  are taken first to a combined login / sign-up screen, and after they
  successfully sign in or create an account they land directly on that same
  movie's player page with the video ready to play — never back on the
  catalogue, and never on a generic home page.
- **REQ-005**: A new visitor can create their own account with an email address
  and a password, from that same screen. No email verification step exists in
  v1; the account is usable immediately after sign-up.
- **REQ-006**: A returning user signs in with the email address and password
  they registered with.
- **REQ-007**: A signed-in user stays signed in between visits and is not asked
  to log in again on every page load (session lifetime is a technical decision,
  recorded below).
- **REQ-008**: The player page streams the selected movie with the standard
  controls a viewer expects: play, pause, a timeline scrubber, volume, and
  fullscreen.
- **REQ-009**: Playback is progressive: the video starts playing without the
  visitor waiting for the whole file to download, and the viewer can seek to
  any point in the movie and playback resumes from there.
- **REQ-010**: A signed-in user can sign out, and after signing out they are
  back to the anonymous browsing experience (catalogue visible, player locked
  per REQ-004).
- **REQ-011**: The entire interface is in English.

## Non-Goals

- No recommendations, personalised or otherwise.
- No subscriptions, payments or billing of any kind.
- No search and no filtering of the catalogue.
- No screen for adding, editing or managing movies. v1 ships a fixed set of
  sample movies (see Dependencies).
- No email verification, password reset, or account recovery in v1.
- No user profiles, avatars, watchlists or viewing history.
- No languages other than English.
- No social features: no sharing, comments or ratings.
- No native mobile app; a browser is the only client.

## Edge Cases

- **REQ-012**: Opening a player page URL directly while signed out (a pasted
  link, a bookmark, a browser refresh after signing out) behaves like
  REQ-004: the visitor is sent to the login screen and returned to that same
  movie after signing in.
- **REQ-013**: Signing in with an unknown email or a wrong password shows a
  clear error message that does not reveal which of the two was wrong, and
  keeps the user on the form so they can try again.
- **REQ-014**: Signing up with an email that already has an account shows a
  clear error, does not create a second account, and does not sign the visitor
  into the existing account.
- **REQ-015**: Sign-up and sign-in input that breaks the rules in Data
  Validation Rules is rejected inline with a message explaining what to fix;
  the user does not lose what they already typed.
- **REQ-016**: If the visitor abandons or dismisses the login screen, they
  remain anonymous and can carry on browsing the catalogue; the movie stays
  locked until they sign in.
- **REQ-017**: A player URL pointing at a movie that does not exist (a stale
  or mistyped link) shows a clear "movie not available" message with a way
  back to the catalogue, rather than a blank page or an error trace.
- **REQ-018**: If the video for a movie cannot be loaded or streamed, the
  player shows a clear error with a retry action — it must not sit on an
  endless loading spinner, and it must not report a broken movie as playable.
- **REQ-019**: If a user's session has expired while they are on a player
  page, the next playback attempt sends them to the login screen and returns
  them to the same movie after signing in (the same behaviour as REQ-012).
- **REQ-020**: Signing out is complete on the server side: after sign-out the
  old session cannot be reused to fetch video or account data, even if the
  browser kept a copy of the cookie.
- **REQ-021**: The same account may be signed in on more than one browser or
  device at the same time; signing in on one device does not sign the user out
  on another.

## Dependencies

- **No other service and no third party.** The catalogue, accounts and video
  files are all part of the Clover Go service itself (see the hand-off notes
  below), so there is no cross-service integration to wait on.
- **Sample content must exist before the feature is usable.** v1 needs a fixed
  set of sample movies, with poster images and public placeholder video files
  that are free to distribute. Where this content lives is decided below.
- **The user's stack decision is a hard constraint**: Go on the server,
  explicitly **without nexcommon and without nexcomponent**, plus a
  browser-based UI, with video served as standard streaming files and nothing
  extra to run.

## Priority

This pass (v1) covers everything in Primary Flow and Edge Cases: browse without
login (REQ-001 to REQ-003), login-gated playback with return-to-movie
(REQ-004, REQ-012, REQ-019), self-service accounts (REQ-005, REQ-006, REQ-010),
session persistence (REQ-007), the player with standard controls and seeking
(REQ-008, REQ-009), English-only UI (REQ-011), and the error/lock behaviours
(REQ-013 to REQ-021).

Explicitly deferred to a later pass:
- Recommendations, subscriptions/payments, search and filtering.
- Any screen for managing the movie catalogue.
- Email verification, password reset, account recovery.
- Profiles, watchlists, viewing history, ratings and comments.

## Success Criteria

Each criterion is directly testable as a pass/fail case.

- **SC-001** (REQ-001, REQ-002): An unauthenticated visitor loading the site's
  first page sees every published v1 movie, each with a poster, a title and a
  short description, with no login prompt required to see them.
- **SC-002** (REQ-004, REQ-005): An unauthenticated visitor selects a movie,
  creates a new account on the login screen, and lands on that same movie's
  player page, signed in, with the video available to play.
- **SC-003** (REQ-004, REQ-006): An unauthenticated visitor selects a movie,
  signs in with an existing account, and lands on that same movie's player
  page.
- **SC-004** (REQ-008, REQ-009): On the player page, play, pause, timeline
  scrub, volume and fullscreen each take effect, and seeking to a new point
  resumes playback from that point rather than restarting the movie.
- **SC-005** (REQ-013): Signing in with a correct email and a wrong password
  returns an error and does not sign the visitor in.
- **SC-006** (REQ-014): Signing up with an already-registered email returns an
  error, and exactly one account exists for that email afterwards.
- **SC-007** (REQ-016): A visitor who abandons the login screen remains
  anonymous and cannot play a movie.
- **SC-008** (REQ-007): A user who signed in, closed the browser and returned
  later is still signed in and can play a movie without re-entering
  credentials.
- **SC-009** (REQ-020): After sign-out, replaying the previously issued session
  cookie against the video or account endpoints is refused.
- **SC-010** (REQ-017, REQ-018): A player URL for a non-existent movie and a
  movie whose video fails to load each show a clear message with a recovery
  action, never a blank page or endless spinner.
- **SC-011** (REQ-024): Sign-up is rejected for a 7-character password, for an
  8-character password with no digit, and for an 8-character password with no
  letter; each rejection explains which rule was broken.
- **SC-012** (REQ-022): Signing up with an address differing from an existing
  account's only by letter case is rejected as a duplicate.

Business metric: none, because v1 has no analytics requirement and no separate
business goal beyond these acceptance criteria.

## Data Validation Rules

Business rules only — how they are enforced is `/code-preparation`'s and
`/go-dev`'s concern.

- **REQ-022**: An account email address is required, must be a syntactically
  valid email address, and is unique across accounts: two accounts cannot share
  an email address, compared case-insensitively (so `Ann@Example.com` and
  `ann@example.com` are the same address).
- **REQ-023**: Sign-in matches the email exactly as registered under the
  case-insensitive rule in REQ-022, and the password must match exactly;
  passwords are never trimmed or case-folded.
- **REQ-024**: A password is required at sign-up, with a minimum of 8
  characters, and it must contain at least one letter and at least one number.
  Whitespace, case and any other characters are taken exactly as typed
  (nothing is trimmed or normalised). This rule applies at account creation
  only; it cannot lock out an account created under a different rule later.
  (source: manager, this session — option (b) of the proposed password rules)
- Movie titles, descriptions and any other catalogue field: **none**, because
  v1 has no screen where a person enters movie data — the catalogue is a fixed
  sample set. A validation rule for a field nobody can type would be filler.

## External Integrations

None, because there is no other service and no third-party API in this
feature. The catalogue, the account store and the video files are all served
by the Clover Go service itself, and the browser talks only to that service.
There is no contract to obtain, and nothing here is inferred from a service
name.

## Hand-off notes for /code-preparation

The manager left the technical shape of v1 to be decided at this stage. These
are decisions, not open questions; `/code-preparation` can refine them in
`architecture.md` but must not contradict the constraint in Dependencies.

- **One Go service, standard library only.** `net/http` for routing and
  handlers; no nexcommon, no nexcomponent, no web framework. This is the
  manager's explicit requirement.
- **Sign-in sessions**: an opaque random session token in an `HttpOnly`,
  `Secure`, `SameSite=Lax` cookie with a 7-day lifetime, with the session
  record held server-side so sign-out (REQ-020) can invalidate it. Chosen over
  a JWT because revocation is a stated requirement and a JWT would need extra
  server state to satisfy it.
- **Passwords**: stored only as a slow one-way hash (bcrypt via
  `golang.org/x/crypto/bcrypt`); the plain password is never stored or logged.
- **Data model**: three tables — `users` (id, email unique case-insensitively,
  password_hash, created_at), `movies` (id, title, short_description,
  poster_url, video_url), `sessions` (id, user_id, expires_at). `users` and
  `sessions` are the only writable tables in v1.
- **Sample data**: seeded from a versioned seed file applied idempotently at
  startup, so a fresh environment comes up with the v1 movies already present.
  No management screen and no manual database step.
- **Video delivery**: files served from local storage through Go's
  `http.ServeContent`, which provides HTTP Range support so seeking works
  (REQ-009) without a streaming server. MP4/H.264 for broad browser support.
- **Posters**: small static images served by the same service.
- **Access control**: the movie catalogue is public; the video stream and any
  account data are reachable only with a valid session (REQ-004, REQ-016,
  REQ-020).

## Knowledge sources

- `knowledge/design-reference.md` — the Dribbble reference the manager supplied
  (the UI must match it). Read and mapped by `/ui-preparation`, which owns
  `UIMap.md`; recorded here so later steps do not lose it.
