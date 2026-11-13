# Clover — design system (FEAT-001)

The contract for Clover's interface. `ui/mock/` is the working reference
implementation of everything below; when this file and the mock disagree, the
mock wins and this file gets fixed. The reading of the manager's video is in
`DesignBrief.md`.

## How the system is meant to be used

Three rules carry most of the look:

1. **One photograph at a time.** A page either has a full-bleed graded still
   (Browse, film detail, sign-in) or none at all. Never two competing images.
2. **One white surface per page.** The stat card on Browse, the primary
   capsule everywhere. Everything else is black, glass, or hairline.
3. **Amber is light, not paint.** It lives in the grade over the still, the
   progress fill, and the focus ring. No amber buttons, no amber text blocks.

## Foundations

### Space

| Token | Value | Use |
| --- | --- | --- |
| `--gutter` | 32 px (24 px ≤ 1279, 20 px ≤ 767) | Page inset, header inset, rail bleed |
| `--nav` | 72 px (64 px ≤ 767) | Header height, page top padding |
| Grid | 12 columns, 24 px gap | Hero foot only |
| Section | 88 px top, 96 px bottom | Poster row stack |
| Row stack | 56 px | Between poster rows |
| Card padding | 22–24 px | Stat card, feature cards |

Width is capped at 1680 px and centred; the hero and its rail bleed edge to
edge regardless.

### Radius

| Token | Value | Use |
| --- | --- | --- |
| `--r-card` | 28 px | Stat card, feature cards, genre cards, dialog |
| `--r-poster` | 20 px | Posters, search bar, film poster |
| `--r-thumb` | 16 px | Rail thumbnail, inputs, square icon buttons |
| — | 26 px | Rail letter chip (the one deliberate oddity) |
| — | 24 px | Player control bar |
| — | 999 px | Pills, circular arrows, capsules |

Radii never go below 16 px. There are no square corners anywhere in Clover.

### Elevation and glass

Two glass recipes, both `backdrop-filter: blur(28px) saturate(1.06–1.08)`:

- **Warm rail** — `linear-gradient(180deg, rgba(255,255,255,.10), rgba(255,255,255,.035))`
  with a 1 px `--glass-edge` top and bottom hairline and an inset top highlight.
- **Dark card** — `linear-gradient(180deg, rgba(46,36,28,.52), rgba(20,18,16,.38))`
  with a 1 px `--glass-edge` border.

The white card and the dialog are opaque — glass is never stacked on glass.

### Type scale

| Role | Size | Weight / family | Notes |
| --- | --- | --- | --- |
| Hero display | `clamp(46px, 11.4vw, 164px)` | 700 Inter Tight | `line-height .92`, `-.028em`, uppercase, one line |
| Detail title | `clamp(40px, 4.6vw, 68px)` | 700 Inter Tight | `-.025em`, uppercase |
| Auth title | `clamp(36px, 4.2vw, 58px)` | 700 Inter Tight | `-.025em`, sentence case |
| Page title | `clamp(36px, 4.4vw, 60px)` | 700 Inter Tight | Browse-adjacent pages |
| Row title | 30 px (26 ≤ 899, 24 ≤ 599) | 700 Inter Tight | |
| Rail title | 30 px | 700 Inter Tight | uppercase, up to two lines |
| Card title | 26 px | 700 Inter Tight | Genre cards |
| Dialog title | `clamp(28px, 3vw, 38px)` | 700 Inter Tight | |
| Lede | 16–17 px | 400 Inter | max 44–52ch, `--text` |
| Body | 14–16 px | 400 Inter | `--text` |
| Poster caption | 14 px / 12 px | 500 / 400 Inter | title / `year · genre` |
| Label | 11–12 px | 600 Inter | uppercase, `0.08–0.10em` |
| Numeral | 40 px | 700 Inter Tight | tabular, `-.02em` |

The hero display is the only element that is *fitted*: `fitTitle()` in
`app.js` resets the size to the clamp, measures the rendered line, and sets
the size so the line fills 98 % of the column — growing a short title and
shrinking a long one. This is what keeps every film title at the reference's
near-full-width scale without wrapping.

### Colour

| Token | Value | Semantics |
| --- | --- | --- |
| `--ink-000` | `#050505` | Page |
| `--ink-100` | `#0b0b0c` | Raised black |
| `--ink-200` | `#141210` | Inputs, dialog, player chrome |
| `--paper` | `#ffffff` | White surfaces and primary CTA |
| `--ink-on-paper` | `#111111` | Type on white |
| `--paper-muted` | `#6e6e6e` | Labels on white |
| `--paper-rule` | `#e6e4e1` | Rules on white |
| `--gold` | `#e0a34e` | Progress, saved state |
| `--gold-deep` | `#8b4d11` | Amber grade base |
| `--gold-glow` | `#ffc46b` | Focus ring |
| `--text` | `rgba(255,255,255,.78)` | Body on black |
| `--text-dim` | `rgba(255,255,255,.56)` | Metadata |
| `--glass-edge` | `rgba(255,255,255,.16)` | Glass hairline |
| `--glass-edge-strong` | `rgba(255,255,255,.45)` | Hover hairline |
| `--danger` | `#ff9b8b` | Error text |
| `--danger-line` | `#e5604d` | Invalid field border |

### The still grade

Every full-bleed still gets the same three layers, in this order (z-index in
brackets): the image at `brightness(.82) saturate(.88) contrast(1.05)` [1], a
screen-blended amber gradient rising from the bottom
`rgba(214,124,32,.62) → rgba(168,86,8,.30) 42% → transparent 76%` [2], and a
black scrim — top-to-bottom from `--ink-000`, left-to-right for legibility,
and again from the bottom [3]. This is what makes thirteen different stock
photographs read as one film. Any new still must go through it.

## Layout

### Browse anatomy

```
┌──────────────────────────────────────────────────────────┐
│ logo        Browse Genres New My list   ⌕   [Sign in →]  │ 72
├──────────────────────────────────────────────────────────┤
│                                                          │
│  ▬ ▬ ▬                                                   │
│  HALCYON BAY                                  ← fitted   │
│  A marine biologist comes back…                          │
│  [Watch now →] (+)                                       │
│                                                          │
│ ┌───────────────────────┬───────────────────────┐        │ full-bleed
│ │ [▣][M] MERIDIAN LINE →│ [▣][P] PAPER LANTERNS→│        │ rail
│ └───────────────────────┴───────────────────────┘        │
│ ┌───────┐  ┌──────┐ ┌──────┐ ┌──────┐                    │
│ │ 8,400 │  │Genres│ │List  │ │Search│                    │
│ │ 92 %  │  └──────┘ └──────┘ └──────┘                    │
│ └───────┘                                                │
└──────────────────────────────────────────────────────────┘
   Continue watching ─────────────────────────────── see all
   [poster] [poster] [poster] [poster] …
```

The hero is `min-height: clamp(680px, 100vh, 940px)`; the rail is pushed to
the bottom with `margin-top: auto`, so the composition holds at any viewport
height. Content is left-aligned throughout — nothing is centred except
circular icons and the dialog.

### Breakpoints

| Width | What changes |
| --- | --- |
| ≥ 1200 | Full composition: 4-column stat block, 6-column feature block, two-up rail |
| 768–1199 | Rail stacks to one column with a top hairline; hero foot becomes one column; gutter 24 |
| ≤ 899 | Nav links and the header CTA collapse into a full-screen sheet behind a menu button; genre cards go 2-up |
| ≤ 599 | Gutter 20; posters 132 px; grids 2-up; rail entries wrap their description under the thumb pair. The still is anchored to the top of the viewport instead of stretching the whole hero, with a stronger amber grade and top scrim, because a 16:9 photograph cropped to a phone column otherwise lands on its darkest part |

## Components

**Header** — fixed, transparent on Browse until `scrollY > 8`, then solid
`rgba(5,5,5,.86)` with blur and 64 px height. The hairline starts at the left
gutter and runs off the right edge. Links: 12 px, 600, `0.10em`, uppercase,
`--text` → white on hover/current.

**Gooey capsule** — the only primary action style. Structure: a background
layer carrying `filter: url(#goo)` (blur 9, alpha matrix, composite atop) with
two white shapes — a pill body and a 48 px (56 px large) circular knob — and
an unfiltered foreground layer holding the label and the arrow. The filter is
what produces the fillet where the knob meets the body. Label 12–13 px, 700,
`0.10em`, uppercase, `#111`. On hover the arrow moves 2 px up-right.

**Square icon button** — 44 × 44, radius 16, `--glass-edge` border,
`rgba(255,255,255,.04)`. `.on` turns it `--gold` with dark type (the saved
state). Circular variant is 38 px.

**Poster card** — 180 px wide (132 px ≤ 599), 2:3, radius 20, caption below.
Hover lifts 2 px with a warm shadow and scales the image to 1.04. The progress
bar is a 3 px strip pinned to the bottom, `--gold` on `rgba(255,255,255,.20)`.

**Rail item** — 64 px thumb (radius 16) + 64 px letter chip (radius 26,
overlapping the thumb by 14 px) + title/description + a 38 px white circular
arrow. Two per row; a hairline separates them.

**Stat card** — white, radius 28, 340–420 px wide, two rows split by
`--paper-rule`. Numerals count up over 1.2 s with a cubic ease-out; the `%` is
22 px.

**Feature card** — 190 px tall, dark glass, 46 px outlined icon circle at the
top, label and one-line description pinned to the bottom. Links, not buttons.

**Chips** — 32 px tall, radius 16, `--glass-edge`, 11 px uppercase. The rating
chip carries a 12 px play glyph.

**Fields** — 56 px tall, radius 16, `rgba(20,18,16,.72)` fill, 1 px
`--glass-edge`. Focus is a 2 px `--gold-glow` ring with a 2 px offset. Invalid
adds a `--danger-line` border and a 12 px message with a 12 px glyph.

**Dialog (sign-in gate)** — 440 px, `rgba(20,18,16,.94)`, radius 28, over a
`rgba(5,5,5,.72)` + 10 px blur backdrop. Close button top-right.

**Player** — full-bleed 16:9 frame; a top bar with back, title and metadata
over a black-to-transparent gradient; a floating control bar 16 px from the
edges (24 px radius, glass) with play/pause, elapsed, scrub, duration,
subtitles, full screen. Chrome fades out after 2.6 s of no pointer movement.

**Footer** — hairline, copyright left, three 11 px links right.

## Motion

| Where | What | Duration / easing |
| --- | --- | --- |
| Hero title | rises out of an overflow mask | 600 ms `cubic-bezier(.16,1,.3,1)` |
| Logline / actions | fade up 10 px | 450 ms, delayed 120 / 220 ms |
| Rail items | fade in from blur(10px) | 700 ms, delayed 100 / 220 ms |
| Stat card | fade only (counters handle the rest) | 400 ms, delayed 550 ms |
| Feature cards | fade up | 600 ms, delayed 300 / 380 / 460 ms |
| Slide change | all of the above replay | on click or every 9 s |
| Hovers | transform / background | 180–220 ms |
| Poster hover | lift 2 px, image scale 1.04 | 220 ms |

Stagger never exceeds 120 ms between siblings and nothing animates on its own
after the entrance except the carousel and the counters.

## Accessibility floor

- Focus is always visible: `2px solid var(--gold-glow)`, 2 px offset.
- Body text is ≥ 12.2:1 on black, metadata ≥ 6.5:1; labels on white are 5.1:1; error text is
  9.2:1 on `--ink-200`.
- The skip link jumps to `#main`, which every page provides. The page takes
  focus on every route change (`tabindex="-1"`, no ring), so the skip link and
  a keyboard user both land at the start of the new content.
- The mobile sheet has a close control, closes on Escape, returns focus to the
  menu button, and is `inert` when closed with the page scroll locked while it
  is open. The gate and the sheet still need a focus trap in the React build
  (the mock does not trap focus).
- Images are decorative where a title sits next to them (`alt=""`); the film
  poster carries `alt="Poster for <title>"`.
- `prefers-reduced-motion: reduce` removes every entrance, the carousel and
  the counters — and leaves the layout identical.
- Controls are real `<button>`/`<a>` elements with `aria-label`s; the
  carousel dashes expose the featured film's name.

## Assets

| Asset | Path | Notes |
| --- | --- | --- |
| Inter 400 / 500 / 600 | `ui/mock/assets/fonts/inter-*.woff2` | SIL OFL 1.1, self-hosted |
| Inter Tight 700 | `ui/mock/assets/fonts/inter-tight-700.woff2` | SIL OFL 1.1, display |
| Stills ×3 | `ui/mock/assets/img/still-*.jpg` | 1920 × 1080 Unsplash placeholders |
| Posters ×13 | `ui/mock/assets/img/p-*.jpg` | 480 × 720 Unsplash placeholders |
| Mark | inline `<symbol id="i-clover">` | Four leaves + stem, 1.3 px stroke |

All imagery is placeholder under the Unsplash licence and must be replaced
with Clover's licensed key art and stills before launch. The grade in
"Foundations" is applied in CSS, so replacements do not need re-grading.

## Handing this to `clover-fe`

The React build keeps the token names in `:root`, the same numbers, and the
same component boundaries (`Header`, `Hero`, `Rail`, `StatCard`,
`FeatureCard`, `PosterRow`, `PosterCard`, `Gate`, `Player`). Two behaviours
are easy to lose in a rewrite and must not be:

1. the hero title's fit-to-width measurement (grow and shrink, 98 %),
2. the fixed order of the three still layers, and
3. focus moving to `#main` after each route change.

The mock has no build step and no dependencies; it is a design artefact, not
the application.
