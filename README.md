# Mirror for PowerSchool

A fast, glassy mirror of a student's PowerSchool: today's classes, a week timetable, courses, term grades, assignment scores, attendance and archived report cards. It works on phone, iPad and desktop.

```
server/   Go API (gin). Talks to PowerSchool; stateless.
web/      React + Tailwind + shadcn app.
```

## Run

```sh
make dev          # synthetic data (MOCK=1 API on :8080 + Vite on :5173); any username/password signs in
make server web   # real school: put SCHOOL_BASE_URL in .env (see .env.example)
make test         # Go + Vitest
make live-test    # one real login with the .env account; prints counts only
make gen          # regenerate web/src/api/*.gen.ts after changing Go DTOs
make build        # bin/server + web/dist
docker compose up -d --build   # one container on :8000 (PORT=…): Gin serves the API and the built app; reads .env
```

When the web app is hosted on a different origin from the API, set `VITE_API_BASE=https://api.example.org` at build time and `ALLOWED_ORIGINS=https://app.example.org` on the server.

## How it works

The browser can't call PowerSchool directly: the mobile API needs HTTP Digest auth, custom headers and SOAP requests, and PowerSchool sends no CORS headers. The Go server is a thin relay to one configured school host:

- `internal/powerschool/mobile`: the PowerSchool mobile app's SOAP-in, JSON-out protocol (`loginToPublicPortal`, `getStudentData`, `getStartStopTimeForAllSections`), following [powerschool-api-analysis](https://github.com/HenryXiaoYang/powerschool-api-analysis).
- `internal/powerschool/portal`: the web portal's form login. It is used only for report cards, which are re-served as script-free HTML with images inlined.
- `internal/student`: joins and school rules: assignment ↔ score, attendance ↔ course, A/B rotation days, Asia/Shanghai dates, suppressed percents. It never computes a grade.
- `internal/session`: both upstream sessions are sealed (AES-GCM) into the bearer token. The server stores nothing.
- `internal/httpapi`: routes, CORS, gzip, sign-in rate limits (strict per account and IP, looser per IP, so classmates on one school network don't block each other), and logs that never include bodies or tokens.

| Endpoint | |
|---|---|
| `POST /api/session` | `{username, password}` → `{token, reportCards}` |
| `GET /api/student` | everything for the views (current month's schedule included) |
| `GET /api/schedule?year=&month=` | another month's meetings |
| `GET /api/report-cards`, `GET /api/report-cards/{id}` | archived report cards |

Errors are `{error: {code, message}}`. On `session_expired` the app signs in again silently with the stored credentials and retries the request once.

## Credentials and data on the device

- The password is encrypted with a non-extractable WebCrypto AES key kept in IndexedDB. The encrypted copy lives in `sessionStorage` (for this tab) or, with **Remember me**, in `localStorage`. This protects against someone copying the browser's storage; it does not protect against script running on the page.
- Student data is cached in IndexedDB so the app opens instantly. It refreshes on open or focus when it is more than 10 minutes old. Signing out wipes the cache, the stored credentials and the key.

## Server settings

`SCHOOL_BASE_URL` (required), `SESSION_KEY` (`openssl rand -base64 32`; otherwise a random key per process), `ALLOWED_ORIGINS`, `SCHOOL_TIMEZONE` (default `Asia/Shanghai`), `SESSION_MAX_AGE` (default `12h`), `TRUSTED_PROXIES`, `ADDR` (default `:8080`), `MOCK=1`, `WEB_DIR` (serve the built app, e.g. `../web/dist`; unknown paths get `index.html`).

## Known limits

- Only the first student on an account is shown (parent accounts with several children would need a switcher).
- Term grades appear only once the school posts them. When the next term's grades are out, run `make live-test` and check `finalGrades=` is non-zero. If the mobile API stays empty while the portal shows grades, the fallback is scraping `guardian/home.html`.
- SSO/MFA accounts are not supported (the protocol uses username/password).
- `server/internal/powerschool/fake/fixture.json` is synthetic: real timetable shapes, invented people and scores.
