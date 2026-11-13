# UI review — round 2 (v1 rebuild, re-check)

The one re-check the supervised gate allows. Everything from round 1's findings
was re-captured at 1440 × 900, 768 × 1024 and 390 × 844 — every screen in
`screens.json` — and the two flows were re-run end to end. The docs were then
checked line by line against what the mock actually renders: `UIMap.md`,
`DESIGN.md`, `DesignBrief.md` and `knowledge/design-reference.md` all now
describe the v1 screens only.

## Scores

| Category | Lane | Round 1 | Round 2 |
| --- | --- | --- | --- |
| Reference fidelity | Look | 86 | 91 |
| Typography | Look | 84 | 92 |
| Colour and contrast | Look | 90 | 91 |
| Layout and composition | Look | 86 | 91 |
| Responsive at 768 | Look | 85 | 90 |
| Responsive at 390 | Look | 88 | 90 |
| Visual polish | Look | 80 | 91 |
| Entrance choreography | Motion | 90 | 90 |
| State changes and player feedback | Motion | 74 | 90 |
| Reduced motion | Motion | 92 | 92 |
| Performance | Motion | 88 | 90 |
| Requirement coverage | Coverage | 86 | 94 |
| State reachability | Coverage | 92 | 96 |
| Navigation and cross-linking | Coverage | 88 | 94 |
| Edge cases | Coverage | 70 | 88 |
| Doc completeness | Coverage | 55 | 94 |
| Doc accuracy | Coverage | 50 | 94 |
| Preview contract | Coverage | 90 | 95 |

## Verified in this round

- Every one of the 13 screens renders at all three widths with no overflow, no
  clipped headline and no console error.
- The catalogue grid is 5 / 4 / 3 / 2 columns as the breakpoints say; the
  headline is one line at every width.
- Sign up from a film: tab switch, invalid submit that keeps the typed email,
  valid submit, landing on `#/play/halcyon` with the video present and focus on
  `#main`. Sign in takes the same path.
- Signed-out catalogue shows "Sign in"; after account creation the header shows
  the account initial and "Sign out"; signing out returns to the anonymous
  catalogue.
- A signed-out `#/play/halcyon` redirects to `#/auth?next=halcyon` and does not
  render the player.
- The failed-video panel's "Try again" returns to the player for the same film.
- The scripted pass reports `errors=none` across all steps, including a full
  route change into and out of the player.
- Docs: every route, state, token and timing named in the three documents exists
  in the mock with the same name and value. The preview flags (`?u=1`,
  `?demo=1`, `?playing=1`, `?idle=1`, `?error=video`, `err=…`, `email=…`) are
  documented as mock-only and do not appear as product features.

## Fixed in this round

- The missing-film explanation wrapped with a single word on its second line
  (`max-width: 40ch` inside a 900 px panel). Widened to 54ch, so both player
  messages sit on one line.

## Still below the bar, and why it stays

- **Edge cases (88).** Two real failure paths cannot be exercised in a static
  mock: a failed catalogue fetch and a genuine media error. Both have final copy
  and layout in the mock; both are triggered by a preview flag. They are
  backend work, recorded in `UIMap.md`.
- **Reference fidelity (91).** The reference has no catalogue, no form and no
  player, so fidelity is judged on composition, type scale, colour, material and
  motion, not on a matching arrangement. The catalogue's masthead is a 42 vh
  band rather than a full screen because SC-001 needs films on first load.
- **Motion (90 for entrance).** The reference's single entrance is one rise and
  a fade; there is nothing to copy for state changes, so the player chrome's
  fade and the idle behaviour are house decisions at reference-consistent
  durations.
- **The muted-autoplay fallback.** A capture of `?playing=1` shows the mute
  icon because headless Chromium refuses unmuted autoplay. Real interaction
  unmutes; the state is correct, the capture is an artefact of the browser.

## Not changed in this round

No design was added and no scope was added. The round only verified the fixes
from round 1 and closed the documentation gap. The v1 interface is ready for the
manager's review in `/ui-preview`.

## Method notes

- Captures: headless Chromium CLI at 1440 × 900, 768 × 1024, 390 × 844,
  `--virtual-time-budget=2600`, one file per state, 19 captures in all.
- Flow: the same temporary iframe harness as round 1 (tabs, validation, submit,
  sign-out, guard, retry), deleted after the run. Log: invalid input keeps
  `neve`; email and password messages both fire; after submit
  `hash=#/play/halcyon`, `video=true`, `mainFocus=main`; `who=true` after
  sign-up; `who=false` after sign-out; guard lands on `#/auth?next=halcyon`;
  retry lands on `#/play/halcyon?demo=1`; `errors=none`.

## Preview contract, checked after the review

The manager's `/ui-preview` showed no page list. The preview that ran cloned the
docs repo at `main`; on `main` the feature folder exists (v0) but has no
`UIMap.md`, so the viewer had no pages. A preview built from this branch has
three. Two contract faults were found and fixed while checking it:

- The six account screens had `page` values `Sign in` and `Create account`,
  which match no `### ` heading. They now say `Sign in / Create account`.
- The viewer opens a screen by setting the mock's hash to its `screens.json`
  id, and the mock did not know those ids, so every screen fell through to the
  catalogue. Each id now routes to the state it names.

### Where the preview reads from

`/ui-preview` clones the docs repo at its base branch (`main`) and checks out a
feature's PR branch **only when the feature is missing from `main`**.
`clover/feature/FEAT-001` already exists on `main` as the rejected v0 draft and
has no `UIMap.md`, so a preview that reads only the merged repo shows no page
list even though this branch has three pages.

The preview does read a draft from the session workspace (`sessionDraftDir` in
the preview command) when a copy with `spec.md` and `UIMap.md` sits under it.
This session keeps the branch checkout at `sandbox-docs/`, so a plain
`/ui-preview service=clover feature=FEAT-001` reports it as the draft and builds
the viewer from it. Without that checkout in the session, the page list is
empty: this branch is never read while its folder exists on `main`.

Checked after adding the checkout: the viewer reports "UI map v2 in-review" and
its Pages menu lists Catalogue, Sign in / Create account and Player, built from
the same folder as the mock.

Checked on the branch build (the same generator, run locally): the page list
has Catalogue, Sign in / Create account and Player; the pages hold 2, 6 and 5
screens; all 13 ids open their own state; opening a state from a page's State
switcher shows that state in the stage; `errors=none`. The preview's inner CSP
allows no media, so the player is checked there with the poster-still clock: it
plays, pauses and seeks, and reports `0:02` after two seconds.

## After this round: the manager's feedback, 2026-11-13

The manager rejected the composition this round scored: the catalogue was still
a 42 vh masthead, and the reference's full-screen hero, its right-edge frosted
rail, its white stat card and its three glass feature cards were missing or
confined to the player. The catalogue and the account screen were rebuilt on
the reference (see `DesignBrief.md`, "Revision after the manager's review"), and
`UIMap.md`, `DESIGN.md` and `ui/mock/screens.json` were updated to match.

This was a self-check, not a third round — the supervised gate allows two. What
was checked on the rebuilt mock:

- All 13 `screens.json` ids still open their own state; the three pages and
  their screen counts are unchanged.
- The catalogue at 1440 × 900, 1280 × 980, 768 × 1024 and 390 × 844: the display
  line is one line at ~92 % of the column, the grids are 5 / 4 / 3 / 2 columns,
  the facts rail is hidden below 1100 px, the film rail stacks below 900 px, and
  nothing overflows or clips.
- The flow still holds: signed-out `/play/halcyon` redirects to the auth screen,
  sign-up returns to that film, play/pause and scrub act on the video, sign-out
  returns to the anonymous header, and the in-page scroll buttons move focus
  without disturbing the hash router.
- The `/ui-preview` stage: 3 pages, 13 screens, the rebuilt hero renders, and the
  hero asset inlines under the real CSP. The hero is capped at 980 px because
  the viewer sizes the stage to the mock's content height — an uncapped `100svh`
  stretched it.
- No console errors in any of these runs.

The rebuilt screens have not been scored against the round-2 table above; that
scorecard describes the previous composition and is kept as the record of the
round the gate allowed.

