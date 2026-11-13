# clover-be

Clover is a small movie-streaming web app: open the site, browse a public
catalogue, pick a film, and — because watching needs an account — sign in or
create one, then watch that film.

This repository holds the single Go service that serves the JSON API, the static
UI, the posters and the video stream, and owns all persistence.

## Local stack

```sh
cp .env.example .env   # then choose your own local passwords
docker compose up --build
```

The API and the UI are on <http://localhost:8080>; `/metrics` is served on the
separate internal listener (`:9101`) and is reachable only inside the compose
network. `.env` is gitignored and the compose file refuses to start without it.

## CI

`ci/github-actions.yml` is the pipeline: gofmt, vet, build, `go test` with a
JUnit report published as an artifact, then the image build. Move it to
`.github/workflows/ci.yml` to activate it — it sits outside `.github/` only
because the branch was pushed with a token that lacks the `workflow` scope.
