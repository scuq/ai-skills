# The Go server

## BLUF

This file gives the code for the server of section 4 of the skill.
It covers the start, the config, the middleware chain, the embedded client with a CSP nonce, the health routes, and the shutdown.
Copy the parts, and change the names.
Every snippet compiles with Go 1.27 and the standard library, except where the import names a module.

## Start and shutdown

`cmd/app/main.go` reads the config, builds the dependencies, and runs the server until a signal arrives.

```go
package main

import (
	"context"
	"errors"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var version = "dev"

type config struct {
	addr        string
	dataDir     string
	databaseURL string
	origins     []string
	devClient   string
}

// loadConfig reads the flags. Each flag takes its default from an
// environment variable, so a container and a shell give the same result.
func loadConfig() config {
	var c config
	flag.StringVar(&c.addr, "addr", env("APP_ADDR", ":8080"), "listen address")
	flag.StringVar(&c.dataDir, "data-dir", env("APP_DATA_DIR", "./data"), "data directory")
	flag.StringVar(&c.databaseURL, "database-url", env("APP_DATABASE_URL", ""), "PostgreSQL or MongoDB URL")
	flag.StringVar(&c.devClient, "dev-client", env("APP_DEV_CLIENT", ""), "Vite dev server URL, empty in production")
	flag.Func("origin", "allowed browser origin, repeatable", func(v string) error {
		c.origins = append(c.origins, v)
		return nil
	})
	flag.Parse()
	return c
}

func env(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func main() {
	cfg := loadConfig()
	log := slog.New(slog.NewJSONHandler(os.Stdout, nil)).With("version", version)
	slog.SetDefault(log)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := run(ctx, cfg, log); err != nil {
		log.Error("server stopped with an error", "err", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, cfg config, log *slog.Logger) error {
	// Build the store, the hub, the session manager, and the handlers here.
	handler := newHandler(cfg, log)

	srv := &http.Server{
		Addr:              cfg.addr,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       120 * time.Second,
		BaseContext:       func(net.Listener) context.Context { return ctx },
	}

	errc := make(chan error, 1)
	go func() { errc <- srv.ListenAndServe() }()
	log.Info("listen", "addr", cfg.addr)

	select {
	case err := <-errc:
		return err
	case <-ctx.Done():
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil && !errors.Is(err, context.DeadlineExceeded) {
		return err
	}
	// Close the hub, the listener, and the pool after the server, in this order.
	return nil
}
```

- `WriteTimeout` does not apply to a hijacked WebSocket connection.
  The WebSocket handler sets its own deadlines, see `realtime.md`.
- Go sets no `Server` header.
  A reverse proxy in front adds its own, so turn it off there.
- The other server fields stay at their defaults, for example `MaxHeaderBytes`.
  Change one only with a measurement that asks for it.
- `BaseContext` gives every request the server context.
  A handler that starts work for the whole process watches this context.
- Add `net` to the imports for `net.Listener`.

## The middleware chain

```go
type middleware func(http.Handler) http.Handler

// chain applies the middleware in the order of the arguments. The first
// argument sees the request first.
func chain(h http.Handler, mws ...middleware) http.Handler {
	for i := len(mws) - 1; i >= 0; i-- {
		h = mws[i](h)
	}
	return h
}

func newHandler(cfg config, log *slog.Logger) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", healthz)
	mux.HandleFunc("GET /readyz", readyz)
	mux.Handle("GET /api/items", requireUser(listItems))
	mux.Handle("GET /api/items/{id}", requireUser(getItem))
	mux.Handle("PATCH /api/items", requireUser(patchItems))
	mux.Handle("GET /api/me/settings", requireUser(getSettings))
	mux.Handle("PUT /api/me/settings", requireUser(putSettings))
	mux.Handle("GET /ws", requireUser(serveWS))
	mux.Handle("/", clientHandler(cfg))

	csrf := http.NewCrossOriginProtection()
	for _, o := range cfg.origins {
		if err := csrf.AddTrustedOrigin(o); err != nil {
			log.Error("origin is not valid", "origin", o, "err", err)
		}
	}

	return chain(mux,
		recoverPanic(log),
		requestID,
		logRequest(log),
		securityHeaders,
		rateLimit,
		sessionManager.LoadAndSave,
		csrf.Handler,
	)
}
```

- `requireUser` reads the session, and answers 401 without a user.
  `auth-security.md` gives it.
- `http.NewCrossOriginProtection` rejects a state-changing request from another origin.
  It reads `Sec-Fetch-Site` and `Origin`, and it lets a request without both headers through.
  The session cookie with `SameSite=Lax` covers the browsers without `Sec-Fetch-Site`.
- `sessionManager.LoadAndSave` comes from `scs`.
  It must run before the handlers and before `csrf.Handler`, so that a 403 does not lose the session.

## Request id and log

```go
type ctxKey int

const requestIDKey ctxKey = 0

func requestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := rand.Text()
		w.Header().Set("X-Request-Id", id)
		ctx := context.WithValue(r.Context(), requestIDKey, id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func logRequest(log *slog.Logger) middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(sw, r)
			log.InfoContext(r.Context(), "request",
				"id", r.Context().Value(requestIDKey),
				"method", r.Method,
				"path", r.URL.Path,
				"status", sw.status,
				"ms", time.Since(start).Milliseconds(),
			)
		})
	}
}

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

// Unwrap lets http.ResponseController reach the hijacker for the WebSocket.
func (w *statusWriter) Unwrap() http.ResponseWriter { return w.ResponseWriter }
```

- `rand.Text` is `crypto/rand.Text`, available since Go 1.24.
- A wrapper around `http.ResponseWriter` must have `Unwrap`.
  Without it, the WebSocket upgrade cannot hijack the connection.
- Do not log the query string, a header, or the body.

## Security headers and the CSP nonce

```go
const nonceKey ctxKey = 1

func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nonce := rand.Text()
		h := w.Header()
		h.Set("Content-Security-Policy",
			"default-src 'self'; script-src 'nonce-"+nonce+"' 'strict-dynamic'; "+
				"style-src 'self'; img-src 'self' data:; connect-src 'self'; "+
				"object-src 'none'; base-uri 'none'; frame-ancestors 'none'; form-action 'self'")
		h.Set("Strict-Transport-Security", "max-age=63072000; includeSubDomains")
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("Referrer-Policy", "same-origin")
		h.Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
		h.Set("Cross-Origin-Opener-Policy", "same-origin")
		ctx := context.WithValue(r.Context(), nonceKey, nonce)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
```

- Vite writes `<script type="module" crossorigin src="/assets/index-HASH.js">` into `index.html`.
  The template adds `nonce="{{.Nonce}}"` to that tag, and `strict-dynamic` trusts the modules that it imports.
- A style attribute from a Svelte transition needs `style-src 'self' 'unsafe-inline'` or a hash.
  Prefer CSS classes, and add `'unsafe-inline'` for styles only when a library needs it.
- Add `ws:` or `wss:` to `connect-src` only when the WebSocket host differs from the page host.

## The embedded client

```go
package web

import (
	"embed"
	"html/template"
	"io/fs"
	"net/http"
	"os"
	"path"
	"strings"
)

//go:embed all:dist
var dist embed.FS

// Handler serves the client. With devDir, it reads the files from disk on
// every request, so that a Vite build shows at once.
func Handler(devDir string, nonce func(*http.Request) string) http.Handler {
	var files fs.FS
	if devDir != "" {
		files = os.DirFS(devDir)
	} else {
		sub, err := fs.Sub(dist, "dist")
		if err != nil {
			panic(err)
		}
		files = sub
	}
	static := http.FileServerFS(files)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p := strings.TrimPrefix(path.Clean(r.URL.Path), "/")
		if p != "" && p != "index.html" {
			if f, err := files.Open(p); err == nil {
				f.Close()
				if strings.HasPrefix(p, "assets/") {
					w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
				}
				static.ServeHTTP(w, r)
				return
			}
		}
		serveIndex(w, r, files, nonce(r))
	})
}

func serveIndex(w http.ResponseWriter, r *http.Request, files fs.FS, nonce string) {
	src, err := fs.ReadFile(files, "index.html")
	if err != nil {
		http.Error(w, "client is not built", http.StatusInternalServerError)
		return
	}
	// Vite emits the script tag without a nonce. Add it.
	html := strings.ReplaceAll(string(src), "<script type=\"module\"", "<script type=\"module\" nonce=\"{{.Nonce}}\"")
	tmpl, err := template.New("index").Parse(html)
	if err != nil {
		http.Error(w, "index template is not valid", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	tmpl.Execute(w, struct{ Nonce string }{nonce})
}
```

- Every path that is not a file falls back to `index.html`.
  The client router then shows the view, or a not-found view.
- `embed` reads only files below the package directory.
  Set `build.outDir` in `vite.config.ts` to `../internal/web/dist`, and commit a `.gitkeep` there, so that `go build` works before the first client build.
- Parse the template once at start in production.
  The snippet parses on every request, so that the development mode sees a new build.
- Use `clientHandler(cfg)` in `newHandler` to wrap `web.Handler(cfg.devClient, nonceFrom)`.
  In development with a Vite dev server, replace it with `httputil.NewSingleHostReverseProxy`.
- Put a `popout.html` with an empty body in `web/public/`.
  A dockview popout window loads it, and the fallback would load the whole application instead.
- `WriteTimeout` does not apply to a hijacked WebSocket connection, but `ReadTimeout` does not either.
  The hub sets its own deadlines.

## Health routes

```go
func healthz(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}

// readyz answers 503 while the store does not answer. The orchestrator
// then keeps the traffic away without a restart.
func readyz(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()
	if err := store.Ping(ctx); err != nil {
		http.Error(w, "store does not answer", http.StatusServiceUnavailable)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
```

## JSON helpers

```go
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		slog.Warn("response encode failed", "err", err)
	}
}

// readJSON decodes one JSON object of at most 1 MiB, and rejects unknown
// fields and a second value.
func readJSON(w http.ResponseWriter, r *http.Request, v any) error {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(v); err != nil {
		return err
	}
	if dec.More() {
		return errors.New("request body has more than one value")
	}
	return nil
}
```

- Answer an error as `{"error": "text", "request_id": "..."}`.
  The text says what happened, and it holds no internal detail.
- Use `http.StatusUnprocessableEntity` for a validation error, and `http.StatusConflict` for a version mismatch.

## Build

```sh
cd web && npm ci && npm run build && cd ..   # writes internal/web/dist
CGO_ENABLED=0 go build -trimpath -ldflags "-s -w -X main.version=$(cat VERSION)" -o dist/app ./cmd/app
go version -m dist/app | grep CGO_ENABLED
```

Build for `linux/amd64` and `linux/arm64` with `GOOS` and `GOARCH`.
