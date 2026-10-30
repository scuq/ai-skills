---
name: scuq-go-spa-stack
description: >
  The technology stack and the rules for a single-page web application
  with a Go server as one static binary, and no database, PostgreSQL 18,
  or MongoDB 8. Load when you plan, write, extend, or review such an
  application: the Go server, WebSocket push, database change
  notifications, the Svelte client, modal dialogs, per-user settings,
  authentication, security headers, the workstation layout with a
  docking grid, and tests.
---

# scuq-go-spa-stack

This skill fixes the stack and the rules for a single-page application (SPA) with a Go server.
The server is one static binary that serves the client files, the API, and a WebSocket endpoint.
The client is a TypeScript application with Svelte 5.
The data store is memory and files, PostgreSQL, or MongoDB.

The rules come from the sources in `reference/sources.md`.
When a source and this file disagree, this file decides.
Check the version floors in this file against the package registries when you start an application.

## 1. The variants

An application uses one of three variants.

| Variant | Data store | Change source | Session store | User settings |
|---|---|---|---|---|
| `none` | Memory, and JSON files in a data directory | In-process event bus | `memstore` | One JSON file for each user |
| `postgres` | PostgreSQL 18.6 or later | Trigger, `pg_notify`, and an outbox table | `pgxstore` | Table `user_settings` with a JSONB column |
| `mongo` | MongoDB 8.0 or later, as a replica set | Change stream | `mongodbstore` | Subdocument `settings` in the user document |

- Decide the variant before you write code.
- Put the data access behind one Go interface for each entity, so that a test can use a fake.
- Do not write a storage layer that supports all three variants in one build, unless the task asks for it.
- The `none` variant has no durability promise beyond the data directory.
  Say so in the README of the application.

## 2. The server stack

| Component | Choice | Version floor | Reason |
|---|---|---|---|
| Language | Go | 1.27 | The latest stable release, see the rule below |
| Router | `net/http.ServeMux` | Go 1.22 patterns | Method and wildcard patterns, no dependency |
| Static files | `embed` and `fs.Sub` | | One binary |
| WebSocket | `github.com/coder/websocket` | 1.8 | Context-aware API, small, maintained |
| PostgreSQL driver | `github.com/jackc/pgx/v5` | 5.11 | `LISTEN`, `COPY`, batches, pool |
| SQL code | `sqlc` | 1.29 | Typed Go from SQL at compile time |
| Migrations | `github.com/pressly/goose/v3` | 3.24 | SQL files in `embed.FS`, library and CLI |
| MongoDB driver | `go.mongodb.org/mongo-driver/v2` | 2.9 | Change streams, official |
| Sessions | `github.com/alexedwards/scs/v2` | 2.8 | Server-side data, OWASP pattern, stores for pgx and MongoDB |
| Password hash | `golang.org/x/crypto/argon2` | | argon2id |
| Passkeys | `github.com/go-webauthn/webauthn` | 0.17 | FIDO2 conformant |
| OIDC | `github.com/coreos/go-oidc/v3` and `golang.org/x/oauth2` | | ID token validation, PKCE |
| Rate limit | `golang.org/x/time/rate` | | Token bucket for each key |
| Logging | `log/slog` | | JSON output, no dependency |
| Tracing | `go.opentelemetry.io/contrib/bridges/otelslog` | optional | Trace ids in the log records |

- Always use the latest stable Go version, 1.27 or later.
  Set the `go` directive in `go.mod` to it, and raise it with each Go release.
  Go gives fixes to the two most recent minor versions only.
- Keep the `net/http` server at its defaults.
  Set only the timeouts, the body limits, and the WebSocket limits of sections 4 and 5, and change another default only with a measurement.
- Use the standard library when it does the job.
- Do not add an ORM or a web framework.
- Do not add a dependency that needs cgo.
  Build with `CGO_ENABLED=0`, `-trimpath`, and `-ldflags "-s -w -X main.version=VERSION"`.
- Do not use `encoding/json/v2` in a release build.
  It is still an experiment behind `GOEXPERIMENT=jsonv2`.
- Section 4 and `reference/server.md` give the layout and the code.

## 3. The client stack

The client stack is a suggestion, not a requirement.
For a new application, use it as the first choice, and change a row only with a reason in the README.
For an existing application, keep its framework and its libraries.
Plan a change to this stack as its own task, test it in a branch, and adopt it only when the tests pass.

| Component | Suggestion | Version floor | Reason |
|---|---|---|---|
| Language | TypeScript, `strict` | 5.8 | Types at the boundary |
| Build | Vite | 7 | Fast dev server, one config for build and tests |
| Framework | Svelte 5 with runes | 5.30 | Compiled, small runtime, `mount` for vanilla libraries |
| Type check | `svelte-check` | 4 | Types in the components. Pin TypeScript to a version that it supports |
| Lint and format | Biome | 2 | One tool for lint and format |
| Unit tests | Vitest and `@testing-library/svelte` | 3 | Reuses the Vite config |
| End-to-end tests | Playwright | 1.50 | Real browsers, grid and dock behavior |
| Validation | valibot | 1 | Small, tree-shakable schemas |
| Router | A route module of your own | | Fewer than 80 lines, see `reference/client.md` |

The workstation profile adds these libraries:

| Component | Suggestion | Version floor | Reason |
|---|---|---|---|
| Docking | `dockview` | 8 | Tabs, groups, floating and popout panels, layout JSON, MIT. The vanilla package, not `dockview-core` |
| Grid | `@revolist/revogrid` | 4.30 | Web component, row and column virtualization, pinning, editing, MIT |
| Grid, alternative | SVAR Svelte DataGrid, `wx-svelte-grid` | 2 | Svelte-native, virtual scrolling, frozen columns, inline editing, MIT. The Svelte 5 line was in beta in October 2026 |
| Palette and overlays | `bits-ui` | 2.20 | Headless Command and Dialog for Svelte 5 |

Rules that hold for every framework:

- Use the conventional profile by default.
- Use the workstation profile when the main view is a large grid, or when the user works in several panels at the same time.
- One application can have both: a dashboard view in the conventional profile and a grid view in the workstation profile.
  The open view is a saved setting, and a row in the dashboard opens the record in the grid.
- Do not use the workstation profile for a public page.
  Its bundle is about 200 kB to 300 kB gzipped.
- Keep the conventional bundle under 100 kB gzipped.
- Do not add a data cache library such as TanStack Query.
  The WebSocket stream replaces cache invalidation.
- Do not add a CSS framework.
  Use CSS custom properties, see section 10.
- Do not use `any`.
  Check every API and WebSocket payload against a schema at the boundary.
- Put the realtime client, the settings store, the dialog rules, and the design tokens of sections 5 to 11 in the framework of the application.
  The reference files show them in Svelte, and the rules do not depend on Svelte.

Rules for the Svelte stack:

- Write the dockview adapter from `reference/workstation.md`.
  Do not use `dockview-svelte` or another community port.
- Before you add dockview or a grid to an existing application, build one view with it in a branch.
  Run the Playwright tests of section 13 on that view.
  Adopt the library only when the tests pass and the bundle stays in the limit.

## 4. Server layout

```text
cmd/app/main.go          Config, wiring, start, and shutdown
internal/api/            HTTP handlers, one file for each resource
internal/auth/           Sessions, local login, passkeys, OIDC
internal/realtime/       Hub, envelope, listener, watcher
internal/store/          Interfaces, and one package for each variant
internal/settings/       Per-user settings
internal/web/            Embedded client files, index template, fallback
internal/web/dist/       The Vite build output. `embed` reads only below its own package directory
web/                     The Vite project, with `build.outDir` set to `../internal/web/dist`
migrations/              SQL files for goose
```

- Register the API routes and `/ws` before the client fallback.
- Use method patterns, for example `GET /api/items/{id}`.
- Write middleware as `func(http.Handler) http.Handler`.
- Chain the middleware in this order: recover, request id, log, security headers, rate limit, session, cross-origin protection, handler.
- Serve `index.html` through a template, so that each response gets a fresh CSP nonce.
- In development, serve the client from `web/dist` on disk, or proxy to the Vite dev server.
- Read the config from flags with environment variable defaults.
  Do not read a config file.
- Stop with `signal.NotifyContext` and `Server.Shutdown`.
  Close the hub, the listener, and the pool after the server.
- Give `/healthz` for liveness, and `/readyz` for readiness with a store check.
- Log with `slog` in JSON.
  Give each request a request id, and put it in every log record of the request.
- Never log a session token, a password, a cookie, or a request body.

## 5. Realtime push

The server pushes changes to the clients over one WebSocket for each client.
The client never polls.

- The data store wakes the server.
  PostgreSQL uses a trigger and `pg_notify`.
  MongoDB uses a change stream.
  The `none` variant publishes on an in-process bus.
- The notification carries the table, the operation, and the id.
  It does not carry the row.
  The payload limit of `pg_notify` is 8000 bytes.
- Run the PostgreSQL listener on a dedicated connection, not on a pooled one.
  A transaction-mode pooler such as PgBouncer breaks `LISTEN`.
- Coalesce notifications.
  A burst of changes becomes one read of the changed rows.
- If a client must not miss a change, add an outbox table with a `bigserial` sequence.
  `NOTIFY` is the wake-up, and the outbox is the truth.
- Persist the MongoDB resume token after you processed the event, not before.
- Every message is a JSON envelope with the fields `v`, `type`, `topic`, `seq`, and `data`.
- A list endpoint answers `{seq, data}`.
  The handler reads `seq` from the hub before it reads the rows, so that no change between the two reads gets lost.
  The client sets `since` to that `seq` before it subscribes.
- The client subscribes to topics, and sends its last `seq` as `since`.
  The server answers with a snapshot and then deltas, or with `resync` when `since` is too old.
- The client applies a delta by id.
  A delta that arrives twice changes nothing.
- The hub never blocks on a slow client.
  If the send buffer of a client is full, close the connection.
- Send a ping every 30 seconds, and close a connection without a pong after 60 seconds.
- Limit a message to 64 KiB, and a client to one connection for each browser tab.
- The client reconnects with exponential backoff and jitter, from one second to 30 seconds.
  It sends its subscriptions again after each open.
- `reference/realtime.md` gives the trigger, the listener, the watcher, the hub, and the client module.

## 6. Authentication

The application supports local login, OIDC, or both.

- Use OIDC when the operator has an identity provider.
  Use local login when there is none.
- Store the session on the server with `scs`.
  The cookie holds only the random token.
- Set the session cookie with `HttpOnly`, `Secure`, `SameSite=Lax`, and `Path=/`.
- Renew the session token after login and after a change of privilege.
- Use an idle timeout of 12 hours and an absolute timeout of seven days, unless the task gives other values.
- Hash a local password with argon2id, 64 MiB of memory, three passes, and a 16-byte salt.
- Offer passkeys as the second local method.
  Store the credential id, the public key, the sign count, and the backup flags.
- For OIDC, use the authorization code flow with PKCE.
  Keep the `state` and the `nonce` in a short-lived cookie.
- Authenticate the WebSocket at the upgrade with the session cookie.
  Do not put a token in the URL.
- Rate limit the login routes by client address and, with a looser limit, by username.
- Bound the password checks in flight, because one argon2id check holds 64 MiB.
- Trust `X-Forwarded-For` only from a configured proxy list.
- Run the login checks of `reference/auth-security.md` after each change to the login.
- `reference/auth-security.md` gives the flows and the code.

## 7. Security

- Use `http.CrossOriginProtection` on every route, and keep `SameSite=Lax` on the cookie.
  The two together are the CSRF defense.
- Check the `Origin` header at the WebSocket upgrade against an allowlist.
  A missing `Origin` is a non-browser client, and it must have a valid session too.
- Send these headers on every response: `Content-Security-Policy`, `Strict-Transport-Security`, `X-Content-Type-Options: nosniff`, `Referrer-Policy: same-origin`, `Permissions-Policy`, and `X-Frame-Options: DENY`.
- Send `Cache-Control: no-store` on every API, auth, and WebSocket answer.
- The CSP uses `script-src 'nonce-...' 'strict-dynamic'`, `object-src 'none'`, `base-uri 'none'`, and `connect-src 'self'`.
  Do not use `unsafe-inline` or `unsafe-eval`.
- Test a new CSP with `Content-Security-Policy-Report-Only` first.
- Limit each request body with `http.MaxBytesReader`.
- Set `ReadHeaderTimeout`, `ReadTimeout`, `WriteTimeout`, and `IdleTimeout` on the server.
  Give the WebSocket handler its own deadlines.
- Open a file only through `os.Root` in the data directory.
- Serve TLS 1.3 with `crypto/tls`, or run behind a reverse proxy that does.
  Trust `X-Forwarded-For` only from the proxy address.
- Pin each dependency in `go.sum` and `package-lock.json`, and run `govulncheck` and `npm audit` in CI.

## 8. Data access

- PostgreSQL: write SQL in `.sql` files, and generate Go with `sqlc`.
  Use `pgxpool` for the application, and one `pgx.Conn` for the listener.
- Run the migrations at start with goose from `embed.FS`.
  Stop if a migration fails.
- Use `uuidv7()` for new ids, and `timestamptz` for every time.
- Use `RETURNING` with `OLD` and `NEW` when a handler needs both values.
- MongoDB: use typed structs with `bson` tags, and `bson.D` for pipelines.
  Do not use `bson.M` for a pipeline, because the key order matters.
- Run MongoDB as a replica set, also with one member.
  A standalone server has no change streams.
- The `none` variant: hold the state in memory under a mutex, and write each change to a JSON file with a temporary file and `os.Rename`.
- `reference/data.md` gives the schemas and the code.

## 9. Per-user state

A user gets the same filters, columns, and layout on every device.

- Store the state on the server, keyed by the user id.
  `localStorage` holds only state that can get lost, for example the open tab.
- Store the state as one JSON object with a `version` number.
  PostgreSQL: a `jsonb` column with `CHECK (jsonb_typeof(settings) = 'object')`.
  MongoDB: the `settings` subdocument.
  The `none` variant: one file for each user.
- Give `GET /api/me/settings` and `PUT /api/me/settings`.
  `PUT` sends the full object and the `version` that the client read.
  If the version does not match, answer 409 and the current object.
- Check the object against a valibot schema on the client, and against a Go struct on the server.
  Drop unknown keys.
- Debounce the writes on the client by one second.
  Write at once before the page unloads.
- Push a settings change to the other tabs and devices of the user over the WebSocket topic `me.settings`.
- Name the fields after the view, for example `inventory.filters`, `inventory.columns`, `dock.layout`, and `theme`.
- Keep the dock layout JSON under a `layout_version`, and offer a reset layout command.

## 10. Client rules

- Keep the code in `web/src/lib/` for modules, `web/src/routes/` for views, and `web/src/components/` for shared parts.
- Put the API client in one module, with one function for each route and a valibot schema for each response.
- Put the WebSocket client in one module.
  Views subscribe to topics and read stores.
- Put shared state in `.svelte.ts` modules with `$state` and `$derived`.
  Do not use the Svelte 4 store API in new code.
- Define the design tokens once on `:root`: colors, spacing, font sizes, row height, and radius.
  Redefine the colors under `prefers-color-scheme: dark`.
- Use `font-variant-numeric: tabular-nums` in every table and grid.
- Use `dvh` units for full-height layouts, and container queries inside a panel.
- Show a loading state and a failure state for each request.
  Never show a failed request as an empty list.
- If the application shows system state, load the installed `scuq-hmi-semantics` skill for the status colors, alert levels, and confirmations.
  This skill does not restate them.
- If the application has a chart, a stat tile, or a dashboard, load the `dataviz` skill when it is installed.
  Three of its rules matter most here.
  A severity is a status color with a label.
  A part-to-whole is a segmented bar, not a donut.
  A stat tile gets a sparkline with a hover value.
- The server keeps no time series unless the task asks for one.
  A sparkline then starts at page load, and the tile says so.
- Give the server a `-simulate` flag for development, which changes one record on an interval.
  The realtime path is then visible without a second client.
- `reference/client.md` gives the project layout, the configs, the API client, the router, and the settings store.

## 11. Modal dialogs

- Use the native `<dialog>` element with `showModal()`.
  Do not write a focus trap, an `aria-modal` attribute, or an Escape handler.
  The element has them.
- Give each dialog a title, and point `aria-labelledby` at it.
- Set `closedby="closerequest"` by default.
  Use `closedby="any"` for an informational dialog, and `closedby="none"` for a form with unsaved changes.
- Keep the unsaved-changes guard in the `cancel` event handler, because an old browser ignores `closedby`.
- Size the dialog with `width: min(90vw, 40rem)` and `max-height: 90dvh`.
  A confirmation uses `28rem`, and a form uses `40rem`.
- Keep the header and the footer fixed, and let only the body scroll.
- Set `overscroll-behavior: contain` on `dialog:modal`.
  Do not set `overflow: hidden` on `body`.
- Return the focus to the element that opened the dialog.
- Set the first focus on the title or the first field, not on a control at the end of a long body.
- Put the primary action on the right, and name it with one verb.
- One `Modal.svelte` component holds these rules.
  `reference/client.md` gives it.

## 12. The workstation profile

The workstation profile gives the application the behavior of a desktop program on a 2K or 4K monitor.

- The layout fills the window.
  There is no page container with a maximum width.
- Density is compact: 28 px rows, 12 px to 13 px text, and 4 px to 8 px spacing.
- `dockview` holds the panels.
  Each panel is a Svelte component in a registry, and the adapter mounts it with `mount()` and `unmount()`.
- The grid is `<revo-grid>`.
  Set its properties from `$effect`, and apply WebSocket deltas by row id.
- A persistent bulk-edit panel edits the selected rows.
  It sends one `PATCH` with the ids and the changed fields, and the server pushes the result.
- The command palette opens with Ctrl+K.
  Every panel action and every menu action is also a command.
  The palette is the keyboard path to the docking layout.
- Save the dock layout in the user settings under `dock.layout`, with a `layout_version`.
- Offer the commands reset layout, open panel, close panel, and focus panel.
- Test grid and dock behavior with Playwright.
  A unit test covers the stores and the adapters.
- `reference/workstation.md` gives the adapter, the grid wrapper, the palette, and the bulk-edit pattern.

## 13. Tests

- Server: `go test ./...` with `net/http/httptest` for the handlers and `testing/synctest` for backoff and hub timing.
- Server: run the store tests against a real database when `TEST_DATABASE_URL` or `TEST_MONGODB_URI` is set.
  Skip them without the variable.
- Server: test the WebSocket with a real `httptest.Server` and a `coder/websocket` client.
- Client: Vitest for schemas, stores, the API client, and the WebSocket module with a fake socket.
- Client: Playwright for the login, one view, one modal, the settings round trip, and a WebSocket update.
- Workstation: Playwright for the grid, the dock, the bulk edit, and the command palette.
- CI order: Biome, `tsc --noEmit`, Vitest, `go vet`, `govulncheck`, `go test`, build, Playwright.
- `reference/testing.md` gives the patterns.

## 14. Documentation

Each application has a `README.md` in the structure of the scraibe standard.
Load the scuq-scraibe skill before you write it.

The README gives:

- The variant, and the database version floor
- The config flags and environment variables
- The authentication methods
- The WebSocket topics and the envelope
- The user settings fields
- The build, the test, and the start commands

## 15. Review checklist

Check every item before you finish.

- `go vet ./...`, `go test ./...`, and `govulncheck ./...` pass.
- `go version -m BINARY` shows `CGO_ENABLED=0`.
- `biome check`, `tsc --noEmit`, `vitest run`, and `playwright test` pass.
- The API routes and `/ws` are registered before the client fallback.
- Every response has the security headers, and the CSP has a nonce and no `unsafe-inline`.
- `http.CrossOriginProtection` wraps every route.
- The session cookie has `HttpOnly`, `Secure`, and `SameSite=Lax`.
- The login checks of `reference/auth-security.md` pass.
- The WebSocket upgrade checks the `Origin` and the session.
- The listener runs on a dedicated connection, and the notification payload holds no row.
- The hub closes a slow client instead of a block.
- The client reconnects with backoff and jitter, and sends its subscriptions again.
- Every delta handler is idempotent on the id.
- Every API and WebSocket payload passes a valibot schema.
- The user settings have a version, and `PUT` answers 409 on a mismatch.
- Every dialog is a `<dialog>` with a title, `closedby`, and a scrolling body.
- The workstation layout saves and restores, and the reset layout command exists.
- No log record and no error message holds a secret.
- The README lists every flag, topic, and settings field.
