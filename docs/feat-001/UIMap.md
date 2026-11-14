---
feature_code: feat-001
service: clover
version: 1
status: draft
spec_version: 1
figma_files: []
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
derived_confirmed: null
pr_history: []
pr_url: null
last_updated: 2026-11-14
---

# UI Map — feat-001

Clover is a dark, warm, photo-led film catalogue: a full-bleed golden image, a heavy uppercase
headline, glass cards and one paper-white facts card, on near-black. There is no Figma file: the
design is the manager's Dribbble reference, built as a clickable mock in `ui/mock/`, and this map
is the source of truth for it. Requirements live in `spec.md` (v1); this file never restates them.

The design is the mock: `ui/mock/index.html#<screen id>` (screen ids below), one screen per page
state, `ui/mock/screens.json` lists them for `/ui-preview`.

## Figma sources

No Figma file exists for this feature. The manager supplied a video reference instead; it is the
visual truth, and the mock is the artifact the review reads.

| File | URL | Version / last modified | Pages used |
|---|---|---|---|
| MICROLENS landing page (Dribbble shot, `.mp4`, 8 s, 60 fps, 1280 px product window) | https://cdn.dribbble.com/userupload/44295925/file/9d7c0609260c04d650c6bc1d5d8528c4.mp4 | unversioned; supplied 2026-11-14 | The single landing screen (settled at frame-004, 3.00 s) and its two entrance beats (motion, 0.0 s and 2.8 s) |

The shot is a MacBook on a dark backdrop. The read-attachment pass reported `PRESENTATION FRAME`
and cropped every frame to the product window: the desk, the laptop body, its rounded corners and
its shadow are **not** part of this design and are not built into the mock.

The reference is a microbiology site. Its content is not: Clover's copy, films and controls are the
product's own. What is carried over is the visual language (below).

## Requirement mapping

Every active `REQ-*` in `spec.md` v1 appears here. Because there is no Figma file, **frame name**
holds the mock screen id (`ui/mock/index.html#<id>`), **frame link** is empty, and **follows** is
empty — no pattern frame was followed.

`derived` means the element is visible in the reference's settled frame or measured in its motion;
`assumed` means the mock adds it in that same language.

| REQ | Requirement (short) | Page | Frame name (mock screen) | Frame link | Follows | Source | Notes |
|---|---|---|---|---|---|---|---|
| REQ-001 | Visitor lands on the catalogue | Catalogue `/` | `#catalogue` | — | — | derived | The reference is this landing screen: full-bleed photo, headline, glass cards |
| REQ-002 | Catalogue lists all 20 films | Catalogue `/` | `#catalogue` | — | — | assumed | Grid of glass cards; the reference shows one row of glass cards, the grid is our extension |
| REQ-003 | Pick a film signed in → its player | Catalogue `/` → Player | `#catalogue-signed-in` → `#player` | — | — | assumed | Player page is not in the reference |
| REQ-004 | Pick a film signed out → sign in → that film's player | Catalogue `/` → Sign in → Player | `#signin` | — | — | assumed | The picked film is carried in the target chip |
| REQ-005 | Sign in with username or email + password | Sign in `/sign-in` | `#signin`, `#signin-direct` | — | — | assumed | One identifier field, labelled "Username or email" |
| REQ-006 | No self sign-up | Sign in `/sign-in` | `#signin` | — | — | assumed | One line under the form: "No sign-up in v1. Accounts are created by Clover." |
| REQ-007 | Session persists across reloads | Catalogue `/` | `#catalogue-signed-in` | — | — | assumed | Shown as the header action reading "Sign out" instead of "Sign in" |
| REQ-008 | Sign out, playback locks again | Catalogue `/` | `#catalogue-signed-in` → `#catalogue` | — | — | assumed | Header action; toast "Signed out."; cards read "Sign in to watch" again |
| REQ-009 | Player with standard controls | Player `/watch/{movieId}` | `#player` | — | — | assumed | Play/pause, timeline, volume with mute, fullscreen |
| REQ-010 | Progressive playback, seeking | Player `/watch/{movieId}` | `#player` | — | — | assumed | Timeline is a range input; no buffering spinner in the design |
| REQ-011 | Direct player address while signed out | Sign in `/sign-in` | `#signin` | — | — | assumed | Same screen as REQ-004, reached without a click; the film is inferred from the address |
| REQ-012 | Film title beside the video | Player `/watch/{movieId}` | `#player` | — | — | assumed | Title, description and "Film NN / 20" sit left of the stage |
| REQ-014 | Interface is English | All pages | all screens | — | — | assumed | All copy is English, no locale switch |
| REQ-015 | Abandoning sign-in keeps browsing | Sign in `/sign-in` | `#signin` | — | — | assumed | "Cancel and keep browsing" quiet button, plus "Back to catalogue" in the header |
| REQ-016 | Wrong credentials → one generic error | Sign in `/sign-in` | `#signin-error` | — | — | assumed | One alert above the actions: "Incorrect username/email or password."; the password is cleared, the identifier kept |
| REQ-017 | Empty identifier or password rejected | Sign in `/sign-in` | `#signin-required` | — | — | assumed | Per-field error text under the empty field, focus moved to it |
| REQ-018 | Player address for a missing film | Film not found `/watch/{movieId}` | `#not-found` | — | — | assumed | Title "Film not found", one way back to the catalogue |
| REQ-019 | Video cannot load → player says so | Player `/watch/{movieId}` | `#player-error` | — | — | assumed | Error card over the stage with a "Try again" primary |
| REQ-020 | Session expires on the player | Sign in `/sign-in` | `#signin` | — | — | assumed | Lands on the sign-in screen with the same film in the target chip |
| REQ-023 | Empty catalogue | Catalogue `/` | `#catalogue-empty` | — | — | assumed | Glass panel, "No films available", no counts |
| REQ-024 | Poster fails → card still useful | Catalogue `/` | `#catalogue-poster-error` | — | — | assumed | Poster area shows "Poster unavailable"; title, description and link stay |
| REQ-025 | Identifier rules | Sign in `/sign-in` | `#signin-required`, `#signin-error` | — | — | assumed | UI labels the field and pre-checks presence only; format is validated by the backend and answered with the REQ-016 alert |
| REQ-026 | Password rules | Sign in `/sign-in` | `#signin-required`, `#signin-error` | — | — | assumed | Same: presence only in the browser, rules enforced by the backend |
| REQ-028 | No lockout after failed attempts | Sign in `/sign-in` | `#signin-error` | — | — | assumed | The form stays usable and the identifier stays filled; nothing disables it |

### Requirements without UI

| REQ | Reason |
|---|---|
| REQ-013 | Every film plays a sample video: content, not interface |
| REQ-021 | Sign-out is enforced on the server: the UI only offers the sign-out action |
| REQ-022 | The same account on several browsers: no screen differs per device |
| REQ-027 | Unique usernames and emails: a seed-data rule with no visitor-facing surface |
| REQ-029 | Catalogue fields come from seed data: nothing on the site validates or edits them |

### Frames without a requirement

| Frame | What it is | Decision |
|---|---|---|
| frame-004, right edge: vertical social-icon stack and white close disc | The reference site's own chrome | Removed — Clover has no social links, no menu and no modal |
| frame-004, bottom centre: three cropped glass tiles ("INTERACTIVE VISUALS") | The reference's next section, cut by the window edge | Removed — the film catalogue takes this place |
| frame-004, left card stat "2X" / "32K" thumbnails | The reference's content | Replaced by the film poster and the film index badge |
| The whole shot: desk, laptop body, rounded window corners, drop shadow | How the shot was presented | Excluded from the product |

## Reference scenes

Key frames are in the attachment output (`frame-001.png` … `frame-008.png`, 1 s apart). Only one
screen exists in the reference; the rest of Clover is drawn in its language.

| Scene | Frame / time | What the reference shows | What the mock does |
|---|---|---|---|
| S1 — first beat in motion | frame-001, 0.00 s | The photo is already there and steady; the headline and paragraph are still settling | Catalogue arrives in one beat: the photo stays static while the text and the facts card rise together (MO-001 … MO-005) |
| S2 — hero settled | frame-003, 2.00 s | Photo sharp, headline and small paragraph in place; the lower block still absent | Everything is in place about 0.7 s after the page arrives; the mock does not hold the lower block back to 2.8 s |
| S3 — second beat settled | frame-004, 3.00 s | Glass pill cards, white facts card and tiles have arrived | The facts card and the browse chip arrive with the same beat, on the same easing, with no second wait (MO-004, MO-005) |
| S4 — still | frame-005 … frame-008 | Nothing moves after 3.0 s | Static; only hover and interaction motion from here |

## Pages

### Catalogue `/`

- **Purpose:** show the whole public catalogue and the two ways forward (pick a film, or sign in first).
- **Roles:** visitor (signed out) and viewer (signed in) — the catalogue is public either way.
- **Entry / exit:** entry is the site root, the brand, "Back to catalogue", cancel on sign-in, or a
  successful direct sign-in; exit is a film card (to Sign in when signed out, Player when signed in).
- **Content and fields**

| Field | Type | Required | Rule | Source |
|---|---|---|---|---|
| Hero eyebrow | text | — | "Clover / Film catalogue" | assumed |
| Headline | text | — | "Pick a film. Press play." | assumed |
| Lede | text | — | 20 films, free to browse, sign in to watch (REQ-001, REQ-004) | assumed |
| Facts card | two value/label rows on paper | — | Value 1 is the catalogue length (REQ-002); value 2 is "Free" / "browse without an account" | assumed |
| Browse the catalogue chip | button | — | Scrolls to the grid | assumed |
| Film card | link card | — | Poster, index badge, title, two-line description, arrow disc; the whole card is the link (REQ-002) | assumed |
| Film poster | image | — | Decorative alt (the card carries the name); fallback when it fails (REQ-024) | assumed |
| Film title | text | — | Uppercase, up to two lines | assumed |
| Film description | text | — | Clamped to two lines | assumed |
| Catalogue count | text | — | "N films", bound to the same data as the cards (REQ-002) | assumed |
| Header action | button | — | "Sign in" signed out, "Sign out" signed in (REQ-007, REQ-008) | assumed |

- **Actions:** pick a film (REQ-003 / REQ-004), browse to the grid, sign in or out from the header,
  skip to content.
- **States:** default (`#catalogue`), signed in (`#catalogue-signed-in`), loading — skeleton cards,
  no numbers (`#catalogue-loading`), empty — panel, count and facts card hidden, the lede says the catalogue came back empty (REQ-023, `#catalogue-empty`), poster
  failed — two cards show the fallback (REQ-024, `#catalogue-poster-error`).
- **Responsive:** grid 4 columns above 1180 px, 3 from 1180 px, 2 from 820 px, 1 below 620 px;
  the hero's two-column foot stacks at 1020 px; the header's nav links hide at 820 px and the brand
  wordmark at 380 px; the browse chip goes full width at 620 px. Below 620 px the photo gets an extra
  `--c-scrim-mobile` veil and the lede steps to `--c-text`, because the photo sits directly behind the
  copy at phone widths.
- **Screens:** `#catalogue`, `#catalogue-signed-in`, `#catalogue-loading`, `#catalogue-empty`,
  `#catalogue-poster-error`.

### Sign in `/sign-in`

- **Purpose:** authenticate a seeded account, and carry the film the visitor picked through to its player.
- **Roles:** visitor (signed out).
- **Entry / exit:** entry is a film card while signed out (REQ-004), a direct player address
  (REQ-011, REQ-020), or the header's "Sign in"; exit is the player for the carried film, the
  catalogue after a direct sign-in, or the catalogue via cancel (REQ-015).
- **Content and fields**

| Field | Type | Required | Rule | Source |
|---|---|---|---|---|
| Username or email | text input | yes | Required, trimmed; username 3–50 of `[A-Za-z0-9_]`; valid email; case-insensitive (REQ-005, REQ-025) | assumed |
| Password | password input | yes | Required; 8–128 chars with at least one letter and one number; taken exactly as typed (REQ-026) | assumed |
| Sign in | primary button | — | Submits; shows a spinner and disables while submitting (REQ-005) | assumed |
| Cancel and keep browsing | quiet button | — | Returns to the catalogue (REQ-015) | assumed |
| Target chip | poster thumb + "Signing in to watch" + film title | — | Shown when a film was picked or addressed (REQ-004, REQ-011, REQ-020) | assumed |
| No-sign-up note | text | — | "No sign-up in v1. Accounts are created by Clover." (REQ-006) | assumed |
| Generic error | alert | — | "Incorrect username/email or password." above the actions (REQ-016) | assumed |
| Field error | text under the input | — | "Enter your username or email." / "Enter your password." (REQ-017) | assumed |

- **Actions:** submit, cancel, sign in from the header (hidden on this page — the form's button is the
  only primary), skip to content.
- **States:** for a film (`#signin`), direct — no target chip (`#signin-direct`), fields required
  (`#signin-required`), wrong credentials (`#signin-error`), submitting (`#signin-loading`), success —
  leaves for the player with a "Signed in." toast.
- **Responsive:** one centred glass panel; full width inside the page gutter below 620 px.
- **Screens:** `#signin`, `#signin-direct`, `#signin-required`, `#signin-error`, `#signin-loading`.

### Player `/watch/{movieId}`

- **Purpose:** play the selected film and let the viewer control it.
- **Roles:** viewer (signed in).
- **Entry / exit:** entry is a film card while signed in (REQ-003) or a successful sign-in
  (REQ-004); exit is "Back to catalogue" in the header.
- **Content and fields**

| Field | Type | Required | Rule | Source |
|---|---|---|---|---|
| Film title | text | — | Names the film beside the stage (REQ-012) | assumed |
| Description and index | text | — | Short description; "Film NN / 20" | assumed |
| Video stage | video surface | — | Streams the film; the poster fills the stage until playback starts (REQ-009, REQ-010) | assumed |
| Play / pause | icon button | — | Toggles playback; the centre disc also plays (REQ-009) | assumed |
| Timeline | range, 0–1000 | — | Seeks; filled track and elapsed/total timecodes (REQ-010) | assumed |
| Volume | range, 0–100 | — | Sets volume; hidden below 620 px (REQ-009) | assumed |
| Mute | icon button | — | Mutes and restores the previous volume (REQ-009) | assumed |
| Fullscreen | icon button | — | Expands the stage (REQ-009) | assumed |
| Playback error + Try again | card over the stage | — | Says the film cannot play and offers a retry (REQ-019) | assumed |
| Loading overlay | spinner + "Loading film" | — | Shown while the video loads (REQ-009) | assumed |

- **Actions:** play/pause, seek, mute, volume, fullscreen, retry, back to catalogue.
- **States:** ready (`#player`), loading (`#player-loading`), video failed (`#player-error`).
- **Responsive:** title block left of the stage above 1020 px; stacked with the stage first below
  1020 px; total timecode and volume slider hide below 620 px; the centre play disc shrinks at 380 px.
- **Screens:** `#player`, `#player-loading`, `#player-error`.

### Film not found `/watch/{movieId}`

- **Purpose:** answer an address that names no film in the catalogue.
- **Roles:** anyone; it reveals nothing about the catalogue.
- **Entry / exit:** entry is a bad or stale player address (REQ-018); exit is the catalogue.
- **Content and fields**

| Field | Type | Required | Rule | Source |
|---|---|---|---|---|
| Eyebrow | text | — | "Error 404" | assumed |
| Title and copy | text | — | "Film not found" and one line of explanation (REQ-018) | assumed |
| Back to the catalogue | primary button | — | Returns to the catalogue | assumed |

- **Actions:** back to the catalogue.
- **States:** missing film (`#not-found`) — no loading, empty or error variants.
- **Responsive:** one centred column; the hero type scale applies at 1440 px and 390 px.
- **Screens:** `#not-found`.

## Navigation

| From | Trigger | To | REQ |
|---|---|---|---|
| Catalogue (signed out) | Film card | Sign in, film carried in the target chip | REQ-004 |
| Catalogue (signed in) | Film card | Player for that film | REQ-003 |
| Sign in | Successful submit | Player for the carried film, toast "Signed in." | REQ-004 |
| Sign in (direct) | Successful submit | Catalogue, toast "Signed in. Pick a film to watch." | REQ-005 |
| Sign in | "Cancel and keep browsing" | Catalogue | REQ-015 |
| Sign in | "Back to catalogue" (header) | Catalogue | REQ-015 |
| Player | "Back to catalogue" (header) | Catalogue | — |
| Player (direct address, signed out) | Page load | Sign in with the addressed film carried | REQ-011 |
| Player (session expired) | Any playback action | Sign in with the film carried | REQ-020 |
| Player (unknown film) | Page load | Film not found | REQ-018 |
| Any page | Brand | Catalogue | — |
| Catalogue / Player / Film not found | Header "Sign out" | Catalogue, toast "Signed out.", cards locked again | REQ-008 |
| Catalogue (signed out) | Header "Sign in" | Sign in, no target chip | REQ-005 |

The header action is hidden on Sign in, so that page has exactly one primary action.

## Motion & interaction catalog

The reference was measured frame by frame (`attachments-out/ref/motion.txt`): beat 01 starts at
0.0 s, runs about 1500 ms, is closest to **linear** (progress 0.25/0.46/0.71), has **no zoom and no
drift**, peaks near **2 px blur** on a 1280 px page, and staggers its regions over **300 ms**; beat
02 starts at **2.8 s**, runs about 500 ms, is closest to **ease-in**, has no blur, and staggers over
**400 ms**. Beat 01 is the shot's own camera move over the presentation laptop and beat 02 is the
product entrance, so the mock builds **one** entrance on beat 02's shape — 520 ms, ease-in, regions
180 ms apart — and plays it whenever a screen arrives, first paint included. Nothing waits.

| ID | Trigger | Element | Effect | Duration | Easing | Frame | Source | Note |
|---|---|---|---|---|---|---|---|---|
| MO-001 | Any screen arriving | Hero photo / poster / stage | None: the media holds still, no fade, no zoom, no drift | — | — | frame-001→004 | derived | motion.txt measures the light as steady and the media as still; the moving parts are the blocks around it |
| MO-002 | Any screen arriving | Headline | Fade + rise 18px + blur 2px→0 | 520 ms | ease-in | frame-003→004 | derived | line 02 measures about 500 ms and closest to ease-in; the measured 2 px blur is carried over from line 01; the 18 px rise is ours |
| MO-003 | Any screen arriving | Eyebrow, nav, lede | Fade + rise 18px + blur 2px→0 | 520 ms, delays 40–60 ms | ease-in | frame-003→004 | derived | line 02's 400 ms region spread, tightened to 180 ms so the beat lands together |
| MO-004 | Any screen arriving | Paper facts card, watch meta | Fade + rise 18px + blur 2px→0 | 520 ms, delay 40–140 ms | ease-in | frame-003→004 | derived | line 02 measures the ease-in and a steady light with no blur; the blur and the 18 px are ours |
| MO-005 | Any screen arriving | Browse chip | Fade + rise 18px + blur 2px→0 | 520 ms, delay 180 ms | ease-in | frame-003→004 | derived | line 02 (same beat, same easing); it is the last region of the beat |
| MO-006 | Catalogue arriving | Film cards | Fade + rise 22px + scale 0.985→1 + blur 4px→0, 25 ms apart, capped at 8 cards | 520 ms, delays 0–180 ms | ease-in | — | assumed | The grid is the mock's own; the stagger and easing follow the measured beat |
| MO-007 | Leaving any page | Page being left | Removed in one frame: no fade, no scale, no blur | 0 ms (a cut) | — | — | derived | The reference changes screens in a single cut; the arriving screen's beat does the work |
| MO-008 | Arriving on Sign in, Player, Film not found | Form panel, stage, meta, copy | Fade + rise 18px + blur 2px→0 | 520 ms, delays 0–180 ms | ease-in | — | assumed | Same entrance as MO-002 … MO-005, on the same beat |
| MO-009 | Hover a film card | Card and its poster | Card lifts 6px, border brightens, glass brightens, shadow appears; poster scales 1.05 | 240 ms (poster 240 ms) | ease-out | — | assumed | Hover is never visible in the reference; this is the mock's own |
| MO-010 | Hover a film card | Arrow disc | Background paper→amber, disc rotates 45° | 140 ms / 240 ms | ease-out | — | assumed | — |
| MO-011 | Press any button, icon button or card | Pressed control | Scale 0.97 (buttons), 0.94 (icon buttons), 0.99 (cards) | 140 ms | ease-out | — | assumed | — |
| MO-012 | Sign-in rejected | Form alert | Fade + rise −6px→0 | 240 ms | ease-out | — | assumed | — |
| MO-013 | Sign-in submitting | Primary button | Label → spinner, disabled | 700 ms per turn | linear | — | assumed | The spinner is the only perpetual motion on the screen |
| MO-014 | Catalogue loading | Skeleton cards | Shimmer sweep | 1600 ms, infinite | linear | — | assumed | — |
| MO-015 | Player loading | Stage spinner | Rotate 360° | 900 ms, infinite | linear | — | assumed | — |
| MO-016 | Sign-in success, sign-out | Toast | Appears in one frame, auto-hides after 4 s | 0 ms | — | — | assumed | — |
| MO-017 | Play / pause | Centre play disc | Fades and scales to 0.86 while playing | 240 ms | ease-out | — | assumed | — |
| MO-018 | `prefers-reduced-motion: reduce` | Every animation and transition | Duration and delay 0.01 ms, iteration 1, `scroll-behavior: auto` | — | — | — | assumed | — |

The timings above were re-measured off the mock's own recording after the manager's revision
(`ui/review/round-2.md`, 2026-11-14). The first paint is one `rise-in` beat starting at 0.00 s, about
700 ms end to end with the stagger, a 2 px blur that clears by the end and regions starting 200 ms
apart; a screen change is the same beat measured at 2.3 s, about 600 ms, regions 100 ms apart.
`motion-diff.mjs` (`ui/review/motion-diff.txt`) pairs the two against the reference and reports:

- `OK` — reference **beat 02** (2.8 s, 500 ms) against the mock's screen change (2.3 s, 600 ms). The
  product entrance the reference shows is built at the reference's speed and curve.
- `DIFFERS` — reference **beat 01** (0.0 s, 1500 ms, linear) against the mock's first paint (0.0 s,
  700 ms): weak on duration and on the midpoint. This one is deliberate. Beat 01 is the shot's own
  camera move over the presentation laptop (`motion.txt` measures the page light as steady while the
  whole frame moves), and the manager reported the mock's 4.4 s two-beat landing as too slow.
  Rebuilding 1500 ms into the product would rebuild the shot, so the quicker beat stays and this
  line stays honestly red.

The per-element curve is `cubic-bezier(0.42, 0, 1, 1)` — the reference's ease-in — but a whole-frame
fit reports a looser curve for both beats (mock progress 0.11/0.69/0.83 on the first paint and
0.32/0.43/0.61 on a screen change), because the staggered regions and the photo's own brightness
overlap in the measurement. The strips (`motion-01.png`, `motion-02.png` in the measurement folder)
show the blocks rising and settling as specified.

## Design tokens

All values are `:root` custom properties in `ui/mock/styles.css`; no screen rule carries a raw
colour, type size, radius, blur or shadow. `mock-sweep.mjs` reports no `DRIFT` line.

### Colour

| Token | Value | Where used | Source | Note |
|---|---|---|---|---|
| `--c-ink` | `#060504` | Page background | derived | frame-004 |
| `--c-ink-soft` | `#0c0a08` | Card media bed, inputs | assumed | — |
| `--c-paper` | `#f6f1e7` | Facts card, primary button, arrow discs | derived | frame-004 (white card, white pill) |
| `--c-paper-hover` | `#ffffff` | Hover on paper surfaces | assumed | — |
| `--c-paper-deep` | `#e9e1d3` | Paper gradient's far end | assumed | — |
| `--c-paper-ink` | `#17120c` | Text on paper | derived | frame-004 |
| `--c-paper-muted` | `#5b5245` | Labels on paper | assumed | — |
| `--c-text` | `#f7f2e9` | Primary text | derived | frame-004 |
| `--c-text-muted` | `#c9bda9` | Paragraphs, card descriptions | derived | frame-004 |
| `--c-text-faint` | `#a99d8a` | Meta lines, footer | assumed | — |
| `--c-accent` | `#e8a24b` | Text selection, the now-playing dot, the played part of the timeline | derived | the photo's amber |
| `--c-accent-bright` | `#f3b963` | Eyebrow labels, the hovered arrow disc, the end of the timeline fill, the spinner and the toast icon | assumed | — |
| `--c-glass` | `rgba(255,238,216,.10)` | Glass cards and panels | derived | frame-004 |
| `--c-glass-strong` | `rgba(255,238,216,.17)` | Glass hover, ghost button | assumed | — |
| `--c-glass-deep` | `rgba(18,14,10,.55)` | Control bar, index badge | assumed | — |
| `--c-header` | `rgba(6,5,4,.72)` | Sticky header | assumed | — |
| `--c-line` | `rgba(255,245,232,.16)` | Hairline borders | derived | frame-004 |
| `--c-line-strong` | `rgba(255,245,232,.30)` | Hover borders | assumed | — |
| `--c-line-paper` | `rgba(20,16,11,.16)` | Divider between the facts-card rows, spinner track on the paper button | derived | frame-004 |
| `--c-shadow` | `rgba(0,0,0,.55)` | Shadow base | derived | frame-004 |
| `--c-scrim-top` | `rgba(6,5,4,.78)` | Photo scrim, top | derived | frame-004 |
| `--c-scrim-mid` | `rgba(6,5,4,.30)` | Photo scrim, middle | derived | frame-004 |
| `--c-scrim-bottom` | `rgba(6,5,4,.94)` | Photo scrim, bottom | derived | frame-004 |
| `--c-scrim-mobile` | `rgba(6,5,4,.60)` | Extra veil under the hero and sign-in scrims below 620 px, where the photo sits tight behind the copy | assumed | — |
| `--c-overlay` | `rgba(6,5,4,.42)` | Stage overlay | assumed | — |
| `--c-input` | `rgba(6,5,4,.38)` | Field background | assumed | — |
| `--c-skeleton` | `rgba(255,238,216,.08)` | Loading blocks | assumed | — |
| `--c-progress-track` | `rgba(255,245,232,.24)` | Timeline and volume tracks | assumed | — |
| `--c-toast` | `rgba(18,14,10,.88)` | Toast background | assumed | — |
| `--c-focus` | `#f3b963` | Focus ring on dark and over photos | assumed | — |
| `--c-focus-paper` | `#17120c` | Focus ring on paper | assumed | — |
| `--c-danger` | `#ffb49b` | Field error text, the alert's left rule | assumed | — |
| `--c-danger-text` | `#ffd9cc` | The alert's message | assumed | — |
| `--c-danger-line` | `rgba(255,150,120,.45)` | Border of a field in error and of the player's error card | assumed | — |

Gradients: `--g-hero-scrim` (top/mid/bottom), `--g-page-scrim` (the denser scrim the Player and
Film-not-found carry at every width), `--g-glass` and `--g-glass-hover` (155° glass),
`--g-glass-panel` (dark panel), `--g-paper` (168° paper), `--g-image-scrim` (poster foot),
`--g-stage-scrim`, `--g-page-glow`, `--g-skeleton`.

### Typography

One family: **Instrument Sans** (open licence), 400 / 500 / 700 in use. Two files ship — Regular and
Bold; the Regular face also serves the 500 labels, so nothing is synthesised. The reference's font is not
identifiable from the video; this is the closest open match for its tight, heavy grotesque.
Letter-spacing tokens: `--ls-tight −0.03em`, `--ls-snug −0.01em`, `--ls-label 0.08em`,
`--ls-wide 0.14em`, `--ls-wider 0.18em`. Line heights: `--lh-tight 0.9`, `--lh-snug 1.12`,
`--lh-normal 1.55`.

| Token | Value | Rendered at 1440 px | Role | Source | Note |
|---|---|---|---|---|---|
| `--fs-hero` | `clamp(3rem, 6.6vw, 6.75rem)` | 95 px | Catalogue and Film-not-found headline | derived | frame-004's headline scale |
| `--fs-title` | `clamp(2rem, 4.2vw, 3.9rem)` | 60.5 px | Player title | assumed | our step down from the hero scale |
| `--fs-display` | `clamp(2rem, 3.4vw, 3rem)` | 48 px | Section title, empty-state title | assumed | our section scale, not in the reference |
| `--fs-panel-title` | `clamp(1.9rem, 3vw, 2.6rem)` | 41.6 px | Sign-in title | assumed | — |
| `--fs-card` | `1.25rem` | 20 px | Film title, error card title | assumed | — |
| `--fs-lede` | `1.0625rem` | 17 px | Hero paragraph | derived | frame-004's copy scale |
| `--fs-body` | `0.9375rem` | 15 px | Body copy, inputs | assumed | — |
| `--fs-small` | `0.8125rem` | 13 px | Card descriptions, footer | assumed | — |
| `--fs-tiny` | `0.6875rem` | 11 px | Labels, timecodes, index badge | assumed | — |
| `--fs-eyebrow` | `0.6875rem` | 11 px | Uppercase eyebrow labels | derived | frame-004 |
| `--fs-nav` | `0.75rem` | 12 px | Header links | derived | frame-004 |
| `--fs-btn` | `0.75rem` | 12 px | Button and pill labels | derived | frame-004 |

### Spacing and shape

| Token | Value | Where used | Source | Note |
|---|---|---|---|---|
| `--s-1` … `--s-9` | 4, 8, 12, 16, 24, 32, 48, 64, 96 px | Gaps, padding, section rhythm | assumed | — |
| `--page-x` | `clamp(20px, 4vw, 48px)` | Page gutter | assumed | — |
| `--r-xs` | 8px | Proto list rows | assumed | — |
| `--r-sm` | 12px | — | assumed | — |
| `--r-md` | 18px | Fields, control bar, tiles | derived | frame-004's card corners |
| `--r-lg` | 26px | Film cards | derived | frame-004 |
| `--r-xl` | 34px | Sign-in panel, empty panel | assumed | the shot's largest corner is the presentation window, not a product surface |
| `--r-pill` | 999px | Buttons, chips, badge, toast | derived | frame-004 |
| `--r-round` | 50% | Discs: arrow, play, spinner, dots, slider thumbs | assumed | — |
| `--blur-glass` | 18px | Cards, chips, empty panel | derived | frame-004's frosted cards |
| `--blur-panel` | 28px | Control bar, toast, header | assumed | — |
| `--blur-photo` | 8px | Blurred photo fallbacks | assumed | — |
| `--shadow-card` | `0 24px 60px rgba(0,0,0,.45)` | Sign-in panel, player stage, empty panel, toast | derived | frame-004 casts one under its glass cards |
| `--shadow-paper` | `0 22px 48px rgba(0,0,0,.38)` | Paper facts card, centre play disc | derived | frame-004 |
| `--shadow-lift` | `0 30px 80px rgba(0,0,0,.58)` | Film card on hover only | assumed | — |

Two distinct shadows are visible in the swept states — the panel/stage one and the paper facts
card's; the lift appears on card hover only.

### Motion tokens

`--dur-fast 140ms` (hover tints, press, arrow), `--dur-base 240ms` (poster zoom, alert, play disc),
`--dur-slow 520ms` (the screen entrance and its stagger); `--ease-out cubic-bezier(.22,1,.36,1)`,
`--ease-in cubic-bezier(.42,0,1,1)`, `--ease-linear cubic-bezier(0,0,1,1)`. The entrance duration
and easing are derived from reference beat 02; the rest are assumed.

## Components

None of these has a Figma `Code` (no Figma file), and **nexcomponent is not used anywhere** — the
manager's constraint is a separate browser frontend with no nexcomponent. Each piece is defined once
in `ui/mock/styles.css` and every screen is built from these pieces.

| Component | Figma `Code` / variants | nexcomponent export | Resolution status | States | Used on | Note |
|---|---|---|---|---|---|---|
| Brand mark + wordmark | — | not used | built in the mock | default | Header, all pages | Small clover outline + "Clover" |
| Header | — | not used | built in the mock | catalogue, other page, signed in | All pages | Sticky, glass-tinted; links hide at 820 px |
| Primary capsule button | — | not used | built in the mock | default, hover, press, focus-visible, disabled/loading | Header, Sign in form, error cards, Film not found | White capsule, 12 px bold uppercase black label, 14 px up-right arrow inside the tail, and a wavy right edge: a 44 × 44 px tail with one concave bite against the body and a blunt tip (frame-004, manager's button reference) |
| Ghost pill button | — | not used | built in the mock | default, hover, press | Header when signed in | Glass fill over the photo |
| Quiet button | — | not used | built in the mock | default, hover, focus | Cancel on Sign in | Underlined small text |
| Icon button (disc) | — | not used | built in the mock | default, hover, press | Player controls | Paper disc, 40 px |
| Centre play disc | — | not used | built in the mock | default, hover, press, hidden while playing | Player stage | Large paper disc |
| Arrow disc | — | not used | built in the mock | default, hover | Film card, browse chip | 36 px paper circle; rotates and turns amber on card hover |
| Glass film card | — | not used | built in the mock | default, hover, press, skeleton, poster failed | Catalogue | 16:10 poster, index badge, title, description, arrow |
| Paper facts card | — | not used | built in the mock | default | Catalogue hero | Two value/label rows, divider, drop shadow (frame-004) |
| Browse chip | — | not used | built in the mock | default, hover, press | Catalogue hero | Glass pill with a down arrow disc |
| Field | — | not used | built in the mock | default, focus, filled, error | Sign in | Label, glass input, error line |
| Form alert | — | not used | built in the mock | hidden, shown | Sign in | `role="alert"`, danger tint |
| Target chip | — | not used | built in the mock | hidden, shown | Sign in | Poster thumb + film title |
| Player stage | — | not used | built in the mock | ready, playing, loading, error | Player | Poster bed, scrim, controls bar |
| Timeline and volume | — | not used | built in the mock | default, focus, dragged | Player | Range inputs styled to the tokens |
| Toast | — | not used | built in the mock | hidden, shown | Sign in, sign-out | Bottom centre, pill, 4 s |
| Poster fallback | — | not used | built in the mock | shown | Catalogue | Icon + "Poster unavailable" |
| Skeleton card | — | not used | built in the mock | shimmering | Catalogue loading | Same footprint as a card |
| Spinner | — | not used | built in the mock | spinning | Sign-in button, player overlay | Two sizes |
| Empty panel | — | not used | built in the mock | shown | Catalogue empty | Glass panel, circular icon, title, copy |

## Assets

| Asset | Location in Figma | Format | Note |
|---|---|---|---|
| `hero.webp` | — | WebP, 58 KB | "Gold" by Yogendra174, BY 2.0 — the golden liquid photo behind the catalogue hero and the other page backgrounds |
| 20 film posters (`<film-id>.webp`) | — | WebP, one per film | Openly licensed photographs, licence and source per file in `ui/mock/assets/img/CREDITS.md` |
| `InstrumentSans-Regular.ttf`, `InstrumentSans-Bold.ttf` | — | TTF | Open Font Licence, bundled locally; no network call. The Regular face serves the 500 labels |

The mock folder stays under 5 MB and every image under 300 KB. All paths are relative; the mock makes
no network request.

## Open design gaps

- The reference shows **one screen only**. Sign in, Player and Film not found are drawn in its
  language and need the manager's confirmation in `/ui-preview` (owner: manager).
- The **video itself** is not in the mock: the stage shows the film's poster with working controls.
  The real stream is `/code-preparation` and `/fe-dev` work (owner: engineering).
- The reference's **social stack, close disc and bottom tile row** are deliberately not carried over
  (see "Frames without a requirement"); say so if any of them should exist in Clover (owner: manager).
- **Responsive behaviour is ours** — the reference is a desktop shot inside a laptop. The breakpoints
  and stacked layouts need a look at 390 px in `/ui-preview` (owner: manager).
- **Copy** is final English product copy written for Clover; the manager may want to change the hero
  headline and the empty-state wording (owner: manager).

## Polish review

Reviewed against the polish list; "fixed in the mock" is this run's equivalent of "fixed in Figma",
because there is no Figma file. Sweep evidence: `mock-sweep.mjs` reports no DRIFT, contrast,
overflow, console, focus-ring or reduced-motion finding across 14 screens × 4 widths.

| Item | Frame | Finding | Decision |
|---|---|---|---|
| Filler badges or decorative pills | frame-004 | The `2X` / `32K` stats are the reference's own content; Clover's index badge ("01") carries ordering | none found |
| Labels repeating a heading or value | frame-004 | "Film 01 / 20" adds the position in the catalogue that the title does not | none found |
| Italics on random words | all | No italics anywhere | none found |
| Off-system values | all | Raw tracking values and an 11 px font size were folded into tokens; no raw colour, size, radius, blur or shadow remains in any screen rule, and `mock-sweep` reports no DRIFT | fixed in the mock |
| Placeholder content | all | All copy is final product copy; the films are the seeded catalogue | none found |
| Main feature not dominant | frame-004 | The hero keeps the reference's composition; the catalogue grid is one scroll below and carries the product's real job | none found |
| Vague headline | frame-004 | "Pick a film. Press play." names the task | none found |
| Competing primary actions | Catalogue, Sign in, Film not found | Sign in hides the header's "Sign in" so the form's button is the only primary; on Film not found the header action is demoted to a ghost so "Back to the catalogue" wins. One primary per screen | fixed in the mock |
| No single accent colour | frame-004 | One amber accent: the eyebrows, the timeline fill, the hovered arrow disc and the toast icon | none found |
| Embellishments with no purpose | frame-004 | Glass, blur and shadow follow the reference and separate layers over the photo; the hero photo's glow is the photo itself | none found |

## Revision — manager review round 2 (2026-11-14)

The manager looked at the mock in `/ui-preview` and asked for two changes before approving: the
arrival is too slow, and the primary button does not match the button reference they attached. The
work stays on the same branch and updates PR #69; nothing is merged. The decision is also recorded
in `ui/review/round-2.md`.

**Motion.** The two-beat landing (1500 ms + a 2800 ms wait) is gone. Every screen now arrives on one
beat measured off reference beat 02 — 520 ms, `cubic-bezier(0.42, 0, 1, 1)`, regions 0–180 ms apart,
2 px blur clearing to 0 — and the first paint uses that same beat, so the old "wait for the grid"
delay no longer exists. Moving between screens is now a cut: the leaving screen is removed in one
frame and only the new screen animates (MO-007), so the entrance never plays twice. The mock's own
recording measures the first paint at 0.00 s, about 700 ms with the stagger, 2 px blur and regions
200 ms apart, and a screen change at about 600 ms with regions 100 ms apart. `motion-diff.mjs`
(`ui/review/motion-diff.txt`) now reports **`OK` for reference beat 02** — the product entrance — and
one difference that is intentional: it pairs reference beat 01, the shot's 1500 ms linear camera move
over the presentation laptop, with the mock's first paint and calls the mock weak on duration and
midpoint. Building that 1500 ms back would rebuild the shot and reintroduce the slowness the manager
reported, so the quicker beat stays.

**Primary button.** White capsule, 12 px bold uppercase black label, a 14 px up-right arrow sitting
inside the tail, and a wavy right edge: a 44 × 44 px tail drawn with `clip-path` — one concave bite
diving out of the body's top edge, a blunt tip at the far right, and a lower edge sweeping back into
the capsule's bottom — over a fill that stops 44 px short so the bite stays transparent. The
geometry was measured off the reference's own button in frame-004, column by column on the white
pixels: a 102 × 35 px capsule whose bite carries about 69% of its height, whose tail ends blunt at
the right, and whose arrow sits inside the tail. Those proportions were scaled to Clover's 44 px
button, which is the taller of the two. Hover lightens the capsule to white and nudges the arrow
up-right, press scales to 0.97, focus draws the stock outline around the whole silhouette. It is
applied to every primary button: header "Sign in"/"Sign out", the Sign in form's submit, "Try again"
on the player error and "Back to the catalogue" on Film not found; the ghost button keeps no arrow.
Label changes go through a dedicated label span so the arrow survives the sign-in and sign-out text
swaps.

The manager's attached image (`uploads/5c4cf17b-b2d7-469b-b68c-9f98836dc9d8-image.png`) was not in
the session — the attachment reader reports "file not found" and `uploads/` holds only the earlier
summary — so the shape was measured from the reference video's own button and checked against the
manager's written description. If that image shows a different tail, the difference is in the bite
depth and the tip only; the fill, label, arrow and states are unchanged by it.

**Verification.** `mock-sweep.mjs` reports no findings across 14 screens × 1440/1024/768/390 px
(console, overflow, broken images, contrast, focus ring, fixed-layer overlap, dead links,
reduced-motion); the button's default, hover, press and focus states were captured at 3× and the
contrast walk treats the capsule's gradient as an image, as designed.
