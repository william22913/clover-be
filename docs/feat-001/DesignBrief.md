---
feature_code: feat-001
service: clover
spec_version: 1
source: UIMap.md
last_updated: 2026-11-14
---

# Design Brief — feat-001 (clover)

The short, readable digest of `UIMap.md`. The map is the source of truth; this file never holds a
value the map does not. The design itself is the clickable mock: `ui/mock/index.html#<screen id>`.

## Overview

Clover is a dark, warm, photo-led film catalogue. One full-bleed golden photograph carries the
opening screen: a heavy uppercase headline, one short paragraph, a paper-white facts card and glass
pills, on near-black. The viewer browses all 20 films without signing in; picking one asks for the
seeded account and then lands on that film's player. There is no Figma file — the design was built
from the manager's Dribbble reference (`ui/mock/screens.json` records it), and the mock is the design.

Pages:

| Page | Route | Screens |
|---|---|---|
| Catalogue | `/` | `#catalogue`, `#catalogue-signed-in`, `#catalogue-loading`, `#catalogue-empty`, `#catalogue-poster-error` |
| Sign in | `/sign-in` | `#signin`, `#signin-direct`, `#signin-required`, `#signin-error`, `#signin-loading` |
| Player | `/watch/{movieId}` | `#player`, `#player-loading`, `#player-error` |
| Film not found | `/watch/{movieId}` | `#not-found` |

## What it looks like

A full-screen photograph of golden liquid in a dark room opens the catalogue. On top of it: a line of
small capitals, one huge headline in tight uppercase, a short paragraph, a cream-white card carrying
two facts ("20" films and "Free"), and a soft frosted pill that says "Browse the catalogue". Below,
the twenty films are rounded cards with their poster, a number, a title, one line about the film and
a small round arrow. Everything is warm black and amber; the strongest colour is the photograph
itself. Buttons are pills, the main one white with a bold black uppercase label, a small up-right
arrow, and a wavy right edge — a round lobe at the far right with a concave bite between it and the
body. When something arrives it eases into place in about half a second, never on a slow cue.

Sign in, the player and the "film not found" page are screens we drew in that same language, because
the reference only shows the catalogue; they are the ones that most need your eye in the preview.

The tables below hold the same information written for the people who build it.

## Brand Identity

Cinematic, warm and quiet. The mood is set by the photography — amber light in a dark room — and by
one restrained accent. Type is a tight, heavy geometric grotesque in uppercase for headlines, small
wide-tracked uppercase for labels and buttons, and low-contrast warm grey for body copy. Surfaces are
either frosted glass over the photo or one paper-white card that carries the facts. Corners are
generous (18–34 px), buttons are pills, and the primary button ends in a wavy right edge — a round
lobe at the far right with a concave bite against the body — the reference's signature detail.

## Color Palette

| Token | Value | Used for |
|---|---|---|
| `--c-ink` | `#060504` | Page background |
| `--c-ink-soft` | `#0c0a08` | Card media bed, inputs |
| `--c-paper` | `#f6f1e7` | Facts card, primary button, arrow discs |
| `--c-paper-hover` | `#ffffff` | Hover on paper surfaces |
| `--c-paper-deep` | `#e9e1d3` | Paper gradient's far end |
| `--c-paper-ink` | `#17120c` | Text on paper |
| `--c-paper-muted` | `#5b5245` | Labels on paper |
| `--c-text` | `#f7f2e9` | Primary text |
| `--c-text-muted` | `#c9bda9` | Paragraphs, card descriptions |
| `--c-text-faint` | `#a99d8a` | Meta lines, footer |
| `--c-accent` | `#e8a24b` | Text selection, now-playing dot, played timeline |
| `--c-accent-bright` | `#f3b963` | Eyebrows, hovered arrow disc, timeline fill end, spinner, toast icon |
| `--c-glass` | `rgba(255,238,216,.10)` | Glass cards and panels |
| `--c-glass-strong` | `rgba(255,238,216,.17)` | Glass hover, ghost button |
| `--c-glass-deep` | `rgba(18,14,10,.55)` | Control bar, index badge |
| `--c-header` | `rgba(6,5,4,.72)` | Sticky header |
| `--c-line` | `rgba(255,245,232,.16)` | Hairline borders |
| `--c-line-strong` | `rgba(255,245,232,.30)` | Hover borders |
| `--c-line-paper` | `rgba(20,16,11,.16)` | Dividers on paper |
| `--c-shadow` | `rgba(0,0,0,.55)` | Shadow base |
| `--c-scrim-top` / `-mid` / `-bottom` | `rgba(6,5,4,.78)` / `.30` / `.94` | Photo scrim, top to bottom |
| `--c-scrim-mobile` | `rgba(6,5,4,.60)` | Extra veil under the hero and sign-in scrims below 620 px |
| `--c-overlay` | `rgba(6,5,4,.42)` | Stage overlay |
| `--c-input` | `rgba(6,5,4,.38)` | Field background |
| `--c-skeleton` | `rgba(255,238,216,.08)` | Loading blocks |
| `--c-progress-track` | `rgba(255,245,232,.24)` | Timeline and volume tracks |
| `--c-toast` | `rgba(18,14,10,.88)` | Toast background |
| `--c-focus` | `#f3b963` | Focus ring on dark and over photos |
| `--c-focus-paper` | `#17120c` | Focus ring on paper |
| `--c-danger` | `#ffb49b` | Field error text, the alert's left rule |
| `--c-danger-text` | `#ffd9cc` | The alert's message |
| `--c-danger-line` | `rgba(255,150,120,.45)` | Border of a field in error and of the player's error card |

One accent: amber, used on the focus ring, the toast icon and the hovered arrow disc. Everything
else is warm neutral. Gradients are tokens too: `--g-hero-scrim`, `--g-glass`, `--g-glass-hover`,
`--g-glass-panel`, `--g-paper`, `--g-image-scrim`, `--g-stage-scrim`, `--g-page-glow`,
`--g-skeleton`, `--g-page-scrim` (the denser scrim the Player and Film-not-found carry at every
width).

## Typography

**Instrument Sans**, 400 / 500 / 700. The reference's font cannot be identified from the video;
this is the closest open match for its tight heavy grotesque. Two files ship, Regular and Bold; the
Regular face also serves the 500 labels.

| Role | Token | Size (1440 px) |
|---|---|---|
| Catalogue and Film-not-found headline | `--fs-hero` `clamp(3rem, 6.6vw, 6.75rem)` | 95 px |
| Player title | `--fs-title` `clamp(2rem, 4.2vw, 3.9rem)` | 60.5 px |
| Section / empty-state title | `--fs-display` `clamp(2rem, 3.4vw, 3rem)` | 48 px |
| Sign-in title | `--fs-panel-title` `clamp(1.9rem, 3vw, 2.6rem)` | 41.6 px |
| Film title, error-card title | `--fs-card` `1.25rem` | 20 px |
| Hero paragraph | `--fs-lede` `1.0625rem` | 17 px |
| Body, inputs | `--fs-body` `0.9375rem` | 15 px |
| Card description, footer | `--fs-small` `0.8125rem` | 13 px |
| Labels, timecodes, index badge | `--fs-tiny` `0.6875rem` | 11 px |
| Eyebrow | `--fs-eyebrow` `0.6875rem` | 11 px |
| Header links | `--fs-nav` `0.75rem` | 12 px |
| Buttons | `--fs-btn` `0.75rem` | 12 px |

Tracking: `--ls-tight −0.03em` (headlines), `--ls-snug −0.01em` (card titles, small headings),
`--ls-label 0.08em` (tiny uppercase labels), `--ls-wide 0.14em` (nav, buttons, badges),
`--ls-wider 0.18em` (eyebrows). Line height: `--lh-tight 0.9` (hero), `--lh-snug 1.12`,
`--lh-normal 1.55` (body).

## Spacing and Shape

Spacing scale `--s-1` … `--s-9` = 4, 8, 12, 16, 24, 32, 48, 64, 96 px; page gutter
`--page-x clamp(20px, 4vw, 48px)`.

Radii: `--r-xs 8px`, `--r-sm 12px`, `--r-md 18px`, `--r-lg 26px`, `--r-xl 34px`, `--r-pill 999px`,
`--r-round 50%` (discs). The primary button's notched right edge is drawn, not a radius: a 44 × 44 px
tail with one concave bite cut against the body and a blunt tip.

Blur: `--blur-glass 18px` (cards, chips), `--blur-panel 28px` (control bar, toast, header),
`--blur-photo 8px` (blurred photo fallbacks).

Elevation: `--shadow-card 0 24px 60px rgba(0,0,0,.45)` (sign-in panel, stage, empty panel, toast),
`--shadow-paper 0 22px 48px rgba(0,0,0,.38)` (facts card, play disc), `--shadow-lift 0 30px 80px
rgba(0,0,0,.58)` on card hover only. Two distinct shadows are visible in the swept states.

Breakpoints: 1180, 1020, 820, 620, 380 px. Grid 4 → 3 → 2 → 1 columns; the hero foot stacks at
1020 px; header links hide at 820 px, the brand wordmark at 380 px.

## Components

**No nexcomponent export is used** — the manager's constraint is a separate browser frontend without
nexcomponent, and there is no Figma file to resolve `Code` names from. Every piece below is defined
once in `ui/mock/styles.css` and every screen is built from these pieces:

Brand mark and wordmark · sticky header · primary capsule button (white fill, uppercase label,
up-right arrow, lobed right edge) · ghost pill button · quiet text button · icon-button disc · centre
play disc · arrow disc · glass film card (poster, index badge, title, description, arrow) · paper
facts card · browse chip · field (label, glass input, error line) · form alert · target chip · player
stage (poster bed, scrim, control bar) · timeline and volume sliders · toast · poster fallback ·
skeleton card · spinner · empty panel.

## States and Motion

States per page are listed in the map. The set the design covers: catalogue default, signed in,
loading (skeletons), empty, poster failed; sign-in for a film, direct, fields required, wrong
credentials, submitting; player ready, loading, video failed; film not found. Every state is shown
as its own mock screen and listed in `ui/mock/screens.json`.

The primary button is a white capsule with a 12 px bold uppercase black label, a 14 px up-right arrow
at its right end, and a wavy right edge — a rounded lobe at the far right with a concave bite between
it and the body. Hover lightens the fill and nudges the arrow up-right, press scales to 0.97, and
focus draws a ring around the whole silhouette. It is the only primary action on every screen; the
ghost button carries no arrow.

Motion durations: `--dur-fast 140ms` (hover tints, press, arrow), `--dur-base 240ms` (poster zoom,
alert, play disc), `--dur-slow 520ms` (screen entrance). Easings: `--ease-out
cubic-bezier(.22,1,.36,1)`, `--ease-in cubic-bezier(.42,0,1,1)`, `--ease-linear cubic-bezier(0,0,1,1)`.

The entrance is measured from the reference and deliberately quick:

- **One beat, on every arrival:** the blocks of a screen rise together over 520 ms on **ease-in**,
  with an 18 px rise, blur from 2 px to none, and the regions starting 0–180 ms apart — about 700 ms
  end to end. The photo or poster itself holds still.
- **No second beat and no waiting:** reference beat 02's shape is used as the single entrance; the
  2.8 s cue it plays at in the reference is not reproduced, and the shot's own 1500 ms linear camera
  move is not built into the product.
- **A screen change is a cut:** the leaving screen is removed in one frame, and only the arriving
  screen animates, so the entrance never replays while the user moves around.

The rest is the product's motion: film cards rise in 25 ms apart; cards lift 6 px and their arrow
disc turns amber and rotates 45° on hover; every button presses to 0.97; alerts rise in; skeletons
shimmer; the toast appears for 4 s. `prefers-reduced-motion: reduce` stops all of it.

`motion-diff.mjs` against the reference is honest rather than all-green: the mock's screen change is
`OK` against reference beat 02 (600 ms against 500 ms, same ease-in), and the tool pairs the first
paint with reference beat 01 — the shot's camera move — reporting it weaker on duration by design.
Report: `ui/review/motion-diff.txt`.

## Open design gaps

- The reference shows **one screen**; Sign in, Player and Film not found are drawn in its language
  and need the manager's confirmation in `/ui-preview`.
- The **video** is not in the mock — the stage shows the film's poster with working controls; the
  real stream is implementation work.
- The reference's **social stack, close disc and bottom tile row** are deliberately not carried over.
- **Responsive behaviour is ours** (the reference is a desktop shot) and needs a look at 390 px.
- **Copy** is final English product copy; the hero headline and the empty-state wording can change.
