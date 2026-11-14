---
feature_code: feat-001
service: clover
spec_version: 1
last_updated: 2026-11-14
---

# feat-001 — Movie Streaming (Clover) — PRD

## Overview

Clover v1 is a small movie streaming web app for one business. The movie
catalogue is public: a visitor can browse all films before signing in.
Watching is gated: a visitor who selects a film signs in with a pre-created
account and lands directly on that film's player. This summary is a
helicopter view only — `spec.md` holds the full requirement detail.

## Product Goal

Let a visitor open Clover, pick a movie, sign in when required, and start
watching — with no subscriptions, recommendations or management screens in
v1. (REQ-001, REQ-004, REQ-009)

## Features

| Feature | Description | Expected behavior |
|---|---|---|
| Public catalogue | Grid of all 20 v1 movies, each with poster, title and short description. | Visible without signing in; no pagination, search, filter or sorting. (REQ-001, REQ-002) |
| Pre-created accounts | Sign-in with a username or email plus password; no self sign-up. | Wrong credentials show one generic error; empty fields are blocked; no lockout in v1. (REQ-005, REQ-006, REQ-015, REQ-016, REQ-017, REQ-025, REQ-026, REQ-028) |
| Sign-in gate with return to the picked film | Selecting a movie while signed out opens sign-in first. | After a successful sign-in the viewer lands on that same movie's player, ready to play. (REQ-003, REQ-004, REQ-011) |
| Persistent session and sign-out | A signed-in viewer stays signed in on that device until signing out. | Sign-out returns to public browsing and invalidates the session on the server; playback locks again. (REQ-007, REQ-008, REQ-021, REQ-022) |
| Video player | Streams the selected movie with standard controls. | Play, pause, timeline scrub, volume and fullscreen work; seeking resumes from the chosen point. (REQ-009, REQ-010, REQ-012) |
| Sample catalogue and media | 20 seeded movies, each with a working placeholder video. | Every movie can be selected and played end to end. (REQ-013, REQ-029) |
| English interface, small scale | English-only copy for a single business. | Caters to a few hundred accounts and about 30 concurrent viewers. (REQ-014) |
| Error and empty states | Empty catalogue, unknown movie, broken poster, unplayable video. | Clear messages and recovery actions instead of blank pages or endless spinners. (REQ-018, REQ-019, REQ-023, REQ-024) |

For full requirement detail, acceptance criteria and deferred scope, see
`spec.md`.
