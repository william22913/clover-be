# UI review — round 1 (v1 rebuild)

First review of the rebuilt v1 mock: catalogue, sign-in / create account,
player. The previous review rounds in this folder scored a draft that has since
been removed; this file replaces them, and the same rubric is applied to the
screens that shipped.

Three lanes, as the supervised gate allows: **Look** (reference fidelity,
typography, colour, layout, responsive, polish), **Motion** (entrance, state
changes, reduced motion, performance) and **Coverage** (requirements, states,
navigation, edge cases, docs accuracy, preview contract). Passes were made
against captures of every screen in `screens.json` at 1440 × 900, 768 × 1024
and 390 × 844, plus a scripted pass through the two flows: create an account
from a film, and recover from a failed video.

## Scores

| Category | Lane | Round 1 |
| --- | --- | --- |
| Reference fidelity | Look | 86 |
| Typography | Look | 84 |
| Colour and contrast | Look | 90 |
| Layout and composition | Look | 86 |
| Responsive at 768 | Look | 85 |
| Responsive at 390 | Look | 88 |
| Visual polish | Look | 80 |
| Entrance choreography | Motion | 90 |
| State changes and player feedback | Motion | 74 |
| Reduced motion | Motion | 92 |
| Performance | Motion | 88 |
| Requirement coverage | Coverage | 86 |
| State reachability | Coverage | 92 |
| Navigation and cross-linking | Coverage | 88 |
| Edge cases | Coverage | 70 |
| Doc completeness | Coverage | 55 |
| Doc accuracy | Coverage | 50 |
| Preview contract | Coverage | 90 |

Weakest: visual polish (80), player feedback (74), edge cases (70), and the
docs — which at this point in the round still described the removed draft.

## Findings fixed in this round

1. **Every untagged icon rendered at full size.** `ic()` emitted
   `class=""` for untagged icons, so the auth back chevron drew at ~200 px and
   the goo arrow overflowed the capsule. The class attribute is now built
   correctly and every icon carries `.ic` (16 px, 20 px in the player bar).
2. **The catalogue headline filled only 73 % of the column** against the
   reference's ~92 %. Raised from `clamp(48px,12.5vw,210px)` to
   `clamp(46px,15.4vw,260px)` at `-.034em`; measured at about 90 % at 1440 and
   still one line at 390.
3. **The gooey capsule rendered as a plain pill.** The background body filled
   the entire button, covering the knob. The body is now `inset: 0 30px 0 0`
   with the knob at `right: 0; top: 4px`, and an 8× zoom confirms the fused
   fillet in the reference's shape.
4. **The player's top bar was unreadable over the bright sky.** The gradient is
   now `rgba(5,5,5,.94) → .74 at 46% → 0` with 56 px of bottom padding, plus a
   text shadow on the title and meta.
5. **The error heading wrapped onto two lines** ("THE FILM COULD NOT BE /
   LOADED."). The message panel is 900 px wide with no character cap; both
   messages now sit on one line at 1440.
6. **"Try again" did nothing.** It was a `type="submit"` button outside a form,
   and the retry listener targeted an id that was never rendered. It is now a
   gooey link to the same film's clean player route; the flow test clicks it and
   lands on the player.
7. **The signed-in catalogue did not show the session.** `?u=1` set the user
   after the header had been built, so the header still offered "Sign in". The
   demo session is now applied before rendering, and the header shows the
   account initial and "Sign out".
8. **A duplicate-email message could not be cleared.** The “Sign in instead.”
   link inside the field message was skipped by `clearErrors`, leaving the error
   on screen after switching tabs. Errors are now always hidden and emptied.
9. **An invalid email at sign-up was marked red with no message.** The
   `err=invalid` state now also states "Enter an email address like
   you@example.com.".
10. **Two `id` attributes on the player `<main>`.** Only `main` remains.
11. **The auth meta line was uppercased** by an inherited rule, so "2H 06M"
    read as a label. It is sentence case now.
12. **A `?playing=1` capture showed no chrome** because the idle timer fired
    inside the capture's virtual-time budget. The state now holds its chrome
    until the pointer moves, which is what the screen is meant to show.
13. **The player's mute icon appeared on the playing capture.** Headless
    Chromium blocks unmuted autoplay, so the mock falls back to muted playback.
    The fallback is correct behaviour, not a design state: any real pointer
    interaction unmutes again. Documented here rather than hidden.

## Open items, deliberately not fixed here

- **The catalogue has no failed-load state.** The mock's films are a constant,
  so nothing can fail. The Go build must render an equivalent error panel in the
  same voice; recorded in `UIMap.md` under Not mocked.
- **The player's error state is opened by a preview flag**, not by a real
  media failure. The panel, copy and retry are final; the trigger is server work.
- **Every film plays the same 14 s sample clip.** It exists so the controls can
  be exercised honestly; real films are content, not design.
- **Key art is placeholder.** Thirteen Unsplash posters and three stills; the
  CSS grade means replacements do not need re-grading, but each replacement must
  be chosen at a similar exposure or it will disappear into the scrims.
- **Session, password and cookie behaviour is not represented.** The mock signs
  in on any valid form because it has no server; the 7-day cookie, hashing and
  server-side sign-out are backend requirements.
- **Hover, focus and keyboard behaviour are derived.** The reference is one
  static screen; nothing in it shows interaction. Every such state is a house
  choice, marked in the brief.

## Method notes

- Captures: headless Chromium CLI, `--window-size` at the three widths,
  `--virtual-time-budget=2600` so fonts and entrances settle, one file per state.
- Flows: a temporary page loaded the mock in an iframe, switched tabs, submitted
  invalid then valid credentials, signed out, followed the signed-out player
  URL, and clicked retry, logging each assertion and any console error to the
  page. It reported every step correct and `errors=none`, and was deleted after
  the run.
