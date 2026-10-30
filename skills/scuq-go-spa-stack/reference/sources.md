# Sources

## BLUF

This file lists the sources of the rules in `SKILL.md` and the reference files, by topic.
Read a source when a rule needs more context, or when a version changed.
The research for the first version of the skill is from October 2026.

## Go

- Go 1.25 release notes, `http.CrossOriginProtection`, `testing/synctest`, `os.Root`: https://go.dev/doc/go1.25
- Go 1.26 release notes, Green Tea garbage collector, `slog.NewMultiHandler`: https://go.dev/doc/go1.26
- `net/http` routing patterns: https://pkg.go.dev/net/http#hdr-Patterns-ServeMux
- `http.CrossOriginProtection`: https://pkg.go.dev/net/http#CrossOriginProtection
- A modern approach to preventing CSRF in Go: https://www.alexedwards.net/blog/preventing-csrf-in-go
- `embed`: https://pkg.go.dev/embed
- `os.Root`: https://pkg.go.dev/os#Root
- `log/slog`: https://pkg.go.dev/log/slog
- OpenTelemetry slog bridge: https://pkg.go.dev/go.opentelemetry.io/contrib/bridges/otelslog

## WebSocket

- coder/websocket: https://pkg.go.dev/github.com/coder/websocket
- Go WebSocket library comparison: https://github.com/gosuda/go-websocket-libraries
- Cross-site WebSocket hijacking: https://blog.includesecurity.com/2025/04/cross-site-websocket-hijacking-exploitation-in-2025/
- OWASP WebSocket security cheat sheet: https://cheatsheetseries.owasp.org/cheatsheets/WebSocket_Security_Cheat_Sheet.html

## PostgreSQL

- PostgreSQL 18 release notes, `uuidv7()`, asynchronous I/O, `RETURNING OLD` and `NEW`: https://www.postgresql.org/docs/release/18.0/
- Versioning policy and current minor versions: https://www.postgresql.org/support/versioning/
- `NOTIFY`, the 8000-byte payload limit: https://www.postgresql.org/docs/current/sql-notify.html
- PostgreSQL cookbook, LISTEN/NOTIFY with an outbox: https://dsbasko.github.io/postgres-cookbook/en/09-writes-eventing-and-server-logic/09-04-listen-notify/
- pgx v5: https://pkg.go.dev/github.com/jackc/pgx/v5
- pgxpool: https://pkg.go.dev/github.com/jackc/pgx/v5/pgxpool
- sqlc: https://docs.sqlc.dev/
- goose: https://pkg.go.dev/github.com/pressly/goose/v3
- JSONB for user preferences: https://www.postgresql.org/docs/current/datatype-json.html

## MongoDB

- MongoDB release notes: https://www.mongodb.com/docs/manual/release-notes/
- Change streams: https://www.mongodb.com/docs/manual/changestreams/
- Go driver v2: https://pkg.go.dev/go.mongodb.org/mongo-driver/v2/mongo
- Go driver change streams: https://www.mongodb.com/docs/drivers/go/current/fundamentals/crud/read-operations/changestream/
- Resume token in a pipeline, SERVER-37669: https://jira.mongodb.com/browse/SERVER-37669

## Authentication

- scs: https://pkg.go.dev/github.com/alexedwards/scs/v2
- argon2: https://pkg.go.dev/golang.org/x/crypto/argon2
- OWASP password storage cheat sheet: https://cheatsheetseries.owasp.org/cheatsheets/Password_Storage_Cheat_Sheet.html
- go-webauthn: https://pkg.go.dev/github.com/go-webauthn/webauthn/webauthn
- go-oidc: https://pkg.go.dev/github.com/coreos/go-oidc/v3/oidc
- OAuth 2.0 for browser-based applications: https://datatracker.ietf.org/doc/html/draft-ietf-oauth-browser-based-apps
- `golang.org/x/time/rate`: https://pkg.go.dev/golang.org/x/time/rate

## Security headers

- Content Security Policy, strict CSP: https://web.dev/articles/strict-csp
- CSP on MDN: https://developer.mozilla.org/en-US/docs/Web/HTTP/CSP
- OWASP HTTP headers cheat sheet: https://cheatsheetseries.owasp.org/cheatsheets/HTTP_Headers_Cheat_Sheet.html
- OWASP session management cheat sheet: https://cheatsheetseries.owasp.org/cheatsheets/Session_Management_Cheat_Sheet.html

## Client

- Svelte 5, runes: https://svelte.dev/docs/svelte/what-are-runes
- Svelte 5, `mount` and `unmount`: https://svelte.dev/docs/svelte/imperative-component-api
- Vite: https://vite.dev/guide/
- Vitest: https://vitest.dev/guide/
- Testing Library for Svelte: https://testing-library.com/docs/svelte-testing-library/intro/
- Playwright: https://playwright.dev/docs/intro
- Biome: https://biomejs.dev/guides/getting-started/
- valibot: https://valibot.dev/guides/introduction/
- The `dialog` element, Chrome guidance with `closedby`: https://developer.chrome.com/docs/modern-web-guidance/guides/ui-components/dialog
- The `dialog` element on MDN: https://developer.mozilla.org/en-US/docs/Web/HTML/Element/dialog

## Workstation

- dockview-core: https://dockview.dev/docs/overview/introduction
- dockview API and theme variables: https://dockview.dev/typedocs
- RevoGrid: https://rv-grid.com/guide/
- RevoGrid for Svelte: https://rv-grid.com/svelte-data-grid
- Bits UI: https://bits-ui.com/docs/introduction
- Bits UI Command: https://bits-ui.com/docs/components/command
- Bits UI Dialog: https://bits-ui.com/docs/components/dialog
