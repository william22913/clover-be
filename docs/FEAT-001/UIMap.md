---
feature_code: FEAT-001
service: clover
version: 2
status: in-review
spec_version: 1
design_source: reference-video (look) + spec.md v1
reference: https://cdn.dribbble.com/userupload/44295925/file/9d7c0609260c04d650c6bc1d5d8528c4.mp4
completeness:
  scope_confirmed: covered
  requirement_mapping: covered
  page_inventory: covered
  navigation: covered
  content_and_fields: covered
  states: covered
  interactions: covered
  motion: covered
  responsive: covered
  design_tokens: covered
  components: covered
  assets: covered
open_questions: []
derived_confirmed: 2026-11-13, the manager's instruction for this pass authorizes every derived decision listed below and fixes the v1 scope at catalogue, sign-in / sign-up and player
pr_history:
  - number: 58
    url: https://github.com/william22913/sandbox-docs/pull/58
    title: "[UI] clover/FEAT-001: v1 catalogue, account and player design"
    state: open
pr_url: https://github.com/william22913/sandbox-docs/pull/58
last_updated: 2026-11-13
---

# FEAT-001 — UI map (clover)

First UI pass for Clover v1, rebuilt after the previous draft on `main` was
rejected for designing screens the spec does not have (search, genres, My list,
a hero carousel). The delivered design covers exactly three things: the public
catalogue, the combined sign-in / quick sign-up screen, and the player. This map
and the mock replace that draft.

## Design source and provenance

| Source | What it fixes |
| --- | --- |
| The manager's reference video (8.0 s, 1600 × 1200, 60 fps, Dribbble) | **The look only.** Warm amber light over near-black, a full-viewport hero with a spread top nav and a white gooey capsule, a huge white all-caps display headline at roughly 92 % of the column, one full-bleed photograph per screen, frosted rails carrying real entries, a white stat card, three glass feature cards, and an entrance in the 180–600 ms range. |
| `spec.md` v1 (24 REQ, 12 SC) | Every screen, field, state, error message and validation rule. |
| `AppFlow.md` | The four journeys behind the routes: browse, sign up and watch, sign in and watch, abandoned sign-in. |
| `testcase.md` | The state and behaviour inventory this map is checked against (TC-001 … TC-045). |
| Manager's instruction, 2026-11-13 | v1 scope only; no search, genres, My list, saved progress, ratings or subscriptions; English copy; feasible for a Go standard-library server with static HTML, CSS and JS and no nexcommon or nexcomponent. |

Provenance tags in the Source columns: `derived` means read off the reference
video or derived from the spec plus the reference's patterns; `designer` means a
design decision made under the manager's instruction in this pass; `assumed`
marks something the mock can only suggest because it belongs to the server;
`figma` is unused — there is no Figma file.

**The reference's subject is not Clover's.** The recording is a microbiology
landing page inside a 3D laptop mockup: none of its words, images or sections
are reproduced, and it has no film catalogue, account screen or player. The
manager confirmed the look is what must be matched; every screen here is
Clover's own, built in that language. The decision is recorded in
`DesignBrief.md`.

## Requirement mapping

| REQ | Requirement | Page(s) | Mock screen | Source | Notes |
| --- | --- | --- | --- | --- | --- |
| REQ-001 | A visitor sees the catalogue immediately, no login prompt | Catalogue | `catalogue` | derived | Catalogue is the first page; no account prompt anywhere on it |
| REQ-002 | Every published movie, no pagination, search, filtering or sorting | Catalogue | `catalogue` | derived | One grid of all 13 sample films, no controls above it |
| REQ-003 | Select any movie and open its player page | Catalogue, Player | `catalogue`, `player` | derived | The whole card is one link (`#/play/<id>`) |
| REQ-004 | Signed-out selection goes to a combined login / sign-up screen, then straight back to that movie | Catalogue, Sign in / Create account, Player | `auth-signin`, `auth-signup`, `player` | derived | The selected film stays on the auth screen; sign-in continues to `/play/<id>` |
| REQ-005 | Self-service account with email and password, no verification | Sign in / Create account | `auth-signup` | derived | "Create account" tab of the same form; usable immediately |
| REQ-006 | Returning user signs in with the registered credentials | Sign in / Create account | `auth-signin` | derived | "Sign in" tab is the default |
| REQ-007 | A signed-in user stays signed in between visits | Catalogue, Player | `catalogue-signed-in`, `player` | assumed | 7-day HttpOnly cookie is server work; the mock shows the signed-in header |
| REQ-008 | Player offers play, pause, timeline scrub, volume and fullscreen | Player | `player`, `player-playing` | derived | Custom bar, not native controls; all five work |
| REQ-009 | Playback is progressive and seeking resumes from the new point | Player | `player-playing` | assumed | The mock streams a real local MP4, so seek works; byte-range serving is server work |
| REQ-010 | Sign out returns to anonymous browsing | Catalogue | `catalogue-signed-in` | derived | "Sign out" in the header; movies lock again |
| REQ-011 | The entire interface is in English | Every page | all | derived | Every label, message and note is English |
| REQ-012 | A player URL opened while signed out behaves like REQ-004 | Player, Sign in / Create account | `auth-signin`, `player` | derived | Signed-out `/play/:id` redirects to the auth screen with `next` |
| REQ-013 | Unknown email or wrong password: one error, no disclosure, stays on the form | Sign in / Create account | `auth-signin-error` | derived | "That email and password do not match an account." Input kept |
| REQ-014 | Sign-up with an existing email is refused, no second account, no sign-in | Sign in / Create account | `auth-signup-duplicate` | derived | Field-level message with "Sign in instead." |
| REQ-015 | Invalid input is rejected inline, explaining what to fix, without losing input | Sign in / Create account | `auth-signup-invalid` | derived | Field message, red border, focus moved to the first bad field |
| REQ-016 | Abandoning the login screen leaves the visitor anonymous and browsing | Sign in / Create account, Catalogue | `auth-signin` | derived | "Back to the catalogue" in the header; the film stays locked |
| REQ-017 | A player URL for a missing movie shows a clear message and a way back | Player | `player-missing` | derived | "This film is not available." plus back link |
| REQ-018 | A video that fails shows a clear error with retry, never a spinner | Player | `player-error` | derived | "The film could not be loaded." + "Try again" + back |
| REQ-019 | An expired session sends the user to sign in and back to the same movie | Sign in / Create account | `auth-expired` | derived | Neutral notice: "Your session ended. Sign in again to keep watching." |
| REQ-020 | Sign-out is complete server-side; the old cookie is refused | — | — | assumed | No screen; the mock's sign-out clears its in-memory session |
| REQ-021 | One account may be signed in on several devices | — | — | assumed | No screen; server behaviour |
| REQ-022 | Email required, valid, unique case-insensitively | Sign in / Create account | `auth-signup-invalid`, `auth-signup-duplicate` | derived | Client checks format; the case-insensitive duplicate is server work |
| REQ-023 | Sign-in matches the registered email case-insensitively and the exact password | Sign in / Create account | `auth-signin`, `auth-signin-error` | derived | No trimming or case-folding of the password, in the mock either |
| REQ-024 | Password at least 8 characters with a letter and a number | Sign in / Create account | `auth-signup`, `auth-signup-invalid` | derived | Hint under the field when signing up; three distinct messages |

### Success criteria coverage

| SC | Shown by | What the mock demonstrates |
| --- | --- | --- |
| SC-001 | `catalogue` | All 13 films, poster + title + description, no login prompt |
| SC-002 | `auth-signup` → `player` | Sign-up tab, submit, lands on `/play/halcyon` as a signed-in viewer |
| SC-003 | `auth-signin` → `player` | Sign-in tab, submit, same destination |
| SC-004 | `player`, `player-playing` | A real 14 s MP4: play, pause, scrub, volume, fullscreen all act on it |
| SC-005 | `auth-signin-error` | One generic error, form still filled, not signed in |
| SC-006 | `auth-signup-duplicate` | Refused with a route to sign in instead |
| SC-007 | `auth-signin` | Back to the catalogue; `#/play/halcyon` still redirects to auth |
| SC-008 | `catalogue-signed-in` | Signed-in header; cookie persistence is server work |
| SC-009 | — | Server-only; no screen can show it |
| SC-010 | `player-missing`, `player-error` | Both messages carry a recovery action, no spinner |
| SC-011 | `auth-signup-invalid` | Length, letter and number rules each have their own message |
| SC-012 | `auth-signup-duplicate` | Duplicate treatment exists; case-insensitive comparison is server work |

## Pages

Every page is reachable in the mock at `ui/mock/index.html#<route>`, and every
state in `ui/mock/screens.json` also opens directly at
`ui/mock/index.html#<id>`. That id is the preview contract: the viewer sets the
hash to the screen id, so the mock routes it to the state it names. States
without an id are listed under **Not mocked**.

The viewer builds its page list from the `### ` headings here, and joins a mock
screen to a page only when `page` in `ui/mock/screens.json` equals that heading
after normalising case and punctuation. Both sides must say the same thing.

### Catalogue

**Purpose:** Show the whole catalogue on the first page load, with nothing
standing between the visitor and the films. Answer "what is on Clover?" before
asking anything. REQ-001 and TC-001 require that no account prompt is in the way.

**Mock screens.** `catalogue` (signed out), `catalogue-signed-in`.

**Content and fields**

| Element | Type | Required | Content and rule | Source |
| --- | --- | --- | --- | --- |
| Brand | link, top left | yes | Clover leaf mark in amber + "Clover" in Inter Tight 700, links home | designer |
| Header links | two text buttons | yes | "Catalogue" and "How it works", 10 px tracked caps, spread across the middle of the header; each scrolls in place | designer |
| Header action, signed out | gooey capsule | yes | White "Sign in" capsule with the arrow knob at the right end of the header | designer |
| Header action, signed in | avatar + text button | yes | White initial disc of the account email + "Sign out" | designer |
| Hero still | image, decorative | yes | One full-bleed photograph under the whole first screen, warm grade, `alt=""`, drifting 1.02 → 1.09 over 34 s | designer |
| Headline | display, one line | yes | "Pick a film." — Inter Tight 700, uppercase, ~92 % of the column | designer |
| Lede | body, max 36 ch | yes | "The whole catalogue is open. Watching takes an account, and Clover only asks when you press play." | designer |
| Hero action | frosted icon button | yes | 46 px rounded glass square with a down arrow, `aria-label="See all 13 films"`, scrolls to the grid | designer |
| Facts rail | frosted pill, right edge | yes | "13 films" and "2021–2025" in glass chips plus a white down button; bleeds off the right edge of the screen | designer |
| Film rail | frosted panel, two entries | yes | Two catalogue films: 56 px poster thumb, year chip, uppercase title, short caption and a white arrow disc; the whole entry opens that film | designer |
| Stat card | white card | yes | "13" / "films in the catalogue, free to browse" and "1" / "account is all watching takes" | designer |
| Step cards | three glass cards | yes | Browse / Sign in / Watch, each with a one-line caption; the row is the "How it works" scroll target | designer |
| Section heading | h2 | yes | "The catalogue" | designer |
| Count | caption | yes | "13 films, free to browse" — the count comes from the film list | derived |
| Film card | link | yes | 2:3 poster, uppercase title, 3-line description; the whole card is one link | derived |

The mock ships 13 placeholder films (`FILMS` in `app.js`) with invented titles,
years, runtimes and one-sentence descriptions. The grid cards show poster, title
and short description only, exactly as REQ-001 and TC-003 require; the rail
entries add a year chip so the two entries read as a pair. Year and runtime also
appear on the auth art and in the player meta.

**States**

| State | Mock id | What it looks like |
| --- | --- | --- |
| Signed out | `catalogue` | The full hero with the gooey Sign in capsule; both rail entries and all 13 cards are selectable |
| Signed in | `catalogue-signed-in` | Initial disc + "Sign out"; the catalogue is unchanged |
| Card hover / focus | interaction | Poster lifts 2 px, gets a warm shadow and scales its image 1.035; the title stays put |
| Rail entry hover | interaction | The glass entry brightens and its white arrow disc shifts 2 px up-right |
| Header scrolled | interaction | After 16 px the fixed header turns to 82 % near-black with a 20 px blur and a hairline |
| Entrance | interaction | See Motion |

**Interactions**

| Action | Result |
| --- | --- |
| Select a card | Signed out → `#/auth?next=<id>`; signed in → `#/play/<id>` |
| Select a rail entry | Same rule as a card, for that entry's film |
| "Catalogue" / down button | Smooth-scrolls in place to the grid (`#the-catalogue`) and moves focus there; the header stays fixed |
| "How it works" | Smooth-scrolls in place to the step cards (`#how-it-works`) |
| "Sign in" | `#/auth` with no film context; signing in returns to the catalogue |
| "Sign out" | Session cleared, header returns to "Sign in", cards lock again |
| Tab through the page | Cards receive the 2 px gold focus ring; the skip link jumps to `#main` |

### Sign in / Create account

**Purpose:** One screen for both account actions, reached either from a film or
from the header. A visitor who came from a film must be able to finish and land
on that film; a visitor who came from the header returns to the catalogue. The
selected film is shown, so the screen answers "what am I signing in for?".

**Mock screens.** `auth-signin`, `auth-signin-error`, `auth-signup`,
`auth-signup-invalid`, `auth-signup-duplicate`, `auth-expired`.

**Content and fields**

| Element | Type | Required | Content and rule | Source |
| --- | --- | --- | --- | --- |
| Brand | link, top left | yes | Same as the catalogue | designer |
| Back link | link, top right | yes | "Back to the catalogue" with a chevron; the only way out | derived |
| Film still | image, decorative | yes | The selected film's still, graded and full-bleed behind the whole screen; the quiet still when no film | designer |
| Film title | display | when a film is selected | The film's own title, uppercase, bottom-left over the art | derived |
| Film description | body, max 42 ch | when a film is selected | The same one-line description as its card | derived |
| Film meta | caption | when a film is selected | Year · runtime, tracked caps, e.g. "2023 · 2h 06m" | derived |
| Form panel | frosted panel, right | yes | 432 px max, radius 26, white gradient glass over the art | designer |
| Clover block | display + body + caption | when no film | "Clover" / "A small cinema. Every film in the catalogue, open to browse." / "Sign in when you want to watch" | designer |
| Heading | h1 | yes | Sign in with a film: "Sign in to watch"; without: "Sign in"; sign-up: "Create an account" | designer |
| Sub-line | body | yes | "Use your account and the film starts." / "Welcome back. The catalogue is waiting." / "One email and a password, then the film starts." / "One email and a password. That is the whole form." | designer |
| Mode tabs | two toggle buttons | yes | "Sign in" and "Create account", `aria-pressed`, gold underline on the active one | derived |
| Notice | status / alert | no | Credentials error, duplicate-email hint or expired-session notice, above the fields | derived |
| Email | input, type=email | yes | `autocomplete="email"`, placeholder "you@example.com", value kept after any error | derived |
| Password | input, type=password | yes | `new-password` when creating, `current-password` when signing in; never trimmed | derived |
| Password hint | helper line | sign-up only | "At least 8 characters, including a letter and a number." | derived |
| Field error | inline message | on error | One line under the field, red; the field border turns red | derived |
| Submit | gooey capsule, full width | yes | "Sign in" or "Create account" with an arrow knob inside the pill | designer |

**States**

| State | Mock id | What it looks like |
| --- | --- | --- |
| Sign in, from a film | `auth-signin` | Film still + title, description, meta; heading "Sign in to watch" |
| Sign in, wrong credentials | `auth-signin-error` | Red notice above the fields: "That email and password do not match an account." Email kept, password cleared |
| Create account, from a film | `auth-signup` | Same layout, heading "Create an account", password hint visible |
| Create account, invalid input | `auth-signup-invalid` | Email marked with a format message and the password with "Use at least 8 characters."; typed values kept |
| Create account, email taken | `auth-signup-duplicate` | Email field message "An account already uses this email. Sign in instead." — the link switches the tab |
| Expired session | `auth-expired` | Gold notice: "Your session ended. Sign in again to keep watching." |
| No film context | interaction | Header "Sign in" leads here: the Clover block replaces the film block |
| Client-side validation | interaction | Empty field, bad format, short password, no letter, no number each get their own message; focus moves to the first invalid field |

**Interactions**

| Action | Result |
| --- | --- |
| Switch tab | Copy, password hint, autocomplete and submit label change in place; typed values stay; errors clear |
| Submit, invalid | Stays on the screen, inline messages appear, first invalid field takes focus |
| Submit, valid, film context | Signs in and goes to `#/play/<id>` |
| Submit, valid, no film | Signs in and goes to `#/` |
| "Sign in instead." | Switches to the sign-in tab with the email kept |
| "Back to the catalogue" | `#/` as an anonymous visitor |

### Player

**Purpose:** Play the selected film with the five controls the spec names, and
give every failure a way out. Full-viewport and quiet: the film is the page.

**Mock screens.** `player`, `player-playing`, `player-idle`, `player-error`,
`player-missing`.

**Content and fields**

| Element | Type | Required | Content and rule | Source |
| --- | --- | --- | --- | --- |
| Video | `<video>`, `playsinline`, `preload=metadata` | yes | The film's still as poster; a real MP4 source; native controls off, custom bar on | derived |
| Back | icon button, top left | yes | 44 px frosted disc with a chevron, `aria-label="Back to the catalogue"` | derived |
| Title | h1 | yes | The film title, uppercase 17 px, over a top scrim | derived |
| Meta | caption | yes | Year · runtime under the title | derived |
| Play / pause | icon buttons | yes | 88 px white disc centred while paused; 40 px bar button always | derived |
| Elapsed / duration | times | yes | `m:ss`, tabular figures; duration follows the loaded media | derived |
| Scrubber | range input | yes | Full-width track, gold fill, white knob; 5 % of a 14 s clip per arrow step | derived |
| Volume | button + range | yes | Mute toggle and a 78 px slider; slider hides under 600 px, mute stays | derived |
| Fullscreen | icon button | yes | Enters on the player container, icon and label switch to "Exit fullscreen" | derived |
| Message panel | h1 + body + actions | on failure | One sentence, one explanation, buttons underneath | derived |

**States**

| State | Mock id | What it looks like |
| --- | --- | --- |
| Paused, opened | `player` | First frame, big play disc, chrome visible |
| Playing | `player-playing` | Chrome visible while playing |
| Playing, idle | `player-idle` | Chrome and cursor gone; any pointer move brings them back |
| Video failed | `player-error` | "The film could not be loaded." / "Check your connection, then try again." + "Try again" + "Back to the catalogue" |
| Film missing | `player-missing` | "This film is not available." / "It may have left the catalogue, or the link may be wrong." + "Back to the catalogue" |
| Ended | interaction | Returns to the paused state with the play disc, chrome visible |

**Interactions**

| Action | Result |
| --- | --- |
| Play / pause | Click the big disc, the bar button, the video itself, or Space / K |
| Seek | Drag or arrow the scrubber, or ← / → (± 5 s); playback continues from that point |
| Volume | Slider, mute button, or ↑ / ↓ in 10 % steps; M mutes |
| Fullscreen | Button or F; Escape exits |
| Idle | While playing, chrome fades after 2.6 s; it stays while paused |
| Retry | Reloads the player for the same film |

Keyboard shortcuts are a house floor, not a spec requirement: Space / K, ← / →,
↑ / ↓, M, F. They are ignored while focus is in an input, button or link.

## Navigation

| From | Action | To | Note |
| --- | --- | --- | --- |
| Catalogue | Film card | Player (signed in) or Sign in (signed out) | The film id travels in the route |
| Catalogue | "Sign in" | Sign in, no film | Returns to the catalogue after success |
| Catalogue | "Sign out" | Catalogue | Session cleared |
| Sign in | "Back to the catalogue" | Catalogue | Anonymous |
| Sign in | Valid submit | Player of the film, or Catalogue | `next` is honoured |
| Sign in | "Sign in instead." | Sign in, same film | Mode switch only |
| Player | Back | Catalogue | The film keeps no progress (out of scope) |
| Player (signed out direct link) | — | Sign in with `next` | Redirect, not an error |
| Unknown route | — | Catalogue | One silent fallback |

Browser back and forward work, because every transition is a hash change. On
each route change the document title changes and `#main` takes focus, so
keyboard and screen-reader users are told the page changed.

## Motion

| Where | What | Duration | Easing | Source |
| --- | --- | --- | --- | --- |
| Catalogue hero still | Fades in over 700 ms, then drifts 1.02 → 1.09 over 34 s, alternating | 700 ms / 34 s | ease / ease-in-out | derived |
| Catalogue headline | Rises from an overflow mask, 105 % → 0 | 700 ms | `cubic-bezier(.16,1,.3,1)` | derived |
| Catalogue lede and down button | Fades up 12 px, starts at 140 ms | 500 ms | same | derived |
| Facts rail | Fades up, starts at 300 ms | 500 ms | same | designer |
| Film rail | Fades up, starts at 220 ms | 550 ms | same | designer |
| Stat card and step cards | Fade up, starts at 340 ms; steps stagger 60 ms | 550–600 ms | same | designer |
| Rail entry hover | Glass brightens, arrow disc shifts 2 px up-right | 200 ms | same | designer |
| Catalogue cards | Fade up 12 px, 28 ms stagger, capped at 6 | 400 ms | same | designer |
| Card hover | Poster lifts 2 px, warm shadow, image scales 1.035 | 250 ms / 500 ms | same | designer |
| Header | Turns solid on scroll | 250 ms | default | designer |
| Auth still | Fades in | 500 ms | ease | derived |
| Auth film block | Fades up, 80 ms delay | 500 ms | reference ease | derived |
| Auth form box | Children fade up, 40 ms stagger | 450 ms | same | designer |
| Player chrome | Fades out when idle, in on move | 250 ms | same | designer |
| Player play disc | Hover scale 1.05 | 220 ms | same | designer |
| Goo knob | Arrow shifts 2 px on hover | 200 ms | same | designer |
| Focus ring | Appears with the control, no animation | — | — | house floor |

Everything is turned off under `prefers-reduced-motion: reduce` — including the
hero drift — with the entrance elements forced to their end state (never stuck
invisible).

## Responsive

| Width | Catalogue | Sign in / Create account | Player |
| --- | --- | --- | --- |
| ≥ 1200 | Hero fills the viewport, capped at 980 px; 5 columns, 26 px gap, 32 px gutter | Full-bleed still, 432 px frosted form panel at the right | Full-viewport, 20 px chrome inset |
| 900–1199 | Facts rail hidden from 1099, hero foot 2 columns; 4 columns, 22 px gap | Same, art foot capped at 58 % from 1099 | Same |
| 600–899 | Rail stacks to two rows, stat card above the steps, hero min 620 px; 3 columns, 24 px gutter | Stacked from 959: 250 px still band, form below, left-aligned, panel dissolved, box ≤ 460 px | Same |
| < 600 | Header links hidden, rail thumb hidden, steps become one compact row each; 2 columns, 16 px gap, 20 px gutter | 210 px still band, art description and meta hidden, 52 px submit | Volume slider hidden, mute kept; 58 px bar, 72 px play disc |

The headline scales with `clamp(46px, 16.2vw, 270px)` and is verified to fit one
line at 1440, 768 and 390. Type scales down in three steps: section heading
38 → 26 → 22 px, card title 12.5 px, auth heading 40 → 26 px.

## Design tokens

Declared once in `:root` in `app.css`; nothing outside that block introduces a
colour or radius.

| Token | Value | Use |
| --- | --- | --- |
| `--ink-000` | `#050505` | Page, player letterbox, the bottom of every scrim |
| `--ink-100` | `#0b0b0c` | Poster placeholder background |
| `--ink-200` | `#141210` | Input background (warm near-black) |
| `--paper` | `#ffffff` | Display type, goo capsule, play disc, avatar |
| `--ink-on-paper` | `#111111` | Text and icons on white |
| `--paper-muted` / `--paper-rule` | `#6e6e6e` / `#e6e4e1` | Reserved for the printed-language surfaces; not used in v1 |
| `--gold` | `#e0a34e` | Brand mark, active tab, scrub fill, focus-adjacent emphasis |
| `--gold-deep` | `#8b4d11` | Warm grade, deep end |
| `--gold-glow` | `#ffc46b` | Focus ring, hover on gold text |
| `--text` / `--text-dim` / `--text-strong` | 78 % / 56 % / 100 % white | Body, captions, headings |
| `--glass` / `--glass-2` | 9 % / 5 % white | Frosted fills: facts rail, film rail, step cards, hero icon button, player chrome |
| `--glass-edge` / `--glass-edge-strong` | 16 % / 45 % white | Hairlines on inputs and glass |
| `--danger` / `--danger-line` | `#ff9b8b` / `#e5604d` | Error text / error border |
| `--gutter` | 32 / 24 / 20 px | Page gutter by breakpoint |
| `--nav` | 72 / 64 px | Header height |
| `--r-panel` / `--r-card` / `--r-poster` / `--r-thumb` / `--r-input` | 26 / 22 / 20 / 14 / 14 px | Radii, largest surface to smallest |
| `--bleed` | gutter + half the space outside `--wrap` | Lets the facts rail and film rail run off the right edge |
| `--wrap` | 1680 px | Maximum content width |
| `--ease` | `cubic-bezier(.16,1,.3,1)` | Every entrance and hover |
| `--sans` / `--tght` | Inter / Inter Tight | Interface / display |

The warm grade is a fixed recipe, not a filter knob per image: image
`brightness(.74) saturate(.94) contrast(1.05)`, then a screen-blended amber
gradient `rgba(214,124,32,.62)` → `rgba(168,86,8,.28)` at 40 % →
`rgba(120,60,10,.08)` at 72 % → transparent, then three black scrims (top, left,
bottom). Any replacement still at a similar exposure grades the same way.

## Components

| Component | States | Notes |
| --- | --- | --- |
| Header | over still, solid after scroll | Fixed; brand left, two spread links centre, capsule or account disc right |
| Gooey capsule | default, hover, active, focus, full width | White pill whose right knob is fused by an SVG goo filter; arrow shifts on hover |
| Facts rail | default | Frosted pill bleeding off the right edge: factual chips plus a white down button |
| Film rail | default, entry hover | Frosted two-entry panel: poster thumb, year chip, uppercase title, caption, white arrow disc |
| Stat card | default | White card, two counted facts, hairline between them |
| Step card | default | Glass card with a circled icon, a tracked-caps label and a one-line caption |
| Poster card | default, hover, focus | One link; poster, uppercase title, three-line description |
| Mode tabs | active, inactive, hover, focus | `aria-pressed`, gold 2 px underline |
| Text field | default, hover, focus, invalid | 54 px, warm near-black, gold focus ring, red when invalid |
| Inline message | error, hint | 12.5 px, sits under its field and is tied to it visually |
| Notice | alert (error), status (info) | Rounded panel above the form |
| Goo link | default, hover | Same capsule as an anchor, used for "Try again" |
| Player chrome | visible, idle | Top scrim + bottom frosted bar, fade together |
| Big play disc | paused, hover | 88 px white disc; hidden while playing |
| Range control | default, focus | Gold fill via `--fill`, 4 px track, 12 px white thumb |
| Message panel | missing, failed | Centred, one heading, one paragraph, one or two actions |

## Assets

| Asset | Source | Licence / note |
| --- | --- | --- |
| `assets/fonts/inter-400 / 500 / 600`, `inter-tight-700` (woff2) | Inter and Inter Tight | SIL Open Font License; self-hosted, no external request |
| 13 posters, 2:3 | Unsplash placeholders | Placeholder only; replace with Clover's licensed posters |
| 3 stills (16:9): `still-halcyon`, `still-quiet`, `still-vesper` | Unsplash placeholders | Reused across films in the mock; each film gets its own still in production |
| `assets/video/clover-sample.mp4` | Generated for this mock from `still-halcyon` with ffmpeg | 14 s, 1280 × 720, h264 + AAC, 699 KB; every film plays it so the controls can be exercised. Not Clover content |
| Clover mark | Hand-drawn SVG, 24 px grid | Inline `<symbol>`, stroke in `currentColor` |

There are no external requests of any kind: fonts, images, video and icons are
all local files.

## Derived decisions

1. **The reference supplies the look, not the product.** Its subject is a
   microbiology site; nothing of its content is reused.
2. **One photograph per page**: the catalogue hero, the auth art panel and
   the player frame. No page has two competing images.
3. **The catalogue opens on a full-viewport hero, as the reference's screen
   does.** The manager rejected the first v1 draft's 42 vh band as not
   following the reference closely enough; the hero now fills the viewport and
   is capped at 980 px so it stays a composition on very tall displays. SC-001
   is still met in the sense that matters above the fold: the two-entry glass
   rail carries real films — poster, title, description — and every entry opens
   that film, so a visitor can pick a film without scrolling. The full 13-film
   grid sits directly under the hero, reached by the hero's down button, the
   facts rail and the "Catalogue" header link.

   The reference's rails, stat card and feature cards are rebuilt on Clover's
   own facts: the frosted right-edge rail carries "13 films" and "2021–2025",
   not social icons; the white card counts 13 films and one account; the three
   glass cards are Browse / Sign in / Watch, the manager's own three steps. No
   copy, number, icon or section is taken from the video. The still drifts
   1.02 → 1.09 over 34 s so the first screen breathes the way the reference's
   recording does; the drift is off under reduced motion.
4. **The headline is the product's own instruction**: "Pick a film." — the
   manager's three-step journey in two words. The lede states the account rule
   once, so the catalogue itself never has to prompt.
5. **Cards show poster, title and description only.** Year, runtime and genre
   are absent because the spec asks for exactly three fields and test case
   TC-003 forbids extra controls. Year and runtime do appear on the auth art
   and the player, where they help identify the film being unlocked.
6. **One screen, two modes.** Sign-in and sign-up share a form and switch by
   tab, which is what "combined login / sign-up screen" means and what TC-006
   checks.
7. **The selected film stays visible while signing in.** The full-bleed still
   behind the form is the film the visitor came for, with its title, description
   and year · runtime bottom-left; without a film it shows the Clover block.
   After success the visitor continues to that film, so the screen is a promise
   the flow keeps.
8. **The gooey capsule is the only heavy shape.** It carries one primary action
   per screen — the header's "Sign in", the auth submit, the player's "Try
   again" — with the arrow knob keeping the reference's organic pill as the
   single memorable element.
9. **Errors are three different things.** A credential failure is an alert
   above the form; a field rule is a message under the field; an expired session
   is a neutral gold notice. They never share a style, so the visitor can tell
   "you typed it wrong" from "your session ended" at a glance.
10. **The player bar is frosted glass over the film** and the top bar is a
    scrim, both following the reference's material language. Chrome auto-hides
    only while playing and never while paused.
11. **Keyboard control is a house floor**, not a spec item: Space / K, ← / →,
    ↑ / ↓, M, F, with the player claiming those keys only while it is mounted.
12. **The mock's player is real enough to test.** A 14 s local MP4 makes play,
    pause, seek, volume and fullscreen genuinely observable, which is what
    SC-004 and TC-010 ask a reviewer to check. Where media cannot load — inside
    `/ui-preview`, whose inner CSP allows none — the poster still and an
    in-page clock take over, so the same controls stay observable; opening the
    mock directly still plays the file.
13. **Screens open by id.** Each `screens.json` id is a hash the mock routes to
    that state (`#auth-signup-invalid`, `#player-idle`, …), because
    `/ui-preview` opens a screen by setting the iframe hash to the id. The page
    list comes from the `### ` headings in this file and the `page` field in
    `screens.json`; the two must match.
14. **Preview flags are not product features.** `?u=1`, `?demo=1`, `?playing=1`,
    `?idle=1`, `?error=video`, and the `err=bad|invalid|dup|expired` and
    `email=` parameters exist so `/ui-preview` can open a state directly. The Go
    build has none of them.
15. **The mock's session lives in memory.** Any valid form signs in; no cookie,
    no server. The 7-day session, the hashed passwords and the server-side
    sign-out are described under Not mocked.
16. **No modal, no focus trap.** Auth is a page, not a dialog, so the focus
    question the previous draft had does not arise. Route changes move focus to
    `#main`.
17. **The catalogue fetch has no error state in the mock** because the data is
    a constant. The Go build serves the same shell and must handle a failed
    catalogue request; that is implementation, not design.

## Not mocked

| Item | Why | Where it belongs |
| --- | --- | --- |
| Real accounts, password hashing, email uniqueness and case-folding | The mock accepts any valid form | Backend (REQ-022, REQ-023, SC-006, SC-012) |
| 7-day session cookie, persistence across browser restarts, server-side sign-out, multi-device | No server in the mock | Backend (REQ-007, REQ-020, REQ-021, SC-008, SC-009) |
| Byte-range video serving and progressive playback over the network | The mock plays a local file; `/ui-preview` blocks media, so there the still and a clock stand in and the controls still work | Backend (`http.ServeContent`, REQ-009) |
| A genuine network failure in the player | The error state is opened by a preview flag | Both: server error handling, plus the same panel |
| A failed or empty catalogue fetch | The mock's catalogue is a constant | Backend plus the same card grid |
| Resume position, watch history, ratings, subscriptions | Explicitly out of scope | Later pass |
| Real key art and a real film | Placeholders by design | Content |
| Captions and audio-description tracks | Not in v1 | Later pass |

## Open questions

None. The reference's subject differs from Clover's, and the manager has
confirmed that its look and mood are what must be matched; the scope of v1 is
fixed by their instruction and by `spec.md`.

## Handoff to the Go build

`/code-preparation` and `/go-dev` should be able to build this with the standard
library only:

- One `index.html` shell plus `app.css` and `app.js`, served as static files.
  The three routes (`/`, `/auth`, `/play/:id`) may be hash routes exactly as in
  the mock, or server paths that return the same shell — the design does not
  depend on which.
- No framework, no build step, no external request: fonts, images, icons and
  video are local files with the paths above.
- The catalogue is rendered from the movies table; card markup, copy and order
  are as in `app.js` (`FILMS` order is the catalogue order).
- The player is a plain `<video>` with native controls off and the custom bar on;
  `http.ServeContent` gives range requests, which is all seeking needs.
- Auth is a full page. Passwords are bcrypt or equivalent; the session is an
  opaque token in an HttpOnly, Secure, SameSite=Lax cookie; sign-out deletes the
  server record.
- The preview query flags stay out of the build. The five player states and six
  auth states come from real conditions: no session, expired session, unknown
  film, media error.
