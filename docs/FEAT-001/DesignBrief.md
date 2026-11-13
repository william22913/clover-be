---
feature_code: FEAT-001
service: clover
version: 2
status: in-review
spec_version: 1
reference: https://cdn.dribbble.com/userupload/44295925/file/9d7c0609260c04d650c6bc1d5d8528c4.mp4
last_updated: 2026-11-13
---

# Design brief — clover FEAT-001

## What this is

Clover is a movie streaming web app. v1 is the manager's whole journey in one
line: "open the website, pick a movie, and start watching." The catalogue is
public — poster, title, short description, nothing else — and watching needs an
account. v1 has no search, no genres, no My list, no saved progress, no ratings
and no subscriptions.

The manager supplied one design reference, a Dribbble screen recording, and
asked the interface to match it closely. This file is the reading of the
reference, the argument for what was taken from it, and the v1 design system
itself. `UIMap.md` is the screen-by-screen contract, and `ui/mock/` is the
working reference implementation — when this file and the mock disagree, the
mock wins and this file gets fixed.

The v1 tokens, type, glass, grade, layout and motion below were carried in
`DESIGN.md` during the first delivery. The design stage of this repo owns only
`UIMap.md`, `DesignBrief.md` and `ui/`, so they are recorded here instead;
`DESIGN.md` on `main` still describes the previous, out-of-scope draft (search,
genres, My list, a carousel) and is not the v1 contract.

## The decision this draft is built on

The reference video is not a film product. It is a screen recording of a
microbiology landing page inside a 3D laptop mockup (8.0 s, 1600 × 1200,
60 fps, one static hero, no scroll and no second screen). Its words, imagery,
subject and layout are not Clover's, so reproducing them would have meant
inventing a product the spec does not describe — which is exactly what the
previous draft did when it produced Search, Genres and My list screens from the
reference's genre and list cards.

The manager has confirmed the reference is to be matched **for its look and
mood**, and that the screens themselves are Clover's v1 scope. So:

- **Taken from the reference:** the near-black `#050505` field, the amber light
  grading the photograph, a full-viewport hero, the spread top nav with a white
  gooey capsule at its right, Inter Tight white caps at near-full column width,
  frosted rails laid over the photograph, the white stat card and the three
  glass feature cards (rebuilt on Clover's own facts), the organic white capsule,
  180–600 ms orchestrated motion, and a slow drift on the still.
- **Not taken:** its content, copy, imagery, its three-dash carousel indicator,
  its close mark and its social pills — the rail that held them now carries two
  real catalogue films instead.
- **Derived for Clover:** the 13-film catalogue grid under the hero, the
  combined sign-in / create-account screen with the chosen film held on the art
  panel, and the player with its custom bar.

The previous draft's `DesignBrief.md` and `ui/mock/*` have been replaced in
place. Its out-of-scope screens are gone, not hidden. `DESIGN.md` on `main`
still describes that previous draft and is not part of the v1 delivery; the v1
system is in this file.

## The direction in one paragraph

A warm, dark cinema. One photograph fills the first screen and is graded into
amber, so every still sits in the same light. A very large white all-caps line
runs almost the full width of the column — "Pick a film." — with one quiet
sentence under it that states the rule once: browsing is open, watching takes an
account. Around it sit the reference's materials, rebuilt on Clover's facts: a
frosted pill at the right edge counting the catalogue, a glass rail carrying the
first two films, a white card counting 13 films and one account, and three glass
cards for the three steps — browse, sign in, watch. Below the hero, the whole
catalogue in a five-across grid of posters, titles and three-line descriptions,
with no controls above it. When a film is chosen, the same language carries the
account screen: the film's still full-bleed as a promise, the form on the right
in a frosted panel, and a white gooey capsule as the one heavy shape. Then the
film, full screen, with a frosted bar for the five controls.

## The reference, and what is taken from it

The recording is 8.0 s at 1600 × 1200 and 60 fps, and shows a single screen
inside a 3D laptop mockup on a ribbed metallic set, so the frame is graded and
the screen is in perspective. Only one screen exists in it: there are no hover,
scroll or post-hero states to copy.

| From the video | In this design |
| --- | --- |
| Near-black field, warm amber light rising from the lower half | `#050505` page; a screen-blended amber gradient over every still, with three black scrims for legibility |
| Full-screen composition: nav, display line, rails and cards all on one screen | The catalogue hero fills the viewport, capped at 980 px so it stays a composition on tall displays |
| Small tracked-caps nav links spread across the top, logo left, white capsule right | 72 px fixed header: brand, "Catalogue" and "How it works" spread across the middle, and the white gooey "Sign in" capsule at the right |
| White all-caps grotesque headline filling about 92 % of the column | "PICK A FILM." at `clamp(46px, 16.2vw, 270px)`, `-.034em`, one line, about 92 % of the column at 1440 |
| Three lines of quiet subcopy under the headline | One 14.5 px sentence, max 36ch, plus a small frosted icon button that jumps to the grid |
| One photograph edge to edge behind the type | One graded still behind the whole hero, drifting 1.02 → 1.09 over 34 s |
| Frosted rail bleeding in from the right | The facts rail at the right edge — "13 films", "2021–2025" — and, below it, a two-entry film rail that bleeds off the same edge |
| White stat card with a large number and a caption | The catalogue stat card: "13" films in the catalogue, "1" account |
| Three glass feature cards with an icon, a label and a caption | The three step cards: Browse / Sign in / Watch, one line each |
| Squircle primary with a round arrow knob | The gooey white capsule: pill body fused to a knob by an SVG goo filter, arrow inside |
| Orchestrated entrance, about 180–600 ms | Title rises from a mask in 700 ms, lede at 140 ms, rails at 220–300 ms, foot at 340 ms with a 60 ms step stagger, cards at 28 ms |

## What is not from the video, and why

- **The catalogue grid.** The reference has no grid; v1's first requirement
  does. It inherits the poster radius (20 px), the card hover lift and the
  reference's type hierarchy. Five across at wide widths because 13 films tile
  unevenly and five gives the closest rows. The hero is full-viewport, as the
  reference's screen is, so the grid begins below it; SC-001's "browse before
  login" is carried above the fold by the film rail, whose two entries are real
  films with poster, title and description and open that film when selected.
- **The rail carries films, not social icons.** The reference's right-edge rail
  holds share icons and a close mark. Clover's holds two catalogue entries and
  the facts pill holds the catalogue count; nothing in either is decorative.
- **The stat card and the feature cards count real things.** 13 films and one
  account are read from the catalogue constant, and the three step cards are the
  manager's own journey; the reference's invented numbers and labels are not
  reused.
- **A second page for accounts.** The reference has no forms. Sign-in and
  sign-up share one page and one form, switched by two text tabs — the "combined
  login / sign-up screen" the spec asks for — rather than a modal, so there is
  no focus trap and no overlay to dismiss.
- **The chosen film on the account screen.** The reference shows a brand
  headline; Clover shows the film the visitor is one step away from watching,
  because the flow promises to return them to it.
- **A player.** The reference shows none. The bar is glass over the film, the
  top bar is a scrim, and every control is a real button; the visuals defer to
  the picture.
- **Empty, error and expired states.** Written in the product's voice: what
  happened and what to do next. A wrong password, a broken video and a missing
  film never share a treatment.
- **Hover and focus behaviour.** Not evidenced in the video. Built to the house
  floor: 180–250 ms, transform and opacity only, a 2 px `--gold-glow` focus
  ring, and keyboard control for the player.

---

# The v1 design system

This half of the file is the system itself, folded in from the delivered
`DESIGN.md` so the whole v1 contract lives in a file the design stage owns. It
covers the three v1 surfaces only: the public catalogue, the combined sign-in /
create-account screen, and the player.

## How the system is meant to be used

Three rules carry the look:

1. **One photograph at a time.** A page either carries one full-bleed graded
   still — the catalogue hero, the auth art panel, the player frame — or
   none. Never two competing images on one screen.
2. **One white surface per page.** The gooey capsule (sign-in, submit, retry),
   the play disc and the catalogue stat card. Everything else is black, warm
   near-black, glass or hairline.
3. **Amber is light, not paint.** It lives in the grade over the still, the brand
   mark, the active tab, the scrub fill and the focus ring. No amber buttons, no
   amber text blocks, no amber backgrounds.

The reference's single heavy shape is the gooey capsule; it is used for exactly
one action per screen so the material stays special.

## Foundations

### Space

| Token | Value | Use |
| --- | --- | --- |
| `--gutter` | 32 px (24 px ≤ 899, 20 px ≤ 599) | Page inset, header inset |
| `--nav` | 72 px (64 px ≤ 767) | Fixed header height, auth top padding |
| `--wrap` | 1680 px | Maximum content width, centred |
| `--bleed` | gutter + half the space outside `--wrap` | Keeps the facts rail and film rail flush to the right viewport edge |
| Hero | `clamp(620px, 100svh, 980px)` | Catalogue first screen; the 980 px cap keeps it a composition on tall displays and in `/ui-preview`, which sizes the stage to the mock's content height |
| Grid gap | 26 px (22 px 4-up, 20 px 3-up, 16 px 2-up) | Poster grid |
| Catalogue section | 6 px top / 96 px bottom (72 px ≤ 767) | Below the hero |
| Hero foot | 22 px gap, `minmax(250px,.92fr)` stat + `minmax(0,3fr)` steps | Stat card and step cards |
| Auth layout | Full-bleed art, 432 px frosted form panel at the right | Desktop |
| Auth stack | 250 px still band at ≤ 959 (210 px ≤ 599), form below, panel dissolved | Tablet and phone |

Content is left-aligned throughout. Nothing is centred except the player's
message panel, the play disc, icon buttons and the avatar.

### Radius

| Token | Value | Use |
| --- | --- | --- |
| `--r-panel` | 26 px | Film rail, auth form panel |
| `--r-card` | 22 px | Stat card, step cards |
| `--r-poster` | 20 px | Poster art, poster link |
| `--r-thumb` | 14 px | Rail poster thumb |
| `--r-input` | 14 px | Text fields |
| — | 24 px | Player control bar (20 px ≤ 599) |
| — | 12 px | Notices, inline raised panels |
| — | 999 px | White capsules, circular buttons, range track |

Radii never go below 12 px. There are no square corners in Clover.

### Glass

Glass is used only where it carries something over a photograph, and glass is
never stacked on glass:

- **Catalogue rails and cards** — the facts rail, the two-entry film rail and the
  three step cards: `linear-gradient(180deg, rgba(255,255,255,.11), rgba(255,255,255,.045))`
  with `blur(20–26px) saturate(1.1–1.15)` and a 13–16 % white hairline. The film
  rail bleeds off the right edge and keeps only its left corners.
- **Hero icon button** — `--glass` (9 % white), `blur(18px)`, 46 px square,
  radius 15 px.
- **Auth form panel** — a white gradient over `rgba(5,5,5,.34)` with
  `blur(30px) saturate(1.15)`, radius 26 px.
- **Header, scrolled** — `rgba(5,5,5,.84)` with `blur(20px) saturate(1.08)` and a
  `rgba(255,255,255,.08)` hairline.
- **Player bar** — `linear-gradient(180deg, rgba(255,255,255,.11), rgba(255,255,255,.04))`
  over `blur(28px) saturate(1.1)`, a 1 px `--glass-edge` border, radius 24.
- **Player back button** — `rgba(255,255,255,.07)` with `blur(18px)`, a 44 px
  disc, `--glass-edge` border.

The fields are not glass: they are a warm opaque near-black so text stays
legible over the art panel's grade.

### Type

Two families, self-hosted: **Inter Tight 700** for display and titles, **Inter
400/500/600** for interface text. No third family, no serif, no mono.

| Role | Size | Family / weight | Notes |
| --- | --- | --- | --- |
| Catalogue display | `clamp(46px, 16.2vw, 270px)` | Inter Tight 700 | `line-height .86`, `-.034em`, uppercase, one line, ~92 % of the column |
| Rail title | `clamp(16px, 1.5vw, 21px)` | Inter Tight 700 | `line-height 1.02`, uppercase |
| Rail caption | 11.5 px | Inter 400 | `--text-dim`, two lines, max 34ch |
| Stat numeral | 30 px | Inter Tight 700 | `-.03em`, `#111` on white |
| Auth film title | `clamp(34px, 3.5vw, 58px)` | Inter Tight 700 | `line-height .95`, `-.028em`, uppercase |
| Auth art title | `clamp(36px, 5vw, 82px)` | Inter Tight 700 | `line-height .92`, `-.03em`, uppercase |
| Auth heading | `clamp(28px, 2.8vw, 40px)` | Inter Tight 700 | `line-height .98`, `-.03em`, uppercase |
| Player message | `clamp(28px, 3.4vw, 50px)` | Inter Tight 700 | `line-height 1.02`, `-.03em`, uppercase |
| Section title | `clamp(24px, 3.2vw, 38px)` (26 ≤ 899, 22 ≤ 599) | Inter Tight 700 | `-.02em`, uppercase |
| Player title | 17 px (14 ≤ 599) | Inter Tight 700 | uppercase, text-shadow for contrast |
| Card title | 12.5 px | Inter Tight 600 | uppercase, `.02em` |
| Lede | 14.5 px (13.5 ≤ 599) | Inter 400 | max 36ch, 82 % white |
| Body | 14.5–15 px | Inter 400 | `--text` |
| Card description | 12.5 px | Inter 400 | three lines, clamped |
| Field input | 15 px | Inter 400 | 54 px tall |
| Label / tab / header link | 11–11.5 px | Inter 600 | uppercase, `0.10–0.12em` |
| Hint, field error | 12.5 px | Inter 400 / 500 | under their field |
| Count, player meta | 12–13 px | Inter 400 | `--text-dim` |

Display is uppercase with tight leading (.86–.95) and negative tracking;
interface copy is sentence case except for labels, tabs, card titles and the
nav, which are uppercase at 10–12.5 px with 0.04–0.18em tracking. The player
clock is tabular.

The catalogue display is fitted by its clamp rather than measured: at 1440 it
renders about 92 % of the column, which is the reference's scale, and it still
fits one line at 390. If the copy changes, re-check all three widths.

### Colour

| Token | Value | Semantics |
| --- | --- | --- |
| `--ink-000` | `#050505` | Page, player letterbox, the end of every scrim |
| `--ink-100` | `#0b0b0c` | Poster placeholder background |
| `--ink-200` | `#141210` | Input fill (warm near-black) |
| `--paper` | `#ffffff` | Display type, goo capsule, play disc, avatar |
| `--ink-on-paper` | `#111111` | Type and icons on white |
| `--paper-muted` | `#6e6e6e` | Labels on white (reserved, unused in v1) |
| `--paper-rule` | `#e6e4e1` | Rules on white (reserved, unused in v1) |
| `--gold` | `#e0a34e` | Brand mark, active tab, scrub fill, focus-adjacent emphasis |
| `--gold-deep` | `#8b4d11` | Warm grade, deep end |
| `--gold-glow` | `#ffc46b` | Focus ring, hover on gold text |
| `--text` | `rgba(255,255,255,.78)` | Body on black (12:1) |
| `--text-dim` | `rgba(255,255,255,.56)` | Captions, metadata (6.3:1) |
| `--text-strong` | `#ffffff` | Headings and titles |
| `--glass` / `--glass-2` | `rgba(255,255,255,.09)` / `rgba(255,255,255,.05)` | Frosted fills: rails, step cards, hero icon button, player chrome |
| `--glass-edge` | `rgba(255,255,255,.16)` | Hairlines on input and chrome |
| `--glass-edge-strong` | `rgba(255,255,255,.45)` | Hover hairline |
| `--danger` | `#ff9b8b` | Error text (10:1) |
| `--danger-line` | `#e5604d` | Invalid field border |

The amber is a light source, never a fill: it grades the photograph, marks the
active tab with a 2 px underline, colours the played part of the scrub track,
and outlines focus. It is never a button background, a filled accent or a text
block. The only filled accents are white.

### The still grade

Every still gets the same three layers, in this order (z-index in brackets):

1. the image at `brightness(.74) saturate(.94) contrast(1.05)` [0], drifting
   `scale(1.02) → scale(1.09) translateY(-1.2%)` over 34 s, alternating;
2. a screen-blended amber rising from the bottom —
   `rgba(214,124,32,.62) → rgba(168,86,8,.28) 40% → rgba(120,60,10,.08) 72% → transparent` [1];
3. black scrims [2] — a top band (`86 % → 40 % at 22 % → 5 % at 48 %`), a left
   legibility wash (`64 % → 26 % at 40 % → 0 at 72 %`) and a bottom ramp
   (`88 % → 22 % at 22 % → 0 at 48 %`), so the hero copy stays legible over a
   bright still.

The auth art uses the same recipe at `brightness(.78)` with a stronger bottom
ramp (`.94`) because the film title sits on it. This is what makes different
placeholder photographs read as one film. Any replacement still must go through
it, and must be chosen at a similar exposure — a very dark still disappears.

### The gooey capsule

The one signature shape, used for the header's "Sign in" (content width, 44 px
with a 38 px knob), the auth submit (full width) and "Try again".

- A background layer carries `filter: url(#goo)` (Gaussian blur 9, alpha matrix
  `24 / -10`, composite atop) over two white shapes: a pill body with
  `inset: 0 30px 0 0`, and a 48 px circular knob at `right: 0; top: 4px`.
- The filter fuses the knob into the body, producing the organic fillet.
- An unfiltered foreground layer holds the label (12 px, 700, `0.10em`,
  uppercase, `#111`) and the arrow knob on the same grid.
- Height 56 px (52 px ≤ 599); the body must never fill the full width or the
  fillet disappears.
- Hover moves the arrow 2 px up-right; active scales to .99.

## Layout

### Catalogue anatomy

```
┌───────────────────────────────────────────────────────────────────────┐
│ ◘ CLOVER      CATALOGUE        HOW IT WORKS        ( SIGN IN  → )      │ 72 fixed
│                                                                       │
│  PICK A FILM.                               ← display, ~92% column    │
│  The whole catalogue is open. Watching       (⌄)  ┌ 13 FILMS ⏐ 2021–25 ┐│ ← facts rail,
│  takes an account…                                └───────────(⌄)─────┘│   bleeds right
│                                                                       │
│  ┌──────────────────────────────┬──────────────────────────────┐      │
│  │ ▭ 2023 HALCYON BAY        (→) │ ▭ 2025 THE QUIET PROVINCE (→)│      │ ← glass film
│  │    A marine biologist…        │    A village's last winter…  │      │   rail, bleeds
│  └──────────────────────────────┴──────────────────────────────┘      │   right
│                                                                       │
│  ┌───────────┐  ┌──────────┐ ┌──────────┐ ┌──────────┐                │
│  │ 13  films │  │ ① Browse │ │ ② Sign in│ │ ③ Watch  │                │ ← white stat
│  │  1 account│  │          │ │          │ │          │                │   card + glass
│  └───────────┘  └──────────┘ └──────────┘ └──────────┘                │   step cards
├───────────────────────────────────────────────────────────────────────┤
│  THE CATALOGUE                            13 films, free to browse     │
│  ┌────┐ ┌────┐ ┌────┐ ┌────┐ ┌────┐    5 columns ≥ 1200                │
│  │ 2:3│ │    │ │    │ │    │ │    │                                   │
│  └────┘ └────┘ └────┘ └────┘ └────┘    Title / three-line description │
└───────────────────────────────────────────────────────────────────────┘
```

The hero fills the viewport and is capped at 980 px; `--bleed` lets the facts
rail and the film rail run flush off the right edge. The grid follows directly
below, so a visitor can pick a film from the rail without scrolling and browse
the whole catalogue by scrolling or by the hero's down button.

### Sign-in / create-account anatomy

```
┌──────────────────────────────────────────────────────────────┐
│ ◘ Clover                             BACK TO THE CATALOGUE   │
│                                                              │
│  full-bleed graded still, the film's own photograph          │
│                                            ┌───────────────┐ │
│                                            │ Sign in to    │ │
│                                            │ watch         │ │
│                                            │ ───────────   │ │
│                                            │ SIGN IN  CREATE│ │
│                                            │ Email  [     ]│ │
│                                            │ Password[    ]│ │
│  HALCYON BAY                               │ [ SIGN IN  → ]│ │
│  A marine biologist comes back…            └───────────────┘ │
│  2023 · 2h 06m                                               │
└──────────────────────────────────────────────────────────────┘
```

The art is full-bleed behind a frosted form panel on the right, and the film
block sits bottom-left on it. Without a film the block shows Clover's own copy.
Below 960 px the art becomes a 250 px band above the form; the form is
left-aligned, capped at 460 px, and the panel dissolves into the page.

### Player anatomy

```
┌──────────────────────────────────────────────────────────────┐
│ (←)  HALCYON BAY                                             │
│      2023 · 2h 06m           top scrim, fades with the bar    │
│                                                              │
│                        ( ▶ )             88px white disc      │
│                                                              │
│  ┌────────────────────────────────────────────────────────┐  │
│  │ ▶  0:04 ─────●──────────── 0:14   🔊 ────  ⛶          │  │ 64 glass
│  └────────────────────────────────────────────────────────┘  │
└──────────────────────────────────────────────────────────────┘
```

The player is `position: fixed; inset: 0` with the video at `object-fit:
contain`. Chrome hides only while playing, after 2.6 s without a pointer move;
paused keeps it.

### Breakpoints

| Width | What changes |
| --- | --- |
| ≥ 1200 | 5-column grid, 26 px gap, 32 px gutter; hero capped at 980 px |
| 1100–1199 | 4-column grid, 22 px gap; facts rail hidden; art foot capped at 58 % |
| 900–1099 | 4-column grid at 24 px gutter; hero foot two columns; auth still full-bleed to 960 |
| 600–899 | 3-column grid, 20 px gap; film rail stacks to two rows, stat card above the steps, hero min 620 px; auth stacked with a 250 px band |
| < 600 | 2-column grid, 16 px gap, 20 px gutter; header links hidden, rail thumb hidden, steps compact; auth band 210 px, art description and meta hidden; volume slider hidden, mute kept; player bar 58 px |

## Components

**Header** — fixed, transparent over the hero, turning to 82 % black with a
20 px blur and a hairline once `scrollY > 16`. Brand left (22 px amber leaf mark
+ 14 px Inter Tight tracked-caps "Clover"), two links spread across the middle
("Catalogue", "How it works", 10 px, 500, `.18em`), one action right: the white
gooey "Sign in" capsule, or a 32 px white initial disc plus "Sign out". The
links and the capsule are buttons that scroll in place, so the hash router is
not disturbed.

**Gooey capsule** — see Foundations. One per screen: the header's "Sign in", the
auth submit, or the player's "Try again". The auth submit is full width with the
label and knob at opposite ends; the header and retry versions are content-width.

**Facts rail** — a frosted pill anchored to the right edge at `top: 30%`,
`right: calc(var(--gutter) - var(--bleed))` so it bleeds off the screen; two
chips of factual copy and a white down button. Hidden below 1100 px.

**Film rail** — a frosted panel `100% + --bleed` wide, left corners 26 px, right
edge flush; two `<a>` entries, each a 56 px poster thumb, a 48 px year chip, an
uppercase title, a two-line caption and a 40 px white arrow disc, with a hairline
between them. Stacks to two rows below 900 px.

**Stat card** — white, radius 22, `#111` on `--paper`; two rows of a 30 px Inter
Tight numeral and an 11.5 px caption with a hairline between, announced as
"Clover at a glance".

**Step card** — glass, radius 22, min-height 156 px; a circled 34 px icon at the
top, a 10 px tracked-caps label and an 11.5 px caption. Three across, then one
compact row each below 600 px.

**Hero icon button** — a 46 px glass square, radius 15, down arrow; hovers 2 px
down; its `aria-label` names the action.

**Auth panel** — the form sits in a 26 px frosted panel: a white gradient over
`rgba(5,5,5,.34)` with `blur(30px) saturate(1.15)`. Below 960 px the panel
dissolves into the plain stacked form.

**Poster card** — one `<a>` wrapping a 2:3 poster, an uppercase title and a
three-line description. Radius 20 on the art, an inset 1 px white 6 % hairline,
and on hover a 2 px lift, a warm shadow and a 1.035 image scale over 250/500 ms.
Focus uses the gold ring on the link.

**Mode tabs** — two toggle buttons with `aria-pressed`, 11.5 px tracked caps,
a shared 12 % hairline beneath, and a 2 px `--gold` underline on the active one.
They change copy in place and never reload the page.

**Field** — 54 px, `--r-input`, `rgba(20,18,16,.72)` fill, `--glass-edge` border;
hover strengthens the hairline, focus turns it gold with a 3 px 18 % gold halo.
Invalid switches the border to `--danger-line` and the halo to red. The label is
11 px tracked caps above; the hint and the error share the 12.5 px slot below.

**Notice** — a 12 px-radius panel above the form. `err`: red text, a 42 % red
border, a 9 % red fill, alert icon, `role="alert"`. `info`: gold text, a 40 %
gold border, an 8 % gold fill, lock icon, `role="status"`.

**Player chrome** — a top scrim from 94 % black to transparent carrying the back
disc, title and meta; a bottom glass bar carrying play/pause, elapsed, scrubber,
duration, mute + volume and fullscreen. Both fade over 250 ms and pointer events
turn off while hidden.

**Big play disc** — 88 px white, centred, 18 px shadow, play glyph offset 3 px
right, hover scale 1.05. Hidden whenever the video is playing.

**Range control** — 4 px track, gold fill up to `--fill`, white 12 px thumb, 22 px
hit area, gold focus ring. Used by the scrubber and the volume slider.

**Message panel** — centred, max 900 px, one uppercase heading, one 40 ch
paragraph, then actions in a 24 px gap row. Missing film: back only. Failed
video: retry and back.

## Motion

| Where | What | Duration | Easing |
| --- | --- | --- | --- |
| Catalogue still | fade in, then drift 1.02 → 1.09 over 34 s, alternating | 700 ms / 34 s | ease / ease-in-out |
| Catalogue headline | rises out of an overflow mask, 105 % → 0 | 700 ms | `cubic-bezier(.16,1,.3,1)` |
| Catalogue lede and down button | fade up 12 px, 140 ms delay | 500 ms | same |
| Facts rail | fade up, 300 ms delay | 500 ms | same |
| Film rail | fade up, 220 ms delay | 550 ms | same |
| Stat card and step cards | fade up, 340 ms delay; steps stagger 60 ms | 550–600 ms | same |
| Catalogue cards | fade up 12 px, 28 ms stagger capped at 6, 120 ms base delay | 400 ms | same |
| Card hover | lift 2 px + warm shadow / image scale | 250 / 500 ms | same |
| Rail entry hover | glass brightens, arrow disc shifts 2 px up-right | 200 ms | same |
| Header | turns solid on scroll | 250 ms | default |
| Auth still | fade in | 500 ms | ease |
| Auth film block | fade up, 80 ms delay | 500 ms | reference ease |
| Auth form box | each child fades up, 40 ms stagger | 450 ms | same |
| Player chrome | fade out idle / in on move | 250 ms | same |
| Play disc | hover scale | 220 ms | same |
| Arrow knob | shift on hover | 200 ms | same |

One orchestrated entrance per page. The only motion that runs on its own is the
catalogue still's 34 s drift — the reference's screen is a recording, and the
drift keeps the first screen alive in that register — plus the player chrome's
idle fade while playing. Hovers are 180–250 ms and move only transform, opacity
and shadow. `prefers-reduced-motion: reduce` removes every animation and
transition, including the drift, and forces the entrance elements to their end
state, so nothing is left invisible and the layout is unchanged.

## Accessibility floor

- Focus is always visible: `2px solid var(--gold-glow)` at 3 px offset. The
  container that takes focus on route change (`#main`) paints no ring.
- Contrast on `--ink-000`: white 20:1, body text 12:1, captions 6.3:1, gold
  9.2:1, error text 10:1. Nothing informative sits below 4.5:1.
- Every page has one `<h1>` and a `#main` landmark; the skip link targets it, and
  `#main` takes focus after every route change so a keyboard or screen-reader
  user is told the page changed.
- Controls are real `<button>` and `<a>` elements. Icon-only buttons carry an
  `aria-label` that changes with state ("Play" / "Pause", "Mute" / "Unmute",
  "Enter fullscreen" / "Exit fullscreen").
- The scrubber and volume are real range inputs with labels and `aria-valuetext`
  ("0:04 of 0:14"), so they are operable by keyboard and audible to a screen
  reader.
- The player owns Space, K, ←, →, ↑, ↓, M and F only while it is mounted, and
  ignores them while focus is in a field, button or link. The listener detaches
  on route change.
- The rail entries are real `<a>` links named by their film title and caption;
  the facts rail and stat card are labelled `<aside>` regions, and the down
  buttons carry `aria-label="See all 13 films"`. In-page scrolling moves focus to
  the target, so the next Tab continues from where the page jumped to.
- Posters carry `alt="Poster for <title>"`; every decorative still is `alt=""`
  because its title is adjacent text.
- `prefers-reduced-motion: reduce` removes all motion without changing layout.
- No modal exists in v1, so there is no focus trap to get wrong: auth is a page.

## Assets

| Asset | Path | Notes |
| --- | --- | --- |
| Inter 400 / 500 / 600 | `ui/mock/assets/fonts/inter-*.woff2` | SIL OFL 1.1, self-hosted |
| Inter Tight 700 | `ui/mock/assets/fonts/inter-tight-700.woff2` | SIL OFL 1.1, display |
| Stills ×3 | `ui/mock/assets/img/still-*.jpg` | 1920 × 1080 Unsplash placeholders, reused across films |
| Posters ×13 | `ui/mock/assets/img/p-*.jpg` | 480 × 720 Unsplash placeholders |
| Sample clip | `ui/mock/assets/video/clover-sample.mp4` | 14 s, 1280 × 720, h264 + AAC, generated from `still-halcyon`; every film plays it in the mock |
| Mark and icons | inline `<symbol>` defs in `index.html` | Clover mark plus play, pause, chevrons, volume, mute, fullscreen, alert, lock; `currentColor` |

All imagery is placeholder and must be replaced with Clover's licensed key art
and stills before launch; the grade is applied in CSS, so replacements do not
need re-grading. There are no external requests: fonts, images, video and icons
are local files.

## Handing this to the Go build

The interface is plain HTML, CSS and JavaScript served by a Go standard-library
server — no framework, no build step, no nexcommon or nexcomponent.

- One shell document with the icon symbols, `app.css` and `app.js`; the same
  three routes the mock uses (`/`, `/auth`, `/play/:id`), hash-based as here or
  server paths that return the shell.
- Keep the token names and values in `:root`; they are the contract.
- Keep the three behaviours that are easy to lose in a rewrite:
  1. the fixed z-order of the still's three layers,
  2. focus moving to `#main` after each route change,
  3. the player claiming its keys only while mounted.
- The player stays a real `<video>` with native controls off; `http.ServeContent`
  provides range requests, which is all seeking needs. Do not rebuild playback
  on a canvas.
- The five player states and six auth states come from real conditions (no
  session, expired session, unknown film, media error, bad credentials, duplicate
  email, invalid input) — not from the mock's `?playing=1`, `?demo=1`,
  `err=…` preview flags, which never ship.

---

## Page inventory

| Screen | Route | States |
| --- | --- | --- |
| Catalogue | `#/` | signed out, signed in |
| Sign in / Create account | `#/auth?next=<film>` or `#/auth` | sign in, create account, bad credentials, invalid input, duplicate email, expired session, with film, without film |
| Player | `#/play/<film>` | paused, playing, chrome idle, video failed, film missing |

That is the whole product. Every state is directly addressable; the full list
with routes is `ui/mock/screens.json`, and the screen contents are in
`UIMap.md`.

## Preview contract

`/ui-preview` reads two things from this folder: the `## Pages` headings in
`UIMap.md` (the page list) and `ui/mock/screens.json` (the screens under each
page). A screen joins a page only when its `page` equals the heading. Opening a
screen sets the mock's hash to its `id`, so the mock routes that id to the
state — `#catalogue-signed-in`, `#auth-signup-invalid`, `#player-idle` and the
rest all open directly. The preview's inner page allows no media, so inside it
the player shows the poster still with a working clock instead of the clip.

## What changed from the previous draft

| Previous draft | Now |
| --- | --- |
| Browse with a rotating hero carousel, a featured rail, a stat card and feature cards | Catalogue with the reference's full-viewport hero, its rails, a stat card and step cards rebuilt on Clover facts, and the full grid below |
| Film detail page with chips, synopsis and "more like this" | Removed — selecting a film goes to the player, as REQ-003 says |
| Sign-in as a modal gate over the detail page | A full account page with two mode tabs and the film held on the art |
| Search, Genres, Genre, My list, New on Clover | Removed — all outside v1 and forbidden by TC-003 |
| Player as a still image with a fixed clock | A real local MP4 with working play, pause, seek, volume and fullscreen |
| React handoff with ten components | Plain HTML, CSS and JS a Go standard-library server can serve; no framework, no build step |
| Review rounds scoring the carousel and the sheet | Two fresh rounds on the v1 screens in `ui/review/` |

## Revision after the manager's review, 2026-11-13

The manager reviewed the first v1 delivery against the same recording and
rejected its composition: the catalogue opened on a 42 vh masthead rather than
the reference's full screen, so the rails, the white stat card and the glass
feature cards — the parts that make the reference recognisable — were missing or
reduced to the player. The mock and these documents were rebuilt:

- The catalogue hero now fills the viewport (capped at 980 px), with the spread
  nav and the gooey "Sign in" pill, the display line at ~92 % of the column, the
  lede with its frosted down button, the facts rail bleeding off the right edge,
  a two-entry glass film rail, the white stat card and the three glass step
  cards. The grid and its 13 films stay, per the manager's requirement.
- The account screen took the same material language: the film's still
  full-bleed, the film title bottom-left, the form in a frosted panel on the
  right, stacked below 960 px.
- The reference's content, copy, imagery and social rail were not copied; the
  rails and cards carry Clover facts and real catalogue entries.
- `UIMap.md`, this file, `ui/mock/screens.json` and `ui/review/round-2.md` were
  updated to describe what the mock now renders. The verification was a
  self-check against the changed screens, not a third review round.
