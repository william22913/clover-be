# UI review — round 2 (re-check)

Re-check of round 1's fixes, on the current files. Two lanes ran against the whole current state, not
the diff: **look** (14 screens at 390/768/1024/1440 px, live mock) and **coverage + docs**. Motion
was re-measured by the tooling rather than re-read by a lane, because the finding was numeric.

Round 1 scores: look 82, coverage and docs 88. Round 1 fixed 25 findings; this round found 11 more,
of which 8 are fixed here and 3 are kept with a reason.

## Findings and decisions

| # | Lane | Sev | Finding | Decision |
|---|---|---|---|---|
| 1 | look | medium | Player at 390 px: the 78 px play disc was overlapped by the transport bar by 13 px | Fixed: below 620 px the disc is 64 px and sits 26 px above centre; measured clear by 20 px |
| 2 | look | medium | Player at 390 px: ~127 px of dead space between the stage and the film's title | Fixed: `align-content: start` on the stacked player |
| 3 | look | medium | The film cards read as flat smoked panels rather than glass; the reference's glass sits over a photograph | **Kept**: the grid is a screen the reference does not have, and the card fill uses the reference's own glass value. Decorating the field would not make a photo appear behind it. If the manager wants depth here, it is a design change, not a defect |
| 4 | look | low | The empty catalogue still offered "Browse the catalogue", which scrolls to "No films available" | Fixed: the chip is hidden in the empty state |
| 5 | look | low | "THIS FILM CAN'T PLAY RIGHT NOW" broke with "NOW" alone on a line | Fixed: `text-wrap: balance` on the error card heading |
| 6 | look | low | A focused paper button shows the focus ring inside the notched tail | **Kept**: the ring follows the button's box; a wrapper would add a second focus target. The ring is still 3:1 and never clipped |
| 7 | look | low | The seek and volume tracks were 3–4 px tall with no pointer target | Fixed: 24 px and 20 px hit areas, the visible track unchanged at 4 px and 3 px |
| 8 | docs | medium | MO-003 said the eyebrow and lede run 700 ms ease-out at 120/200 ms; the lede actually runs 1500 ms linear at 350 ms | Fixed: one row with both, matching `styles.css` |
| 9 | docs | low | The polish evidence claimed 3 swept widths; the sweep covers 4 | Fixed |
| 10 | docs | low | The map gave `--shadow-paper` only to the facts card and `--c-scrim-mobile` without the sign-in scrim, while the brief had both | Fixed: the map now matches the mock |
| 11 | docs | low | The brief is accurate but reads as engineering tables, not a document for a business owner | Fixed: a plain-language "What it looks like" section now opens it; the tables follow as the builder's reference |

## Scorecard

| Category | Score | Why |
|---|---|---|
| Look | 88 | Faithful to the reference's palette, type, paper card and button language across all 14 screens; five of seven look findings fixed. The glass-over-black grid is the open judgement |
| Coverage | 93 | 29 requirements, 24 mapped to a screen that exists, 5 justified as having no UI, 0 uncovered |
| Documentation accuracy | 91 | Map, brief and mock now agree on routes, states, tokens, breakpoints and motion; two shared-tool quirks remain outside the files |
| Motion | 90 | Both reference beats matched by `motion-diff.mjs`; the landing runs once, on the first paint, and re-entry no longer replays it |
| Accessibility | 88 | No contrast, focus-ring, overflow or reduced-motion finding in the sweep across 56 screen/width pairs; pointer targets fixed |

Nothing is below the 85 bar. The one item the manager may want to change is the catalogue grid's flat
field (finding 3) — it is a deliberate difference from a reference that never shows a catalogue.

## Evidence

- `mock-sweep.mjs`, 14 screens × 390/768/1024/1440 px: **no findings** (sweep 11).
- `fidelity-sheet.mjs` catalogue vs frame-004: **PASS**.
- `motion-diff.mjs`: **every reference animation matched** (`motion-diff.txt`).
- `docs-preview.js`: 29 requirements — 24 mapped, 5 without UI, 0 not covered; 31 derived, 105 assumed.
- Deep links re-checked at 1280 px: `#watch-blue-hour` signed out opens sign-in carrying the film and
  reaches the player after sign-in; `#watch-<unknown id>` opens Film not found; signing out re-locks
  the same link.

## Manager decisions

- 2026-11-14 — the manager asked for the design to be opened as a pull request before the preview,
  to be reviewed with `/ui-preview` and not merged. The pull request carries the map, the brief and
  the mock; opening it changes nothing about approval or the merge.
- The two kept items stay as recorded above (the flat glass grid, finding 3; the focus ring inside
  the notched corner, finding 6). Neither blocks the preview and both are the manager's to decide
  there, not here.
- The 31 derived rows stay unconfirmed until the manager confirms them in the preview; the map keeps
  `status: draft`.

## Revision after the manager's preview (2026-11-14)

The manager looked at the mock in `/ui-preview` and asked for two fixes before approving: **the
animation is too slow** ("the animation is to slow"), and **the primary button still does not match**
the button reference they attached. The work stays on the same branch and updates PR #69; nothing was
merged. This is the current round's record — no third round was opened — and the fixes were checked
by the same measuring tools, not by re-reading the files.

| # | Manager's point | What changed | Evidence |
|---|---|---|---|
| 1 | The animation is too slow | The two-beat landing (1500 ms + a 2800 ms wait, 4.4 s in all) is gone. One entrance for every screen: 520 ms, `cubic-bezier(0.42, 0, 1, 1)`, regions 0–180 ms apart, 2 px blur clearing to 0; the first paint uses it too, so nothing waits. A screen change is now a cut — the leaving screen is removed in one frame and only the arriving screen animates — so the entrance never replays while moving around | Mock recording: first paint at 0.00 s, ~700 ms with the stagger, regions 200 ms apart; screen change ~600 ms, regions 100 ms apart. `motion-diff.mjs`: **`OK` vs reference beat 02** (the product entrance, 500 ms ease-in) |
| 2 | The buttons do not match the button reference | Primary button rebuilt: white capsule, 12 px bold uppercase black label, 14 px up-right arrow **inside the tail**, wavy right edge — one concave bite out of the body's top edge, blunt tip at the far right, lower edge sweeping back into the capsule. Applied to every primary button; label swaps go through a label span so the arrow survives | Geometry measured off the reference's own button in frame-004 by scanning the white pixels column by column (102 × 35 px, bite ~69% of the height, blunt tip, arrow inside the tail), scaled to Clover's 44 px button. Default/hover/press/focus captured at 3×; `mock-sweep.mjs` 14 screens × 4 widths: **no findings** (sweep 15) |

Two things stay open and are the manager's to settle in the next preview:

- **The attached button image never arrived.** `uploads/5c4cf17b-b2d7-469b-b68c-9f98836dc9d8-image.png`
  is not in the session — the attachment reader reports "file not found" — so the button was measured
  from the reference video's own "GET STARTED" button and the manager's written description instead.
  If the image differs, the difference is in the bite depth and the tip only.
- **One `motion-diff` line stays red, deliberately.** The tool pairs reference **beat 01** (0.0 s,
  1500 ms, linear) with the mock's first paint and reports it weak on duration and midpoint. Beat 01
  is the shot's own camera move over the presentation laptop — the page light stays steady while the
  whole frame moves — and rebuilding its 1500 ms would reintroduce exactly the slowness the manager
  reported. The product beat the reference actually shows (beat 02) is `OK`.

This revision supersedes the motion line in the scorecard above, which described the pre-revision
build: motion now reads as **one quick beat, `OK` against the product entrance**, with the camera-move
line standing unresolved by design rather than by omission.
