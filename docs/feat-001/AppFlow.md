---
feature_code: feat-001
service: clover
spec_version: 1
last_updated: 2026-11-14
---

<!-- BEGIN:DOCPREP -->

# feat-001 — App Flow (Clover)

## Overview

Clover v1 has two main screens — **Catalogue** and **Player** — plus a
**Sign-in** screen shown only when a visitor who is not signed in selects a
movie. Browsing is public; playback is always behind sign-in. This is the
user-journey half only; routes and components belong to `/code-preparation`
and are appended below in its own section.

## Main pages and actions

- **Catalogue** — all 20 movies as cards; selecting a card opens the Player,
  or the Sign-in screen first when the visitor is not signed in. (REQ-001,
  REQ-002, REQ-003, REQ-004)
- **Sign-in** — username-or-email field plus password field; submit to sign
  in; shows one generic error on a failed attempt. (REQ-005, REQ-016)
- **Player** — the selected movie's video with play, pause, timeline scrub,
  volume and fullscreen; the movie title is shown alongside. (REQ-009,
  REQ-010, REQ-012)
- **Sign-out action** — returns the viewer to the public browsing
  experience. (REQ-008)

## Key User Journeys

### J-01 — Browse without signing in (REQ-001, REQ-002)

1. Visitor opens Clover.
2. The Catalogue appears with all 20 movies, each showing a poster, title and
   short description.
3. The visitor scrolls the full catalogue; nothing is hidden or locked, and no
   sign-in prompt blocks the view.

### J-02 — Pick a movie while signed out (REQ-003, REQ-004, REQ-011)

1. Visitor selects a movie card.
2. The Sign-in screen appears.
3. The visitor enters a username or email and a password.
4. On success, the Player opens for that same movie, ready to play.
5. If the visitor abandons sign-in instead, they remain anonymous and can
   keep browsing; the movie stays locked. (REQ-015)

### J-03 — Pick a movie while signed in (REQ-003)

1. Signed-in viewer selects a movie card.
2. The Player opens directly for that movie.

### J-04 — Watch and control playback (REQ-009, REQ-010, REQ-012)

1. Playback starts without waiting for the whole file to download.
2. The viewer uses play/pause, scrubs the timeline, changes volume and enters
   fullscreen.
3. Seeking to another point resumes playback from there rather than
   restarting.
4. The movie title is visible alongside the video.

### J-05 — Sign out (REQ-007, REQ-008, REQ-021)

1. Signed-in viewer signs out.
2. The Catalogue stays visible, but playback is locked again: selecting a
   movie asks for sign-in (J-02).
3. The old session can no longer be used on the server.

### J-06 — Open a player address directly (REQ-011, REQ-020)

1. A signed-out browser opens a movie's player address directly.
2. The Sign-in screen appears.
3. After a successful sign-in, that same movie's Player opens.
4. If a session becomes invalid while the viewer is watching, the next
   playback attempt follows the same path.

### J-07 — Recover from errors and empty states (REQ-016, REQ-018, REQ-019, REQ-023, REQ-024)

- Wrong username/email or password → one generic message; the form stays
  usable and the identifier is kept.
- Empty identifier or password → blocked with a "required" message before
  submitting.
- A player address for a movie that does not exist → "movie not available"
  with a way back to the Catalogue.
- A video that cannot load → clear message with a retry action, never an
  endless spinner.
- An empty catalogue → a clear "no movies available" state.
- A poster that fails to load → the card still shows the title and short
  description.

<!-- END:DOCPREP -->

<!-- BEGIN:CODEPREP -->

## Routes and components (technical)

Clover v1 is two applications: a Go API (`clover-be`) that serves JSON, poster
images and the video stream, and a React + Vite app (`clover-fe`) that renders
every screen and is reverse-proxied to the API on the same origin. The screens
and their copy are fixed by `UIMap.md`; this section records the routes, the
components and the API calls each screen makes.

### Frontend routes

| Route | Screen | Purpose | Auth |
|---|---|---|---|
| `/` | Catalogue | The public catalogue grid and the header's sign-in/sign-out action | public |
| `/sign-in?next=/watch/{movieId}` | Sign in | Authenticate a seeded account; carries the picked film in `next` | public (visitor) |
| `/watch/{movieId}` | Player | Play the selected film with custom controls | member |
| `/watch/{movieId}` (unknown id) | Film not found | "Film not found" with a way back to the catalogue | any |
| `*` | — | Redirect to `/` | — |

`next` is accepted only when it matches `^/watch/[a-z0-9-]+$`; any other value
is discarded (no open redirect). After a successful sign-in the app goes to
`next` when present, otherwise to `/`.

### Frontend components

| Component | Built from | Used on |
|---|---|---|
| `Header` (brand mark, nav, sign-in/sign-out action) | `UIMap.md` + `ui/mock/styles.css` | all screens |
| `FilmCard` (poster, index badge, title, description, arrow disc) | UIMap / mock | Catalogue |
| `PrimaryCapsuleButton` (white capsule, wavy tail, arrow) | UIMap / mock / manager's button reference | Header, Sign in, error cards |
| `GhostButton`, `QuietButton`, `IconButton` | UIMap / mock | Header, Sign in, Player |
| `FactsCard`, `BrowseChip` | UIMap / mock | Catalogue hero |
| `Field`, `FormAlert`, `TargetChip` | UIMap / mock | Sign in |
| `PlayerStage`, `Timeline`, `VolumeControl` | UIMap / mock | Player |
| `Toast`, `Spinner`, `SkeletonCard`, `EmptyPanel`, `PosterFallback` | UIMap / mock | Catalogue, Sign in, Player |
| `AuthProvider`, `RequireSession` | this design | all screens (session state and route guard) |
| `apiClient` (`fetch`, `credentials: 'same-origin'`, error envelope) | this design | all screens |

`nexcomponent` is not used anywhere. The design tokens and screen rules are
ported from `ui/mock/styles.css`; the mock is the visual source of truth.

### Backend API consumed by the screens

| Screen | Call | When |
|---|---|---|
| Catalogue | `GET /v1/movies` | page load; renders grid, empty panel or loading skeleton |
| Catalogue | `GET /v1/auth/me` | app boot; chooses the header action and whether cards open the player |
| Catalogue | `POST /v1/auth/logout` | header "Sign out" |
| Sign in | `GET /v1/movies/{movieId}` | only when `next` names a film, for the target chip |
| Sign in | `POST /v1/auth/login` | submit; sets the session cookie |
| Player | `GET /v1/auth/me` | route load and every playback revalidation |
| Player | `GET /v1/movies/{movieId}` | title, description and index beside the stage |
| Player | `GET /v1/media/movies/{movieId}/video` | the `<video>` source; supports Range requests for seeking |
| Film not found | `GET /v1/movies/{movieId}` | a 404 response renders the not-found screen |

Every screen has three request states (loading, error, ready), plus the
empty-catalogue and poster-failure states on the Catalogue and the
video-failure state on the Player, exactly as `UIMap.md` lists them.

<!-- END:CODEPREP -->
