---
feature_code: feat-001
service: clover
version: 1
status: complete
spec_version: 1
open_questions: []
pr_url: https://github.com/william22913/sandbox-docs/pull/70
last_updated: 2026-11-14
---

# feat-001 — Movie Streaming (Clover v1) — test design

Black-box test design derived from the merged `spec.md` v1 (spec PR #68,
merged into `sandbox-docs` `main` on 2026-11-14) together with the merged
design PR #69 (`AppFlow.md`, `UIMap.md`, `DesignBrief.md`). Every case below
describes observable behaviour and an expected result that follows from those
approved documents; none has been executed.

Where the spec fixes behaviour, that behaviour is the expected result. Where
only the approved design fixes a route or a piece of copy (for example the
"Film not found" screen or the "Signed out." toast), the case says so and
uses the design's wording. No case prescribes implementation, storage, an
endpoint name or a security mechanism.

## Scope and test data

One environment, seeded per spec.md's Dependencies section: 20 films and the
pre-created accounts, with plain streaming video files served by Clover.

**Accounts** (all pre-created; there is no self sign-up). Names are fixture
labels; the test environment creates them with exactly these properties:

| Fixture | Username | Email | Password | Purpose |
|---|---|---|---|---|
| `alex` | `alex` | `Alex.Rivera@example.com` | `Watch1234` | Baseline account |
| `mira` | `mira_k9` | `mira.k@example.com` | `Play2025x` | Second account (multi-device, isolation) |
| `minuser` | `abc` (3 characters) | `abc@example.com` | `Abcdefg1` | Username minimum |
| `maxuser` | `a` × 50 characters | `max.user@example.com` | `Abcdefg1` | Username maximum |
| `minpw` | `minpw` | `minpw@example.com` | `Abcdefg1` (8 characters) | Password minimum |
| `maxpw` | `maxpw` | `maxpw@example.com` | `A1` + `x` × 126 (128 characters) | Password maximum |
| `spacey` | `spacey` | `spacey@example.com` | `Watch1234 ` (trailing space) | Password taken exactly as typed |

**Films.** The 20 seeded films are referred to as **Film A** (the first card
in the catalogue) and **Film B** (any other card). The test environment can
seed additional rows for the following cases only, and never into the approved
v1 catalogue: an empty catalogue (TC-004), a broken poster URL (TC-005), a
broken video URL (TC-059), a catalogue address naming no film (TC-057,
TC-058), and a 30:01 film for the seed boundary (TC-063).

**Environment controls.** Block or restore a poster/video URL at the proxy;
run with an empty film table; expire a live session server-side; throttle the
network to a bandwidth below the sample file's size; keep a persistent browser
profile between runs (browser-restart case).

**App defaults.** Roles in v1 are only *visitor* (signed out) and *signed-in
viewer*; there is no permission hierarchy (`<parent>:all` style) and no
owner-scoped record — the catalogue is shared and a viewer can only act on
their own session, so role and ownership cases reduce to TC-008 to TC-022 and
TC-061. The interface offers exactly one language, English (TC-006, TC-007).
`app_scale` is `small`, so the suite adds no pagination, rate-limit or
replica-failover cases beyond what REQ-002 itself requires (TC-002).

## Requirement coverage

| REQ | Requirement (short) | Cases | Notes |
|---|---|---|---|
| REQ-001 | Anonymous visitor sees the catalogue (poster, title, description) with nothing blocking it | TC-001, TC-003 | — |
| REQ-002 | All 20 films on one page; no pagination, search, filtering or sorting | TC-001, TC-002 | The non-goals that REQ-002 itself restates are asserted as absence of controls |
| REQ-003 | Signed-in film selection opens that film's player | TC-011 | — |
| REQ-004 | Signed-out film selection opens sign-in, then the same film's player | TC-008, TC-009 | — |
| REQ-005 | Sign-in with username **or** email plus password | TC-009, TC-010, TC-013 | — |
| REQ-006 | No self sign-up; all accounts pre-created | TC-014 | — |
| REQ-007 | Session persists across reloads and browser restarts until sign-out | TC-017, TC-018 | — |
| REQ-008 | Sign-out returns to anonymous browsing and locks playback again | TC-019 | — |
| REQ-009 | Player offers play, pause, timeline, volume and fullscreen | TC-050, TC-051, TC-053, TC-054 | — |
| REQ-010 | Progressive playback; seeking resumes from the chosen point | TC-052, TC-055 | — |
| REQ-011 | Playback only with a valid session; direct player address while signed out behaves like REQ-004 | TC-015, TC-061 | — |
| REQ-012 | Film title shown beside the video | TC-049 | — |
| REQ-013 | Every v1 film has a working sample video and plays end to end | TC-056 | — |
| REQ-014 | Interface entirely in English | TC-006, TC-007 | — |
| REQ-015 | Abandoning sign-in keeps the visitor anonymous and the film locked | TC-012 | — |
| REQ-016 | Wrong identifier or password shows one generic error; form stays usable; identifier kept | TC-023, TC-024 | — |
| REQ-017 | Empty identifier or password rejected before the request with a "required" message | TC-025, TC-026, TC-027 | — |
| REQ-018 | Unknown film address shows "movie not available" and a way back | TC-057, TC-058 | — |
| REQ-019 | Unplayable video shows a clear error with retry; never an endless spinner | TC-059, TC-060 | — |
| REQ-020 | Invalid session on the player sends the viewer to sign-in, then back to the same film | TC-016 | — |
| REQ-021 | Sign-out invalidates the session server-side for video and account data | TC-020, TC-021 | — |
| REQ-022 | Same account may be signed in on several devices at once | TC-022 | — |
| REQ-023 | Empty catalogue shows a clear "no movies available" state | TC-004 | — |
| REQ-024 | A failed poster still leaves the title and description visible | TC-005 | — |
| REQ-025 | Identifier required and trimmed; username 3–50 of letters/digits/underscore; valid email; case-insensitive | TC-027 to TC-038 | Boundary value analysis on length and character set, plus equivalence classes for email |
| REQ-026 | Password required; 8–128 characters; at least one letter and one digit; taken exactly as typed | TC-039 to TC-046 | Boundary value analysis on length and composition |
| REQ-027 | Usernames and emails are unique across accounts, compared case-insensitively | TC-048 | Seed-data rule; verified directly against the seeded account store |
| REQ-028 | No account lockout after repeated failed sign-ins in v1 | TC-047 | — |
| REQ-029 | Seed catalogue requires title, description, poster and video; no film over 30 minutes | TC-062, TC-063 | Seed-data rule with a duration boundary |

## Success criteria cross-check

Each success criterion in spec.md maps to at least one case; this table is a
reading aid, not a second coverage matrix.

| SC | Acceptance criterion (short) | Cases |
|---|---|---|
| SC-001 | Anonymous visitor sees all 20 films with poster, title and description | TC-001, TC-002 |
| SC-002 | Pick a film signed out, sign in with username, land on that film's player | TC-009 |
| SC-003 | Same via email address | TC-010 |
| SC-004 | Signed in survives reload and browser restart | TC-017, TC-018 |
| SC-005 | After sign-out the catalogue stays public and playback locks again | TC-019 |
| SC-006 | Play, pause, scrub, volume and fullscreen take effect; seek resumes from the point | TC-050 to TC-055 |
| SC-007 | Correct identifier + wrong password gives one generic error and leaves the form usable | TC-023 |
| SC-008 | Direct player address signed out → sign-in → the same film | TC-015, TC-016 |
| SC-009 | Unknown film and unplayable video each show a clear message with recovery | TC-057, TC-059 |
| SC-010 | A post-sign-out session replayed against video/account endpoints is refused | TC-020, TC-021 |
| SC-011 | All 20 seeded films can be selected and played end to end | TC-056 |

## Non-goals review

The spec's Non-Goals are deferred capability, not behaviour with a defined
rejection: self sign-up, password reset, account recovery and account
management (TC-014 covers the one that *is* a requirement, REQ-006: no
visitor can create an account); subscriptions, payments and billing;
recommendations and personalised feeds; profiles, avatars, watchlists,
ratings, comments and viewing history; search, filtering, sorting and
pagination (asserted as absent under REQ-002 in TC-002); any screen for
managing films or accounts; uploading, transcoding or editing films, CDN and
external video platforms; other languages (REQ-014 fixes English only);
native mobile apps; social sharing; multi-tenant catalogues.

Only REQ-002 and REQ-006 state a boundary the product can actually be seen to
hold (no such control exists / no account-creation path exists), so only those
two get cases. The rest are absence of a feature with no specified rejection
behaviour and are closed by this manual scope review rather than by
speculative cases.

## Test cases

### Public catalogue, empty and error states

#### TC-001 — An anonymous visitor sees the whole catalogue
- **Type:** positive
- **Spec:** REQ-001, REQ-002
- **Level:** UI
- **Priority:** P1
- **Automation candidate:** yes
- **Preconditions / test data:** Fresh browser profile with no session; the seeded 20-film catalogue.
- **Action / input:** Open the site's root address.
- **Expected result:** The catalogue renders as a grid. Each of the 20 cards shows a poster image, the film's title and a short description. The header offers "Sign in". No sign-in prompt, modal or overlay blocks the catalogue, and no card is hidden or locked.

#### TC-002 — The catalogue fits on one page with no search, filter, sort or pagination
- **Type:** positive
- **Spec:** REQ-002
- **Level:** UI
- **Priority:** P2
- **Automation candidate:** yes
- **Preconditions / test data:** Seeded 20-film catalogue; anonymous visitor.
- **Action / input:** Load the catalogue and scroll from top to bottom.
- **Expected result:** All 20 films are present in a single scrollable page; the count reads "20 films". There is no pagination control, no "next"/"previous" control, no search field, no filter control and no sort control anywhere on the page.

#### TC-003 — Catalogue data is readable without a session
- **Type:** positive
- **Spec:** REQ-001
- **Level:** API
- **Priority:** P1
- **Automation candidate:** yes
- **Preconditions / test data:** No session cookie or token; seeded 20-film catalogue.
- **Action / input:** Request the catalogue data the page loads, with no credentials.
- **Expected result:** The request succeeds (HTTP 200) and returns 20 film records, each with a title, a short description, a poster reference and a video reference. No authentication challenge is returned.

#### TC-004 — An empty catalogue shows a clear empty state
- **Type:** negative
- **Spec:** REQ-023
- **Level:** UI
- **Priority:** P2
- **Automation candidate:** yes
- **Preconditions / test data:** Environment running with zero films in the film store.
- **Action / input:** Open the site's root address.
- **Expected result:** A clear "No films available" state is shown (the approved design's wording: "No films available"). The page is not blank, shows no error trace and no broken cards, and the catalogue count does not claim a non-zero number.

#### TC-005 — A failed poster leaves the card usable
- **Type:** negative
- **Spec:** REQ-024
- **Level:** UI
- **Priority:** P2
- **Automation candidate:** yes
- **Preconditions / test data:** One film's poster URL blocked at the proxy; seeded catalogue.
- **Action / input:** Load the catalogue.
- **Expected result:** The affected card shows the design's "Poster unavailable" placeholder in the poster area. The film's title and short description are still visible, and the whole card is still selectable.

#### TC-006 — The interface is in English
- **Type:** positive
- **Spec:** REQ-014
- **Level:** UI
- **Priority:** P2
- **Automation candidate:** yes
- **Preconditions / test data:** Seeded environment; visitor and signed-in viewer.
- **Action / input:** Visit the catalogue, the sign-in screen, the player, the film-not-found screen and the error states.
- **Expected result:** All interface text — headings, labels, buttons, placeholders, validation messages, error messages and toasts — is in English. There is no language switch or locale selector anywhere.

#### TC-007 — An unsupported or missing language still gets English
- **Type:** boundary
- **Spec:** REQ-014
- **Level:** API
- **Priority:** P3
- **Automation candidate:** yes
- **Preconditions / test data:** Seeded environment.
- **Action / input:** Request the catalogue and the sign-in screen with `Accept-Language: fr-FR`, then with no `Accept-Language` header, then with an unsupported tag such as `xx-XX`.
- **Expected result:** In every case the response is the English interface: English copy in the returned content and English messages. No other language is returned and nothing falls back to an empty or untranslated string.

### Sign-in gate and navigation

#### TC-008 — Selecting a film while signed out opens sign-in for that film
- **Type:** positive
- **Spec:** REQ-004
- **Level:** end-to-end
- **Priority:** P1
- **Automation candidate:** yes
- **Preconditions / test data:** Anonymous visitor; Film A and Film B in the catalogue.
- **Action / input:** Select Film A, then select Film B on a second attempt.
- **Expected result:** Each selection opens the sign-in screen first; no video plays. The sign-in screen names the film that was picked (the design's target chip shows "Signing in to watch" and the film's title) and that film is the one carried forward.

#### TC-009 — Username sign-in lands on the picked film's player
- **Type:** positive
- **Spec:** REQ-004, REQ-005
- **Level:** end-to-end
- **Priority:** P1
- **Automation candidate:** yes
- **Preconditions / test data:** Anonymous visitor; account `alex` / `Watch1234`; Film B selected.
- **Action / input:** Select Film B, enter username `alex` and password `Watch1234`, submit.
- **Expected result:** The viewer is signed in and lands directly on Film B's player page — never back on the catalogue and never on a generic home page. The player is ready to play and shows Film B's title.

#### TC-010 — Email sign-in lands on the picked film's player
- **Type:** positive
- **Spec:** REQ-005
- **Level:** end-to-end
- **Priority:** P1
- **Automation candidate:** yes
- **Preconditions / test data:** Anonymous visitor; account `alex` (`Alex.Rivera@example.com` / `Watch1234`); Film A selected.
- **Action / input:** Select Film A, enter email `Alex.Rivera@example.com` and password `Watch1234`, submit.
- **Expected result:** The viewer is signed in to the same account and lands directly on Film A's player page, ready to play.

#### TC-011 — Selecting a film while signed in opens the player directly
- **Type:** positive
- **Spec:** REQ-003
- **Level:** end-to-end
- **Priority:** P1
- **Automation candidate:** yes
- **Preconditions / test data:** Signed-in viewer (`alex`).
- **Action / input:** From the catalogue, select Film B.
- **Expected result:** Film B's player page opens immediately. No sign-in screen is shown.

#### TC-012 — Cancelling sign-in keeps the visitor anonymous
- **Type:** positive
- **Spec:** REQ-015
- **Level:** end-to-end
- **Priority:** P2
- **Automation candidate:** yes
- **Preconditions / test data:** Anonymous visitor; Film A selected.
- **Action / input:** On the sign-in screen, use "Cancel and keep browsing".
- **Expected result:** The catalogue is shown again and the visitor is still anonymous (the header still offers "Sign in"). Selecting Film A again opens sign-in again; the film does not play.

#### TC-013 — Signing in from the header without a picked film returns to the catalogue
- **Type:** positive
- **Spec:** REQ-005
- **Level:** end-to-end
- **Priority:** P2
- **Automation candidate:** yes
- **Preconditions / test data:** Anonymous visitor; account `alex`.
- **Action / input:** Use the header's "Sign in", then sign in with `alex` / `Watch1234`.
- **Expected result:** The sign-in screen shows no picked-film target chip. After a successful sign-in the catalogue is shown with the viewer signed in (the approved design shows a "Signed in. Pick a film to watch." confirmation). No player opens, because no film was picked.

#### TC-014 — There is no way to create an account
- **Type:** negative
- **Spec:** REQ-006
- **Level:** UI
- **Priority:** P2
- **Automation candidate:** yes
- **Preconditions / test data:** Anonymous visitor.
- **Action / input:** Inspect the catalogue and the sign-in screen for any sign-up, register or "create account" control; attempt a registration route such as `/sign-up` and `/register`.
- **Expected result:** No sign-up, register or account-creation control exists anywhere on the site. The sign-in screen states that accounts are created by Clover ("No sign-up in v1. Accounts are created by Clover."). Registration addresses do not create an account; they return a not-found response or the catalogue, never a registration form.

#### TC-015 — A direct player address while signed out asks for sign-in, then plays that film
- **Type:** positive
- **Spec:** REQ-011
- **Level:** end-to-end
- **Priority:** P1
- **Automation candidate:** yes
- **Preconditions / test data:** Anonymous browser; Film B's player address known.
- **Action / input:** Open Film B's player address directly; sign in as `alex` / `Watch1234`.
- **Expected result:** The sign-in screen appears with Film B carried (target chip names Film B), and no video plays before sign-in. After a successful sign-in, Film B's player opens, ready to play — the same result as picking the film in the catalogue (TC-009).

#### TC-016 — A session that becomes invalid while on the player is handled like REQ-004
- **Type:** negative
- **Spec:** REQ-020
- **Level:** end-to-end
- **Priority:** P1
- **Automation candidate:** yes
- **Preconditions / test data:** Signed-in viewer on Film A's player; the test environment expires that session server-side without the page knowing.
- **Action / input:** Press play (or any playback action).
- **Expected result:** The next playback attempt sends the viewer to the sign-in screen with Film A carried. After signing in again, Film A's player reopens, ready to play. The viewer is not left on a dead player, and no error trace is shown.

### Session persistence and sign-out

#### TC-017 — The session survives a page reload
- **Type:** positive
- **Spec:** REQ-007
- **Level:** end-to-end
- **Priority:** P1
- **Automation candidate:** yes
- **Preconditions / test data:** Viewer signed in as `alex`.
- **Action / input:** Reload the catalogue, then select Film A.
- **Expected result:** The viewer is still signed in after the reload (no sign-in screen), and Film A's player opens and plays without re-entering credentials.

#### TC-018 — The session survives a browser restart
- **Type:** positive
- **Spec:** REQ-007
- **Level:** end-to-end
- **Priority:** P1
- **Automation candidate:** yes
- **Preconditions / test data:** Persistent browser profile; viewer signed in as `alex`.
- **Action / input:** Close the browser completely, reopen it with the same profile, and open the catalogue; then select Film B.
- **Expected result:** The viewer is still signed in and can play Film B without entering credentials again. The session only ends through sign-out (or actual expiry).

#### TC-019 — Signing out returns to public browsing and locks playback again
- **Type:** positive
- **Spec:** REQ-008
- **Level:** end-to-end
- **Priority:** P1
- **Automation candidate:** yes
- **Preconditions / test data:** Signed-in viewer (`alex`); Films A and B visible.
- **Action / input:** Use the header's "Sign out", then select Film A.
- **Expected result:** The catalogue remains visible and complete; the header offers "Sign in" again and the design's "Signed out." confirmation is shown. Selecting Film A opens the sign-in screen and no video plays before signing in again.

#### TC-020 — A signed-out session cannot fetch video
- **Type:** negative
- **Spec:** REQ-021
- **Level:** API
- **Priority:** P1
- **Automation candidate:** yes
- **Preconditions / test data:** Viewer signed in as `alex`; capture the session cookie/token; then sign out.
- **Action / input:** Replay the captured session against the video/file request for Film A, using a normal GET and a ranged request.
- **Expected result:** Both requests are refused (HTTP 401 or 403) and return no video bytes. The refusal does not reveal film contents, and it is not satisfied by any browser-cached copy.

#### TC-021 — A signed-out session cannot fetch account data
- **Type:** negative
- **Spec:** REQ-021
- **Level:** API
- **Priority:** P1
- **Automation candidate:** yes
- **Preconditions / test data:** Viewer signed in as `alex`; capture the session cookie/token; then sign out.
- **Action / input:** Replay the captured session against the request that returns the signed-in account's data.
- **Expected result:** The request is refused (HTTP 401 or 403) with a clean error body and no account data. A fresh, valid session for `alex` still works, proving only the old session was invalidated.

#### TC-022 — The same account can be signed in on two devices at once
- **Type:** positive
- **Spec:** REQ-022
- **Level:** end-to-end
- **Priority:** P2
- **Automation candidate:** yes
- **Preconditions / test data:** Two independent browser profiles, both anonymous; account `mira`.
- **Action / input:** Sign in as `mira` in profile 1, then sign in as `mira` in profile 2; play a film in profile 1.
- **Expected result:** Profile 2 signs in successfully, and signing in there does not sign profile 1 out: profile 1 is still signed in and can still play the film. Both sessions are valid at the same time.

### Sign-in validation

#### TC-023 — A wrong password gives one generic error and leaves the form usable
- **Type:** negative
- **Spec:** REQ-016
- **Level:** UI
- **Priority:** P1
- **Automation candidate:** yes
- **Preconditions / test data:** Anonymous visitor; `alex` selected; password `Wrong1234`.
- **Action / input:** Enter `alex` and `Wrong1234`, submit; then correct the password to `Watch1234` and submit again.
- **Expected result:** The first attempt is not signed in and shows exactly one generic message, "Incorrect username/email or password." It does not say whether the identifier or the password was wrong. The identifier `alex` is still filled in and the form is usable: the second attempt with the correct password succeeds and lands on the player for the picked film.

#### TC-024 — An identifier that matches no account gives the same generic error
- **Type:** negative
- **Spec:** REQ-016
- **Level:** UI
- **Priority:** P1
- **Automation candidate:** yes
- **Preconditions / test data:** Anonymous visitor; `nobody_here` is not an account; password `Watch1234`.
- **Action / input:** Submit `nobody_here` / `Watch1234`, then `nobody_here@example.com` / `Watch1234`.
- **Expected result:** Each attempt is refused with the same generic "Incorrect username/email or password." message as TC-023 — same text, no hint that the account does not exist and no hint about which field was wrong.

#### TC-025 — An empty identifier is blocked before the request is sent
- **Type:** negative
- **Spec:** REQ-017
- **Level:** UI
- **Priority:** P1
- **Automation candidate:** yes
- **Preconditions / test data:** Sign-in screen open; password `Watch1234`; identifier left empty.
- **Action / input:** Submit with the identifier empty.
- **Expected result:** A required-field message is shown for the identifier ("Enter your username or email."), focus moves to that field, and no sign-in request is sent (the network log shows no sign-in call). The visitor stays on the sign-in screen and remains signed out.

#### TC-026 — An empty password is blocked before the request is sent
- **Type:** negative
- **Spec:** REQ-017
- **Level:** UI
- **Priority:** P1
- **Automation candidate:** yes
- **Preconditions / test data:** Sign-in screen open; identifier `alex`; password left empty.
- **Action / input:** Submit with the password empty.
- **Expected result:** A required-field message is shown for the password ("Enter your password."), focus moves to that field, and no sign-in request is sent. The visitor stays on the sign-in screen.

#### TC-027 — A whitespace-only identifier is treated as empty
- **Type:** boundary
- **Spec:** REQ-017, REQ-025
- **Level:** UI
- **Priority:** P2
- **Automation candidate:** yes
- **Preconditions / test data:** Sign-in screen open; password `Watch1234`; identifier entered as four spaces.
- **Action / input:** Submit with the identifier `"    "` (spaces only).
- **Expected result:** After trimming, the identifier is empty: the same required-field message as TC-025 is shown, no sign-in request is sent, and the visitor is not signed in.

#### TC-028 — Surrounding spaces are trimmed from the identifier
- **Type:** boundary
- **Spec:** REQ-025
- **Level:** end-to-end
- **Priority:** P2
- **Automation candidate:** yes
- **Preconditions / test data:** Anonymous visitor; account `alex`; Film A picked.
- **Action / input:** Enter the identifier as `"  alex  "` (two spaces each side) with password `Watch1234`, submit.
- **Expected result:** The identifier is trimmed and matches `alex`; the viewer is signed in and lands on Film A's player.

#### TC-029 — A 3-character username is accepted
- **Type:** boundary
- **Spec:** REQ-025
- **Level:** end-to-end
- **Priority:** P2
- **Automation candidate:** yes
- **Preconditions / test data:** Account `minuser` (`abc` / `Abcdefg1`); anonymous visitor; Film A picked.
- **Action / input:** Sign in with `abc` / `Abcdefg1`.
- **Expected result:** Sign-in succeeds and Film A's player opens.

#### TC-030 — A 2-character username is rejected
- **Type:** boundary
- **Spec:** REQ-025
- **Level:** UI
- **Priority:** P2
- **Automation candidate:** yes
- **Preconditions / test data:** Anonymous visitor; identifier `ab` (2 characters).
- **Action / input:** Sign in with `ab` / `Watch1234`.
- **Expected result:** The sign-in is refused with the generic "Incorrect username/email or password." message; the visitor is not signed in and no account is resolved.

#### TC-031 — A 50-character username is accepted
- **Type:** boundary
- **Spec:** REQ-025
- **Level:** end-to-end
- **Priority:** P2
- **Automation candidate:** yes
- **Preconditions / test data:** Account `maxuser` with username `a` × 50 and password `Abcdefg1`; anonymous visitor.
- **Action / input:** Sign in with that 50-character username / `Abcdefg1`.
- **Expected result:** Sign-in succeeds and the player for the picked film opens.

#### TC-032 — A 51-character username is rejected
- **Type:** boundary
- **Spec:** REQ-025
- **Level:** UI
- **Priority:** P2
- **Automation candidate:** yes
- **Preconditions / test data:** Anonymous visitor; identifier `a` × 51.
- **Action / input:** Sign in with that 51-character identifier / `Watch1234`.
- **Expected result:** The sign-in is refused with the generic error; the visitor is not signed in.

#### TC-033 — Letters, digits and underscore are all allowed in a username
- **Type:** positive
- **Spec:** REQ-025
- **Level:** end-to-end
- **Priority:** P2
- **Automation candidate:** yes
- **Preconditions / test data:** Account `mira` (`mira_k9` / `Play2025x`); anonymous visitor.
- **Action / input:** Sign in with `mira_k9` / `Play2025x`.
- **Expected result:** Sign-in succeeds; the viewer is signed in and lands on the picked film's player. The underscore and digit are accepted as part of the username.

#### TC-034 — A username containing a disallowed character is rejected
- **Type:** negative
- **Spec:** REQ-025
- **Level:** UI
- **Priority:** P2
- **Automation candidate:** yes
- **Preconditions / test data:** Anonymous visitor; inputs `mira-k` (hyphen), `mira k` (space), `mira.k` (dot) and `mira@k` (at sign).
- **Action / input:** Submit each identifier with password `Play2025x`.
- **Expected result:** Every input is refused with the generic "Incorrect username/email or password." message; none signs the visitor in. No format-specific message is shown (the approved design answers format problems with the same generic alert).

#### TC-035 — A syntactically valid email is accepted
- **Type:** positive
- **Spec:** REQ-025
- **Level:** end-to-end
- **Priority:** P1
- **Automation candidate:** yes
- **Preconditions / test data:** Account `mira` (`mira.k@example.com` / `Play2025x`); anonymous visitor.
- **Action / input:** Sign in with `mira.k@example.com` / `Play2025x`.
- **Expected result:** Sign-in succeeds; the viewer lands on the picked film's player.

#### TC-036 — A malformed email is rejected
- **Type:** negative
- **Spec:** REQ-025
- **Level:** UI
- **Priority:** P2
- **Automation candidate:** yes
- **Preconditions / test data:** Anonymous visitor; inputs `alex`, `alex@`, `@example.com`, `alex@@example.com`, `alex example.com`.
- **Action / input:** Submit each input with password `Watch1234`.
- **Expected result:** Every malformed input is refused with the generic error; none signs the visitor in and no format-specific message reveals which rule failed.

#### TC-037 — The username is matched case-insensitively
- **Type:** boundary
- **Spec:** REQ-025
- **Level:** end-to-end
- **Priority:** P2
- **Automation candidate:** yes
- **Preconditions / test data:** Account `alex`; anonymous visitor.
- **Action / input:** Sign in with `ALEX` / `Watch1234`.
- **Expected result:** Sign-in succeeds and the viewer lands on the picked film's player.

#### TC-038 — The email address is matched case-insensitively
- **Type:** boundary
- **Spec:** REQ-025
- **Level:** end-to-end
- **Priority:** P2
- **Automation candidate:** yes
- **Preconditions / test data:** Account `alex` (`Alex.Rivera@example.com`); anonymous visitor.
- **Action / input:** Sign in with `ALEX.RIVERA@EXAMPLE.COM` / `Watch1234`.
- **Expected result:** Sign-in succeeds and the viewer lands on the picked film's player.

#### TC-039 — An 8-character password is accepted
- **Type:** boundary
- **Spec:** REQ-026
- **Level:** end-to-end
- **Priority:** P2
- **Automation candidate:** yes
- **Preconditions / test data:** Account `minpw` (`minpw` / `Abcdefg1`, 8 characters); anonymous visitor.
- **Action / input:** Sign in with `minpw` / `Abcdefg1`.
- **Expected result:** Sign-in succeeds and the player opens.

#### TC-040 — A 7-character password is rejected
- **Type:** boundary
- **Spec:** REQ-026
- **Level:** UI
- **Priority:** P2
- **Automation candidate:** yes
- **Preconditions / test data:** Anonymous visitor; identifier `alex`; password `Abcdef1` (7 characters).
- **Action / input:** Sign in with `alex` / `Abcdef1`.
- **Expected result:** The attempt is refused with the generic "Incorrect username/email or password." message; the visitor is not signed in.

#### TC-041 — A 128-character password is accepted
- **Type:** boundary
- **Spec:** REQ-026
- **Level:** end-to-end
- **Priority:** P2
- **Automation candidate:** yes
- **Preconditions / test data:** Account `maxpw` with password `A1` + `x` × 126 (128 characters); anonymous visitor.
- **Action / input:** Sign in with `maxpw` / that 128-character password.
- **Expected result:** Sign-in succeeds and the player opens.

#### TC-042 — A 129-character password is rejected
- **Type:** boundary
- **Spec:** REQ-026
- **Level:** UI
- **Priority:** P2
- **Automation candidate:** yes
- **Preconditions / test data:** Anonymous visitor; identifier `maxpw`; password `A1` + `x` × 127 (129 characters).
- **Action / input:** Sign in with `maxpw` / that 129-character password.
- **Expected result:** The attempt is refused with the generic error; the visitor is not signed in.

#### TC-043 — A password with letters but no digit is rejected
- **Type:** negative
- **Spec:** REQ-026
- **Level:** UI
- **Priority:** P2
- **Automation candidate:** yes
- **Preconditions / test data:** Anonymous visitor; identifier `alex`; password `Abcdefgh` (8 letters, no digit).
- **Action / input:** Sign in with `alex` / `Abcdefgh`.
- **Expected result:** The attempt is refused with the generic error; the visitor is not signed in.

#### TC-044 — A password with digits but no letter is rejected
- **Type:** negative
- **Spec:** REQ-026
- **Level:** UI
- **Priority:** P2
- **Automation candidate:** yes
- **Preconditions / test data:** Anonymous visitor; identifier `alex`; password `12345678`.
- **Action / input:** Sign in with `alex` / `12345678`.
- **Expected result:** The attempt is refused with the generic error; the visitor is not signed in.

#### TC-045 — The password is case-sensitive
- **Type:** negative
- **Spec:** REQ-026
- **Level:** UI
- **Priority:** P2
- **Automation candidate:** yes
- **Preconditions / test data:** Account `alex` with password `Watch1234`; anonymous visitor.
- **Action / input:** Sign in with `alex` / `watch1234`.
- **Expected result:** The attempt is refused with the generic error; nothing is case-folded and the visitor is not signed in.

#### TC-046 — Whitespace in a password is significant and is not trimmed
- **Type:** boundary
- **Spec:** REQ-026
- **Level:** end-to-end
- **Priority:** P3
- **Automation candidate:** yes
- **Preconditions / test data:** Account `spacey` with password `Watch1234 ` (one trailing space, 10 characters); anonymous visitor.
- **Action / input:** Sign in with `spacey` / `Watch1234 ` (trailing space), then in a second attempt with `spacey` / `Watch1234` (no trailing space).
- **Expected result:** The first attempt, typed exactly as stored including the trailing space, succeeds and the player opens. The second attempt is refused with the generic error, proving the password is neither trimmed nor normalised.

#### TC-047 — Repeated failed sign-ins do not lock the account
- **Type:** negative
- **Spec:** REQ-028
- **Level:** end-to-end
- **Priority:** P2
- **Automation candidate:** yes
- **Preconditions / test data:** Account `alex`; anonymous visitor.
- **Action / input:** Submit 10 consecutive wrong passwords for `alex`, then submit the correct password `Watch1234` immediately.
- **Expected result:** All 10 failures show the same generic error with no lockout message and no increasing delay or disabled form. The 11th attempt with the correct password succeeds at once and lands on the player; the account is not locked.

#### TC-048 — Seeded usernames and emails are unique case-insensitively
- **Type:** positive
- **Spec:** REQ-027
- **Level:** API
- **Priority:** P2
- **Automation candidate:** yes
- **Preconditions / test data:** The seeded account store.
- **Action / input:** Read every seeded account's username and email and compare them case-insensitively.
- **Expected result:** No two accounts share a username and no two share an email, compared case-insensitively. Each seeded identifier resolves to exactly one account when used at sign-in.

### Player

#### TC-049 — The player shows the film's title beside the video
- **Type:** positive
- **Spec:** REQ-012
- **Level:** UI
- **Priority:** P2
- **Automation candidate:** yes
- **Preconditions / test data:** Signed-in viewer; Film B.
- **Action / input:** Open Film B's player.
- **Expected result:** The title shown alongside the video is exactly Film B's catalogue title, together with its short description. The title does not belong to a different film.

#### TC-050 — Play starts playback
- **Type:** positive
- **Spec:** REQ-009
- **Level:** UI
- **Priority:** P1
- **Automation candidate:** yes
- **Preconditions / test data:** Signed-in viewer on Film A's player, paused at the start.
- **Action / input:** Press play.
- **Expected result:** The video starts playing, the elapsed timecode advances, and the control shows the pause state.

#### TC-051 — Pause stops playback
- **Type:** positive
- **Spec:** REQ-009
- **Level:** UI
- **Priority:** P1
- **Automation candidate:** yes
- **Preconditions / test data:** Signed-in viewer on Film A's player, playing.
- **Action / input:** Press pause, wait three seconds, then read the elapsed timecode.
- **Expected result:** Playback stops; the elapsed timecode does not advance while paused and the control shows the play state.

#### TC-052 — Seeking resumes playback from the chosen point
- **Type:** positive
- **Spec:** REQ-010
- **Level:** end-to-end
- **Priority:** P1
- **Automation candidate:** yes
- **Preconditions / test data:** Signed-in viewer on a film longer than 15 minutes, playing.
- **Action / input:** Drag the timeline to about 10:00 (two-thirds of a 15-minute film) and continue playing.
- **Expected result:** The timecode jumps to the chosen point and playback continues from there. The film does not restart from 0:00 and does not fall back to the beginning after resuming.

#### TC-053 — Volume and mute take effect
- **Type:** positive
- **Spec:** REQ-009
- **Level:** UI
- **Priority:** P2
- **Automation candidate:** yes
- **Preconditions / test data:** Signed-in viewer on Film A's player, playing, volume at 100%.
- **Action / input:** Lower the volume to about 30%, then mute, then unmute.
- **Expected result:** The volume level changes audibly and the control reflects the new level; mute silences playback and shows the muted state; unmute restores the previous volume (about 30%), not 0% and not 100%.

#### TC-054 — Fullscreen enters and exits
- **Type:** positive
- **Spec:** REQ-009
- **Level:** UI
- **Priority:** P2
- **Automation candidate:** yes
- **Preconditions / test data:** Signed-in viewer on Film A's player.
- **Action / input:** Press fullscreen, then exit fullscreen.
- **Expected result:** The video stage expands to fill the screen, and the film keeps its playing or paused state. Exiting restores the normal page layout and playback state, with no page reload and no loss of the current position.

#### TC-055 — Playback starts before the whole file has downloaded
- **Type:** positive
- **Spec:** REQ-010
- **Level:** end-to-end
- **Priority:** P1
- **Automation candidate:** yes
- **Preconditions / test data:** Signed-in viewer; network throttled below the sample file's bitrate; Film A about 10 minutes long.
- **Action / input:** Open Film A's player and start playback; observe the transferred byte count and the timecode.
- **Expected result:** Playback begins within a few seconds while the transferred byte count is still far below the file's full size (the film has not finished downloading). The player shows progress rather than waiting for a complete download.

#### TC-056 — Every seeded film can be selected and played
- **Type:** positive
- **Spec:** REQ-013
- **Level:** end-to-end
- **Priority:** P1
- **Automation candidate:** yes
- **Preconditions / test data:** Signed-in viewer; all 20 seeded films.
- **Action / input:** For each of the 20 films: open it from the catalogue and play it for at least 5 seconds.
- **Expected result:** All 20 films open their own player with their own title and all 20 begin playing. No film shows a missing, broken or endless-loading video, and no card leads to the wrong film.

### Player errors and access control

#### TC-057 — An address for a film that does not exist shows "film not found"
- **Type:** negative
- **Spec:** REQ-018
- **Level:** UI
- **Priority:** P1
- **Automation candidate:** yes
- **Preconditions / test data:** Player address naming an id that is not in the catalogue (`/watch/does-not-exist`), opened signed in and signed out.
- **Action / input:** Open the address in both session states.
- **Expected result:** A clear "Film not found" page is shown (the approved design's wording) with a brief explanation and a "Back to the catalogue" action that returns to the catalogue. The page is never blank, never an error trace, and never an endless spinner.

#### TC-058 — An unknown film at the API returns a clean not-found
- **Type:** negative
- **Spec:** REQ-018
- **Level:** API
- **Priority:** P2
- **Automation candidate:** yes
- **Preconditions / test data:** Seeded catalogue; unknown film id.
- **Action / input:** Request the data for the unknown film id.
- **Expected result:** HTTP 404 with a clean, machine-readable error body. There is no HTTP 500, no stack trace, no internal path or database detail, and no fabricated film record.

#### TC-059 — An unplayable video shows an error with retry, never an endless spinner
- **Type:** negative
- **Spec:** REQ-019
- **Level:** end-to-end
- **Priority:** P1
- **Automation candidate:** yes
- **Preconditions / test data:** Signed-in viewer; the selected film's video URL blocked at the proxy.
- **Action / input:** Open the film's player and wait 10 seconds.
- **Expected result:** Within a bounded time the player shows a clear error stating the film cannot be played, with a "Try again" action. It does not spin indefinitely, does not present a black silent stage as if playable, and does not show an error trace.

#### TC-060 — Retry after a video failure restores playback
- **Type:** positive
- **Spec:** REQ-019
- **Level:** end-to-end
- **Priority:** P1
- **Automation candidate:** yes
- **Preconditions / test data:** As TC-059, video URL blocked and the error shown; the proxy rule is then removed.
- **Action / input:** Use "Try again" without reloading the page or returning to the catalogue.
- **Expected result:** The error clears, the film loads and plays from the start (or the same player view), and the viewer does not lose their place in the app.

#### TC-061 — A visitor cannot fetch video without a session
- **Type:** negative
- **Spec:** REQ-011
- **Level:** API
- **Priority:** P1
- **Automation candidate:** yes
- **Preconditions / test data:** No session; a known film's video reference taken from the catalogue data.
- **Action / input:** Request the video file directly (ordinary GET and a ranged request).
- **Expected result:** Both requests are refused (HTTP 401 or 403) and return no video bytes. Catalogue metadata being public (TC-003) does not make the video public.

### Seed data

#### TC-062 — Every seeded film carries the required fields
- **Type:** positive
- **Spec:** REQ-029
- **Level:** API
- **Priority:** P2
- **Automation candidate:** yes
- **Preconditions / test data:** The seeded 20-film catalogue and its media files.
- **Action / input:** For each film, read its title, short description, poster reference and video reference; request each poster and each video reference.
- **Expected result:** All 20 films have a non-empty title, a non-empty short description, a working poster image and a working video file. No film is missing a field or points at a 404 media address.

#### TC-063 — No seeded film exceeds 30 minutes
- **Type:** boundary
- **Spec:** REQ-029
- **Level:** API
- **Priority:** P3
- **Automation candidate:** yes
- **Preconditions / test data:** The seeded media files; a probe that reads each video's duration; the environment can also attempt to seed a 30:01 film.
- **Action / input:** Measure every seeded film's duration, then attempt to seed a film of 30:00 and one of 30:01.
- **Expected result:** Every seeded film measures 30:00 or less. A 30:00 film is valid seed data. A 30:01 film is outside the approved seed rule: it is not accepted into the v1 catalogue (or, if accepted, that is a defect against REQ-029) and never appears at 30:01 in the catalogue.

## Open questions for /document-preparation

None. Every expected result above is fixed by the merged `spec.md` v1, or by
the merged design where the spec leaves copy or route wording open
(`AppFlow.md`, `UIMap.md`). The spec does not prescribe the HTTP status of a
refused request or an exact error body, so TC-020, TC-021, TC-061 and TC-058
accept the ordinary refusal codes (401/403 for a refused session, 404 for an
unknown film) without inventing a house error-code catalogue.

## Approval

The business owner reviewed this test design in plain language on 2026-11-14
and approved the coverage as drafted ("okay go with it"), asking for it to be
finalized and its pull request opened for review. No case was changed after
that approval.

## Execution notes

- None of these cases has been executed; this document defines them for later
  execution by `/go-dev` and QC.
- TC-004, TC-005, TC-057, TC-059 and TC-063 need the environment controls
  listed under "Scope and test data"; they do not require any change to the
  approved catalogue.
- Browser-restart (TC-018) needs a persistent browser profile; the other UI
  cases can run in a clean profile per case.
