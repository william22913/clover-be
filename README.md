# clover-be

Clover is a small movie-streaming web app: open the site, browse a public
catalogue, pick a film, and — because watching needs an account — sign in, then
watch that film.

This repository holds the single Go service that serves the JSON API, the
seeded posters and the video stream, and owns all persistence. It serves no
HTML: the browser app is the separate `clover-fe` repository, which
reverse-proxies `/v1/` here. Accounts are created by Clover and seeded; v1 has
no sign-up.

## Run it

```sh
cp .env.example .env      # DATABASE_URL, PUBLIC_ORIGIN
docker compose up --build
```

The API listens on `:8080` (`/v1`, `/healthz`, `/readyz`); the metrics listener
on `:9090` is internal and never published. `cmd/seed` applies the migrations
and inserts the 20-film catalogue and the demo accounts; compose runs it once
before the API starts.

Development conventions, layout and the feature docs are indexed in
[CLAUDE.md](CLAUDE.md).
