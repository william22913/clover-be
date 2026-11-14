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
binds `127.0.0.1:9090` by default and is never published. `cmd/seed` applies the
migrations and inserts the 20-film catalogue and the demo accounts; compose runs
it once before the API starts.

### Behind the frontend proxy

The sign-in rate limit is per client address. When the API is reached through
the frontend's nginx, every connection carries the proxy's address, so set
`TRUSTED_PROXY_CIDRS` to the proxy's range (compose default:
`172.16.0.0/12`) and let nginx forward the original client:

```nginx
proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
```

Without one or the other, all visitors share a single bucket. A caller reaching
the API directly cannot spoof the header: it is only read from a trusted peer.

Development conventions, layout and the feature docs are indexed in
[CLAUDE.md](CLAUDE.md).

## Continuous integration

`ci/github-actions-ci.yml` is the pipeline: format and vet checks, the full Go
suite against a PostgreSQL service (the database-backed cases run there), the
JUnit XML report published as the `go-test-report` artifact, and an image build.

It is kept under `ci/` rather than `.github/workflows/` because this
environment's git credential has no `workflow` scope, so a push that creates a
workflow file is rejected by GitHub. Install it with:

```sh
mkdir -p .github/workflows && cp ci/github-actions-ci.yml .github/workflows/ci.yml
```
