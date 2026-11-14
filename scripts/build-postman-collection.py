#!/usr/bin/env python3
"""Builds postman/collection.json for the Clover API. Run once; the JSON is the
deliverable, this script is not shipped as part of the service."""
import json

BASE = "{{baseUrl}}"


def req(name, method, path, *, body=None, headers=None, tests=None, description=None, auth=True):
    url = path if path.startswith("{{metricsUrl}}") else BASE + path
    item = {
        "name": name,
        "request": {
            "method": method,
            "header": [{"key": k, "value": v} for k, v in (headers or [])],
            "url": url,
        },
    }
    if description:
        item["request"]["description"] = description
    if body is not None:
        item["request"]["body"] = {
            "mode": "raw",
            "raw": json.dumps(body, indent=2),
            "options": {"raw": {"language": "json"}},
        }
    if tests:
        item["event"] = [{"listen": "test", "script": {"type": "text/javascript", "exec": tests}}]
    return item


def status(code):
    return [f"pm.test('status is {code}', () => pm.response.to.have.status({code}));"]


def error_code(code):
    return [
        f"pm.test('code is {code}', () => {{",
        "  const body = pm.response.json();",
        f"  pm.expect(body.error.code).to.eql('{code}');",
        "  pm.expect(body.error.message).to.be.a('string').and.not.empty;",
        "  pm.expect(body.error.requestId).to.be.a('string').and.not.empty;",
        "});",
    ]


health = [
    req("Liveness", "GET", "/healthz", tests=status(200) + [
        "pm.test('body is ok', () => pm.expect(pm.response.text().trim()).to.eql('ok'));"]),
    req("Readiness", "GET", "/readyz", tests=status(200) + [
        "pm.test('body is ready', () => pm.expect(pm.response.text().trim()).to.eql('ready'));"]),
    req("Metrics exposition", "GET", "{{metricsUrl}}/metrics", tests=status(200) + [
        "pm.test('prometheus text format', () => pm.expect(pm.response.headers.get('Content-Type'))",
        "  .to.include('text/plain'));",
        "pm.test('request counters are exposed', () => pm.expect(pm.response.text())",
        "  .to.include('http_requests_total'));",
        "pm.test('request duration histogram is exposed', () => pm.expect(pm.response.text())",
        "  .to.include('http_request_duration_seconds'));"]),
]

catalogue = [
    req("List the catalogue (anonymous)", "GET", "/v1/movies", tests=status(200) + [
        "const body = pm.response.json();",
        "pm.test('twenty films', () => pm.expect(body.movies).to.have.lengthOf(20));",
        "pm.test('every film carries its fields', () => {",
        "  body.movies.forEach((m) => {",
        "    pm.expect(m.id).to.be.a('string').and.not.empty;",
        "    pm.expect(m.title).to.be.a('string').and.not.empty;",
        "    pm.expect(m.description).to.be.a('string').and.not.empty;",
        "    pm.expect(m.posterUrl).to.eql('/v1/media/movies/' + m.id + '/poster');",
        "    pm.expect(m.videoUrl).to.eql('/v1/media/movies/' + m.id + '/video');",
        "    pm.expect(m.durationSeconds).to.be.within(1, 1800);",
        "  });",
        "});",
        "pm.test('editorial order', () => {",
        "  body.movies.forEach((m, i) => pm.expect(m.position).to.eql(i + 1));",
        "});",
        "pm.collectionVariables.set('filmId', body.movies[0].id);",
        "pm.collectionVariables.set('otherFilmId', body.movies[2].id);",
        "pm.collectionVariables.set('longestFilmId', 'long-evening');",
    ]),
    req("One film", "GET", "/v1/movies/{{filmId}}", tests=status(200) + [
        "const m = pm.response.json().movie;",
        "pm.test('same film', () => pm.expect(m.id).to.eql(pm.collectionVariables.get('filmId')));",
        "pm.test('title and description', () => {",
        "  pm.expect(m.title).to.be.a('string').and.not.empty;",
        "  pm.expect(m.description).to.be.a('string').and.not.empty;",
        "});",
    ]),
    req("Unknown film is a clean 404", "GET", "/v1/movies/does-not-exist",
        tests=status(404) + error_code("MOVIE_NOT_FOUND")),
    req("English whatever Accept-Language says", "GET", "/v1/movies",
        headers=[("Accept-Language", "fr-FR")], tests=status(200) + [
            "pm.test('English titles', () => {",
            "  const titles = pm.response.json().movies.map((m) => m.title);",
            "  pm.expect(titles).to.include('Amber Drift');",
            "});"]),
]

auth = [
    req("Wrong password is one generic refusal", "POST", "/v1/auth/login",
        headers=[("Content-Type", "application/json")],
        body={"identifier": "{{username}}", "password": "Wrong1234"},
        tests=status(401) + error_code("AUTH_INVALID_CREDENTIALS") + [
            "pm.test('no field detail', () => pm.expect(pm.response.json().error.fields).to.be.undefined);"]),
    req("Unknown account is the same refusal", "POST", "/v1/auth/login",
        headers=[("Content-Type", "application/json")],
        body={"identifier": "nobody_here", "password": "{{password}}"},
        tests=status(401) + error_code("AUTH_INVALID_CREDENTIALS")),
    req("Empty identifier names the field", "POST", "/v1/auth/login",
        headers=[("Content-Type", "application/json")],
        body={"identifier": "", "password": "{{password}}"},
        tests=status(422) + error_code("VALIDATION_ERROR") + [
            "pm.test('identifier message', () => pm.expect(pm.response.json().error.fields.identifier)",
            "  .to.eql('Enter your username or email.'));"]),
    req("Cross-origin POST is refused", "POST", "/v1/auth/login",
        headers=[("Content-Type", "application/json"), ("Origin", "https://evil.example")],
        body={"identifier": "{{username}}", "password": "{{password}}"},
        tests=status(403) + error_code("CSRF_REJECTED")),
    req("Sign in with the username", "POST", "/v1/auth/login",
        headers=[("Content-Type", "application/json"), ("Origin", "{{publicOrigin}}")],
        body={"identifier": "{{username}}", "password": "{{password}}"},
        tests=status(200) + [
            "pm.test('signed in as the seeded account', () => {",
            "  pm.expect(pm.response.json().user.username).to.eql(pm.collectionVariables.get('username'));",
            "});",
            "pm.test('session cookie is HttpOnly', () => {",
            "  const cookie = pm.response.headers.get('Set-Cookie') || '';",
            "  pm.expect(cookie).to.include('clover_session=');",
            "  pm.expect(cookie.toLowerCase()).to.include('httponly');",
            "});",
            "pm.test('session cookie carries the S7 attributes', () => {",
            "  const cookie = pm.response.headers.get('Set-Cookie') || '';",
            "  const lower = cookie.toLowerCase();",
            "  pm.expect(lower).to.include('samesite=lax');",
            "  pm.expect(lower).to.include('path=/');",
            "  pm.expect(lower).to.include('max-age=2592000');",
            "});"]),
    req("The account comes from the session", "GET", "/v1/auth/me", tests=status(200) + [
        "pm.test('own account', () => pm.expect(pm.response.json().user.username)",
        "  .to.eql(pm.collectionVariables.get('username')));"]),
    req("Email sign-in, uppercase", "POST", "/v1/auth/login",
        headers=[("Content-Type", "application/json")],
        body={"identifier": "ALEX.RIVERA@EXAMPLE.COM", "password": "Watch1234"},
        tests=status(200) + [
            "pm.test('same account', () => pm.expect(pm.response.json().user.username).to.eql('alex'));"]),
    req("Sign out", "POST", "/v1/auth/logout", tests=status(204)),
    req("Signed out cannot read the account", "GET", "/v1/auth/me",
        tests=status(401) + error_code("UNAUTHENTICATED")),
    req("Signed out cannot fetch video", "GET", "/v1/media/movies/{{filmId}}/video",
        tests=status(401) + error_code("UNAUTHENTICATED")),
    req("Signed out cannot fetch a byte range", "GET", "/v1/media/movies/{{filmId}}/video",
        headers=[("Range", "bytes=0-99")], tests=status(401) + error_code("UNAUTHENTICATED")),
]

media = [
    req("Sign in for playback", "POST", "/v1/auth/login",
        headers=[("Content-Type", "application/json")],
        body={"identifier": "{{username}}", "password": "{{password}}"},
        tests=status(200)),
    req("Poster is public WebP", "GET", "/v1/media/movies/{{filmId}}/poster", tests=status(200) + [
        "pm.test('webp', () => pm.expect(pm.response.headers.get('Content-Type')).to.eql('image/webp'));",
        "pm.test('cacheable', () => pm.expect(pm.response.headers.get('Cache-Control'))",
        "  .to.eql('public, max-age=86400'));"]),
    req("Video plays from a byte range", "GET", "/v1/media/movies/{{filmId}}/video",
        headers=[("Range", "bytes=0-3")], tests=status(206) + [
            "pm.test('partial content', () => pm.expect(pm.response.headers.get('Content-Range'))",
            "  .to.match(/^bytes 0-3\\//));",
            "pm.test('mp4', () => pm.expect(pm.response.headers.get('Content-Type')).to.eql('video/mp4'));",
            "pm.test('never cached', () => {",
            "  pm.expect(pm.response.headers.get('Cache-Control')).to.eql('private, no-store');",
            "  pm.expect(pm.response.headers.get('Vary')).to.eql('Cookie');",
            "});"]),
    req("Video answers HEAD", "HEAD", "/v1/media/movies/{{filmId}}/video", tests=status(200) + [
        "pm.test('range support advertised', () => pm.expect(pm.response.headers.get('Accept-Ranges'))",
        "  .to.eql('bytes'));",
        "pm.test('a known length', () => pm.expect(Number(pm.response.headers.get('Content-Length')))",
        "  .to.be.above(0));"]),
    req("Seek deep into the 16-minute film", "GET", "/v1/media/movies/{{longestFilmId}}/video",
        headers=[("Range", "bytes=100000-")], tests=status(206) + [
            "pm.test('seek honoured', () => pm.expect(pm.response.headers.get('Content-Range'))",
            "  .to.match(/^bytes 100000-/));"]),
    req("Unknown film video is a clean 404", "GET", "/v1/media/movies/does-not-exist/video",
        tests=status(404) + error_code("MOVIE_NOT_FOUND")),
]

def preflight_item(name, path):
    return req(name, "OPTIONS", path, tests=status(405) + error_code("METHOD_NOT_ALLOWED") + [
        "pm.test('no CORS headers: the SPA is same-origin through its proxy', () => {",
        "  pm.expect(pm.response.headers.get('Access-Control-Allow-Origin')).to.be.undefined;",
        "});"])


preflight = [
    preflight_item("Preflight: catalogue", "/v1/movies"),
    preflight_item("Preflight: one film", "/v1/movies/{{filmId}}"),
    preflight_item("Preflight: sign in", "/v1/auth/login"),
    preflight_item("Preflight: sign out", "/v1/auth/logout"),
    preflight_item("Preflight: account", "/v1/auth/me"),
    preflight_item("Preflight: poster", "/v1/media/movies/{{filmId}}/poster"),
    preflight_item("Preflight: video", "/v1/media/movies/{{filmId}}/video"),
]

collection = {
    "info": {
        "name": "Clover API",
        "description": (
            "End-to-end checks for the Clover v1 API. Folders follow the controllers "
            "(Health, Catalogue, Auth, Media) plus one OPTIONS preflight per browser-called "
            "endpoint. Run against a seeded stack; the sign-in requests set and clear the "
            "session cookie like a browser."
        ),
        "schema": "https://schema.getpostman.com/json/collection/v2.1.0/collection.json",
    },
    "variable": [
        {"key": "username", "value": "alex"},
        {"key": "password", "value": "Watch1234"},
        {"key": "filmId", "value": "amber-drift"},
        {"key": "otherFilmId", "value": "long-evening"},
        {"key": "longestFilmId", "value": "long-evening"},
    ],
    "item": [
        {"name": "Health", "item": health},
        {"name": "Catalogue", "item": catalogue},
        {"name": "Auth", "item": auth},
        {"name": "Media", "item": media},
        {"name": "Preflight", "item": preflight},
    ],
}

with open("postman/collection.json", "w") as f:
    json.dump(collection, f, indent=2)
    f.write("\n")
print("wrote postman/collection.json")
