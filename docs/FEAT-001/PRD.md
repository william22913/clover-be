---
feature_code: FEAT-001
service: clover
spec_version: 1
last_updated: 2026-11-13
---

# Clover — Product Requirements Overview (FEAT-001)

## Overview

Clover is a simple movie streaming web app. A visitor opens the site, sees what
is on, picks a movie, and signs in only at the moment they want to watch it.
This document is a short overview only — `spec.md` holds the authoritative
requirement detail, and every line here traces back to a `REQ-*` in it.

## Product Goal

Make "open the website, pick a movie, and start watching" work end to end in a
browser, without accounts getting in the way of browsing. In v1 that means a
public catalogue, self-service email/password accounts, and a player that
actually streams and seeks.

## Features

| Feature | Description | Expected behavior |
|---|---|---|
| Public catalogue | Every published sample movie in a grid with poster, title and short description. | Visible immediately with no login, no pagination, no search or sorting (REQ-001, REQ-002) |
| Pick a movie | Selecting a movie opens its player page. | Works for anonymous visitors too — selection is what triggers the login check (REQ-003) |
| Login-gated watching | A signed-out visitor is sent to a combined login/sign-up screen. | After signing in or signing up they land on the movie they picked, with the video ready to play — never back at the catalogue (REQ-004, REQ-012, REQ-019) |
| Self-service accounts | Email and password sign-up, no verification. | Password: at least 8 characters with a letter and a number. One account per email, case-insensitive. Immediate access after sign-up (REQ-005, REQ-006, REQ-022, REQ-024) |
| Persistent sessions | Server-side sessions with a 7-day cookie. | Users stay signed in between visits; sign-out truly revokes the session; several devices may be signed in at once (REQ-007, REQ-010, REQ-020, REQ-021) |
| Video player | Streaming playback with the standard controls. | Play, pause, timeline scrub, volume, fullscreen; seeking resumes from the new point rather than restarting (REQ-008, REQ-009) |
| Failure and recovery | Clear messages for the predictable mistakes. | Wrong credentials (no hint which was wrong), duplicate email, invalid sign-up input, missing movie, unloadable video, expired session — each with a way forward (REQ-013 to REQ-019) |
| English-only interface | All copy in English. | No other language in v1 (REQ-011) |

## Out of scope

Recommendations, subscriptions and payments, search and filtering, any screen
for managing movies, email verification, password reset, profiles, watchlists,
ratings, comments, and non-English UI. See `spec.md`'s Non-Goals.

## Full detail

`spec.md` — requirements, edge cases, validation rules and acceptance criteria.
