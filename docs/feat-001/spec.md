---
feature_code: feat-001
service: clover
ticket_url: null
has_ui: true
app_scale: small
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
    note: Dribbble video the manager supplied as the exact look, button style and transitions for Clover v1; owned and read by /ui-preparation, recorded here only so later steps do not lose it.
pr_history: []
pr_url: https://github.com/william22913/sandbox-docs/pull/68
last_updated: 2026-11-14
---

# feat-001 — Movie Streaming (clover)

First release of Clover, a movie streaming web app. The manager's own words
for the whole product goal: "open the website, pick a movie, and start
watching." (source: manager, scoping session, relayed in the run's session
summary)

## Problem

There is no Clover product yet. The business has a movie catalogue to present,
but today a visitor has nowhere to browse it and nothing to watch on. v1 must
make the whole journey work end to end — open the site, pick a movie, watch it
— with real streaming video, not a mockup. The catalogue must be usable by
someone who has not signed in, so a first-time visitor can see what is on offer
before being asked to sign in.

## Target User

- **Visitors** with no account: they can browse the full catalogue without
  signing in. They meet this feature on their very first page load, on a
  desktop or mobile browser.
- **Signed-in viewers**: holders of a pre-created account who have picked a
  movie and want to watch it. They meet this feature after signing in, and
  again on return visits when they want to carry on watching.

No admin, operator or back-office role exists in v1.

Scale, as confirmed by the manager: a few dozen accounts today, a few hundred
within a year; about 10 concurrent viewers today, up to about 30 within a
year; a catalogue of 20 movies now, 20–50 within a year; movies no longer than
30 minutes; one business with one catalogue (single tenant); a short
maintenance break is acceptable.

## Roles and language

- **Visitor (not signed in)**: may view the catalogue; may not play any movie.
- **Signed-in viewer**: may view the catalogue and play movies.
- Accounts are pre-created; there is no self sign-up in v1.
- The interface is in **English only** in v1. It is built so another language
  can be added later without a rebuild.

## Primary Flow

- **REQ-001**: A visitor opens Clover and immediately sees the movie catalogue —
  a grid of movie cards, each showing a poster, a title and a short
  description. No sign-in prompt blocks the catalogue.
- **REQ-002**: The catalogue lists every movie in the v1 catalogue (20 sample
  movies in total). It is not paginated and has no search, filtering or
  sorting (see Non-Goals).
- **REQ-003**: Selecting a movie while signed in opens that movie's player
  page.
- **REQ-004**: Selecting a movie while not signed in opens the sign-in screen
  first. After a successful sign-in the viewer lands directly on that same
  movie's player page, ready to play — never back on the catalogue and never
  on a generic home page.
- **REQ-005**: Sign-in uses a pre-created account: the viewer enters either the
  account's username or its email address, together with the password.
- **REQ-006**: There is no self sign-up in v1. All accounts are created ahead
  of time (seeded); no visitor can create an account from the site.
- **REQ-007**: A signed-in viewer stays signed in on that device across page
  reloads and browser restarts until they sign out.
- **REQ-008**: A signed-in viewer can sign out. After signing out they return
  to the anonymous browsing experience: the catalogue stays visible and
  playback is locked again (REQ-004).
- **REQ-009**: The player page streams the selected movie with the standard
  controls a viewer expects: play, pause, a timeline scrubber, volume and
  fullscreen.
- **REQ-010**: Playback is progressive: the video starts without waiting for
  the whole file to download, and the viewer can seek to another point and
  playback resumes from there.
- **REQ-011**: Playback is possible only with a valid signed-in session.
  Opening a player address directly while signed out behaves like REQ-004.
- **REQ-012**: The player page shows the movie's title alongside the video, so
  the viewer can tell what they are watching.
- **REQ-013**: Every movie in the v1 catalogue has a working sample video, so
  the player can be tested immediately; all 20 movies can be selected and
  played end to end.
- **REQ-014**: The entire interface is in English.

## Non-Goals

- No self sign-up, password reset, account recovery or account management in
  v1.
- No subscriptions, payments or billing of any kind.
- No recommendation algorithms or personalised feeds.
- No user profiles, avatars, watchlists, ratings, comments or viewing history.
- No search, filtering, sorting or pagination of the catalogue; 20 movies fit
  on one page.
- No screen for adding, editing or managing movies or accounts; the v1
  catalogue and accounts are seeded.
- No uploading, transcoding or editing of films; no CDN and no external video
  platform.
- No languages other than English.
- No native mobile app; a browser is the only client.
- No social sharing.
- No multi-tenant or multi-business catalogue.

## Edge Cases

- **REQ-015**: If the visitor abandons or dismisses the sign-in screen, they
  remain anonymous and can carry on browsing the catalogue; the movie stays
  locked until they sign in.
- **REQ-016**: A wrong username/email or a wrong password shows one generic
  error that does not reveal which of the two was wrong or whether the account
  exists. The form stays usable, the identifier is kept, and the visitor can
  try again.
- **REQ-017**: An empty identifier or an empty password is rejected before the
  request is sent, with a message saying the field is required.
- **REQ-018**: A player address for a movie that does not exist (a stale or
  mistyped link) shows a clear "movie not available" message with a way back
  to the catalogue, never a blank page or an error trace.
- **REQ-019**: If a movie's video cannot be loaded or streamed, the player
  shows a clear error with a retry action. It never sits on an endless loading
  spinner and never presents a broken movie as playable.
- **REQ-020**: If the session is no longer valid while the viewer is on a
  player page, the next playback attempt sends them to the sign-in screen and
  returns them to the same movie after signing in (the same behaviour as
  REQ-004).
- **REQ-021**: Signing out is complete on the server side: after sign-out the
  old session cannot be reused to fetch video or account data, even if the
  browser kept a copy.
- **REQ-022**: The same account may be signed in on more than one browser or
  device at the same time; signing in on one device does not sign the viewer
  out on another.
- **REQ-023**: If the catalogue is empty, the site shows a clear "no movies
  available" state rather than a blank page.
- **REQ-024**: If a poster image fails to load, the card still shows the title
  and short description.

## Dependencies

- **No other service and no third party.** The catalogue, accounts, sessions
  and video files are all part of Clover itself, so there is no cross-service
  integration to wait on. (source: manager, scoping session)
- **Sample content must exist before the feature is usable.** v1 needs a fixed
  set of 20 sample movies, each with a poster image, a title, a short
  description and a free-to-use placeholder video file.
- **Pre-created accounts must exist before sign-in can be tested.** Each
  account has a username, an email address and a password.
- **The manager's stack decision is a hard constraint**: a Go backend,
  explicitly **without nexcommon and without nexcomponent**, plus a separate
  browser-based frontend (no nexcomponent), with video served as plain
  streaming files and nothing extra to run. (source: manager, scoping session)
- **The manager's design reference** is a Dribbble video; `/ui-preparation`
  owns its reading and screen mapping. The link is recorded in
  `knowledge/design-reference.md` so later steps do not lose it.

## Priority

This pass (v1) covers everything in Primary Flow and Edge Cases: browsing
without sign-in (REQ-001, REQ-002), sign-in-gated playback with
return-to-movie (REQ-004, REQ-011, REQ-020), seeded-account sign-in
(REQ-005, REQ-006), session persistence and sign-out (REQ-007, REQ-008,
REQ-021), the player with standard controls and seeking (REQ-009, REQ-010,
REQ-012), sample playback (REQ-013), English-only UI (REQ-014), and the
error/lock behaviours (REQ-015 to REQ-024).

Explicitly deferred to a later pass:
- Self-service accounts, password reset and account recovery.
- Recommendations, subscriptions and payments.
- Search, filtering, sorting and pagination.
- Any screen for managing the movie catalogue or accounts.
- Profiles, watchlists, viewing history, ratings and comments.
- Additional languages, native mobile apps and social features.

This feature is the entire v1 release; there is no second feature code in this
pass.

## Success Criteria

Each criterion is directly testable as a pass/fail case.

- **SC-001** (REQ-001, REQ-002): An unsigned visitor loading the site's first
  page sees all 20 v1 movies, each with a poster, a title and a short
  description, with no sign-in prompt required to see them.
- **SC-002** (REQ-004, REQ-005): An unsigned visitor selects a movie, signs in
  with an account's username and password, and lands on that same movie's
  player page, signed in, with the video available to play.
- **SC-003** (REQ-005): Signing in with the same account's email address and
  password instead of its username also lands on the intended movie's player
  page.
- **SC-004** (REQ-007): A viewer who signed in, reloaded the page and reopened
  the browser is still signed in and can play a movie without re-entering
  credentials.
- **SC-005** (REQ-008, REQ-011): After sign-out the catalogue is still
  visible, selecting a movie opens the sign-in screen, and no video plays
  before signing in again.
- **SC-006** (REQ-009, REQ-010): On the player page, play, pause, timeline
  scrub, volume and fullscreen each take effect, and seeking to a new point
  resumes playback from that point rather than restarting the movie.
- **SC-007** (REQ-016): Signing in with a correct identifier and a wrong
  password returns one generic error, does not sign the visitor in, and leaves
  the form usable for another attempt.
- **SC-008** (REQ-011, REQ-020): Opening a movie's player address directly in
  a signed-out browser sends the visitor to sign-in and then to that same
  movie after signing in.
- **SC-009** (REQ-018, REQ-019): A player address for a non-existent movie and
  a movie whose video fails to load each show a clear message with a recovery
  action — never a blank page or an endless loading spinner.
- **SC-010** (REQ-021): After sign-out, replaying the previously issued
  session against the video or account endpoints is refused.
- **SC-011** (REQ-013): All 20 seeded movies can be selected and play their
  sample video end to end.

Business metric: none, because v1 has no analytics requirement and no separate
business goal beyond these acceptance criteria.

## Data Validation Rules

Business rules only — how they are enforced is `/code-preparation`'s and
`/go-dev`'s concern. All rules below were confirmed by the manager on
2026-11-14, except where noted as a direct consequence of another rule.

- **REQ-025**: The sign-in identifier (username or email) is required and is
  trimmed of surrounding spaces. A username is 3–50 characters using letters,
  digits and underscore only. An email address must be a syntactically valid
  email address. Both are matched case-insensitively. (source: manager,
  2026-11-14 — confirmed as proposed)
- **REQ-026**: The password is required, is 8–128 characters long, and
  contains at least one letter and at least one number. Case, whitespace and
  any other characters are taken exactly as typed; nothing is trimmed or
  case-folded. (source: manager, 2026-11-14 — confirmed as proposed)
- **REQ-027**: The username and the email address of a pre-created account are
  each unique across accounts; no two accounts share either, compared
  case-insensitively. This follows directly from sign-in by username or email
  (REQ-005): a duplicate identifier would make one account unreachable.
- **REQ-028**: Repeated failed sign-ins do not lock the account in v1; the
  visitor may retry immediately. (source: manager, 2026-11-14 — confirmed as
  proposed)
- **REQ-029**: Movie catalogue fields are filled in the seed data only; no
  person types them on the site in v1, so no visitor-facing validation rule
  applies to them. The seed data requires a title, a short description, a
  poster image and a video for every movie, and no movie may exceed 30 minutes
  of video.

## External Integrations

None, because there is no other service and no third-party API in this
feature. The catalogue, the account store and the video files are all served
by the Clover backend itself, and the browser talks only to Clover. There is
no contract to obtain, and nothing here is inferred from a service name.

## Knowledge sources

- `knowledge/design-reference.md` — the Dribbble reference the manager
  supplied (the Clover v1 interface must match it). Read and mapped by
  `/ui-preparation`, which owns `UIMap.md`; recorded here so later steps do not
  lose the link.

## Feedback responses

None yet.
