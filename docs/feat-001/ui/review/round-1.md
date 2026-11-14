# UI review — round 1

Feature `feat-001` (clover). Reviewed material: `UIMap.md`, `DesignBrief.md`, `ui/mock/` (14 screens),
the sweep of 14 screens × 4 widths, and the reference's measured motion. Three lanes ran against a
frozen copy of that material: **look**, **coverage + docs** (including the preview viewer), **motion**.

Result: 36 findings. The ones that needed a change are in the first table and were fixed; the rest
were repeats of them or wording nits resolved in the same edits. Two were raised against the shared
tooling rather than the design and are recorded as such.

## Fixed

| # | Lane | Finding | Fix |
|---|---|---|---|
| 1 | look | The player's error card was clipped at 390 px: the 16:9 stage is shorter than the card | `.stage.is-error` drops the ratio and takes a 380 px minimum below 820 px |
| 2 | look | Film cards had no press feedback and hovered via `background`, which cannot interpolate | Press scale 0.97 on the card, background removed from the transition |
| 3 | look | The browse chip read as flat glass on a bright photo | Frosted fill, stronger border, press state |
| 4 | look | Quiet buttons and icon buttons had no pressed state | Opacity 0.65 and scale 0.92 |
| 5 | look | Player and Film-not-found copy sat over the bright middle of the photo with no veil | A denser `--g-page-scrim` at every width for those two pages |
| 6 | look | The sign-in error alert looked like a sixth input | Left rule and plain text instead of a field-sized box |
| 7 | look | A field in error did not change its own border | `.field.has-error` rules the input |
| 8 | look | The skip link's focus ring measured 1.09:1 (a fixed element over white) | Skip link is absolute, ring drawn with an outline |
| 9 | coverage | An empty catalogue still showed "20 films", the facts card and a lede promising twenty films | Count and facts hidden, lede states the catalogue came back empty |
| 10 | motion | Every page arrival replayed the whole landing: the grid stayed invisible for 1500 ms and the chip for 2950 ms | The two beats run on the first paint only; later arrivals use the 700 ms rise |
| 11 | motion | The hero photo faded in over 1500 ms, tagged as measured, but `motion.txt` line 01 measures the light as steady | The photo holds still |
| 12 | motion | The second beat carried a blur the reference does not measure | Beat 02 rises without blur (`rise-flat`) |
| 13 | motion | A re-render of the current page (sign in, sign out) replayed the entrance | The arriving class is dropped when the entrance ends |
| 14 | coverage | `/watch/{movieId}` was not routable: no deep link, no missing-film path | `#watch-<id>` routes to the player, to sign-in with the film carried, or to Film-not-found |
| 15 | docs | Source cells carried citations, so the preview counted 0 derived rows | Citations moved to a Note column; the preview now counts 30 derived and 106 assumed |
| 16 | docs | The brief claimed a Medium font cut that is not shipped | Two files ship; the Regular face serves the 500 labels, and the brief says so |
| 17 | docs | The grid breakpoints were wrong (3 columns from 1020, actually 1180) | Corrected |
| 18 | docs | `--r-round` (50%) existed in the CSS but not in the token table | Added |
| 19 | docs | Shadow usage was wrong ("glass cards, toast"; "two shadows visible") | Corrected to the real surfaces and two distinct visible shadows |
| 20 | docs | The accent's usage list was wrong | Rewritten against the CSS |
| 21 | docs | Six token rows and two motion rows were tagged `derived` without evidence in the reference | Retagged `assumed` |
| 22 | docs | The brief carried a "Voice" line the map does not hold | Removed; the map is the source of truth |
| 23 | docs | MO-002's citation pointed at a stagger the measurement does not give | Rewritten with the measured numbers |
| 24 | docs | MO-005 said ease-out; the mock uses ease-in | Corrected |
| 25 | docs | `--c-danger-soft` was documented but no longer used | Token and row removed |

Smaller items folded into the same pass: the `#signin-error` screen had a filled password field while
the map says the password is cleared; the skeleton cards joined the card entrance; `--c-scrim-mobile`
was described as covering every page scrim when it covers the hero and sign-in only.

## Tooling, not the design

| # | Where | What |
|---|---|---|
| 26 | `docs-preview.js` | The preview's asset map does not resolve `ui/mock/assets/img/*` for some poster files; the screens still render in `/ui-preview` |
| 27 | `docs-preview.js` | Rows whose Source cell carried a citation were counted as neither derived nor assumed |

Both live in the shared script, outside this feature's files, so they are reported here and not
patched.

## Evidence after the fixes

- `mock-sweep.mjs`: 14 screens × 390/768/1024/1440 px — **no findings**.
- `fidelity-sheet.mjs` catalogue vs frame-004: **PASS**.
- `motion-diff.mjs`: **every reference animation matched**.
- `docs-preview.js`: 29 requirements, 24 mapped, 5 without UI, 0 uncovered; 30 derived, 106 assumed.

The motion tool pairs animations by shape rather than by time, so it matched beat 02 against a later
screen entry of the same shape; the beat itself plays at 2.8 s (recording animation 03 at 3.0 s).
