# Authentication and security

## BLUF

This file gives the code for sections 6 and 7 of the skill.
It covers the session manager, the user middleware, local login with argon2id, passkeys, the OIDC flow, the rate limiter, and the cookie rules.
The session lives on the server, and the cookie holds only a random token.
Choose OIDC when the operator has an identity provider, and local login when there is none.

## Sessions

```go
package auth

import (
	"net/http"
	"time"

	"github.com/alexedwards/scs/pgxstore"
	"github.com/alexedwards/scs/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

func NewSessionManager(pool *pgxpool.Pool, secure bool) *scs.SessionManager {
	sm := scs.New()
	sm.Store = pgxstore.New(pool)
	sm.IdleTimeout = 12 * time.Hour
	sm.Lifetime = 7 * 24 * time.Hour
	sm.Cookie.Name = "session"
	sm.Cookie.HttpOnly = true
	sm.Cookie.Secure = secure
	sm.Cookie.SameSite = http.SameSiteLaxMode
	sm.Cookie.Path = "/"
	return sm
}
```

- Use `pgxstore` for PostgreSQL, `mongodbstore` for MongoDB, and `memstore` for the `none` variant.
  `memstore` loses the sessions at restart, and the README says so.
- `Secure` is false only on `localhost` in development.
- `SameSite=Lax` lets the OIDC redirect from the identity provider carry the cookie.
  `Strict` breaks that redirect.
- `pgxstore` creates the table `sessions` when it does not exist.
  Put the `CREATE TABLE` in a migration instead, so that the application role needs no DDL right.

## The user middleware

```go
type User struct {
	ID    string
	Name  string
	Roles []string
}

type userKey struct{}

// RequireUser answers 401 without a session user. It puts the user in the
// context for the handler.
func RequireUser(sm *scs.SessionManager, users UserStore) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			id := sm.GetString(r.Context(), "user_id")
			if id == "" {
				http.Error(w, `{"error":"login is necessary"}`, http.StatusUnauthorized)
				return
			}
			u, err := users.ByID(r.Context(), id)
			if err != nil {
				sm.Destroy(r.Context())
				http.Error(w, `{"error":"login is necessary"}`, http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), userKey{}, u)))
		})
	}
}

func UserFrom(ctx context.Context) User {
	u, _ := ctx.Value(userKey{}).(User)
	return u
}
```

- Read the user on every request.
  A user that an administrator disabled loses access at the next request, not at the session end.
- The client answers a 401 with the login view, and keeps the current route for the return.

## Local login

```go
import "golang.org/x/crypto/argon2"

const (
	argonTime    = 3
	argonMemory  = 64 * 1024 // KiB
	argonThreads = 2
	argonKeyLen  = 32
)

// HashPassword returns the PHC string for the password with a new salt.
func HashPassword(password string) (string, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	key := argon2.IDKey([]byte(password), salt, argonTime, argonMemory, argonThreads, argonKeyLen)
	return fmt.Sprintf("$argon2id$v=19$m=%d,t=%d,p=%d$%s$%s",
		argonMemory, argonTime, argonThreads,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(key)), nil
}

// VerifyPassword reports whether password matches the PHC string. It runs in
// constant time over the key.
func VerifyPassword(phc, password string) bool {
	parts := strings.Split(phc, "$")
	if len(parts) != 6 || parts[1] != "argon2id" || parts[2] != "v=19" {
		return false
	}
	var m, t uint32
	var p uint8
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &m, &t, &p); err != nil {
		return false
	}
	salt, err1 := base64.RawStdEncoding.DecodeString(parts[4])
	key, err2 := base64.RawStdEncoding.DecodeString(parts[5])
	if err1 != nil || err2 != nil || len(key) == 0 {
		return false
	}
	got := argon2.IDKey([]byte(password), salt, t, m, p, uint32(len(key)))
	return subtle.ConstantTimeCompare(got, key) == 1
}
```

- Split the PHC string on `$`.
  `fmt.Sscanf` with `%s` reads to a space, not to a `$`, and a test with a real hash shows the defect.

The login handler:

1. Read `{"username", "password"}` with `readJSON`.
2. Take a token from the rate limiter for the client address and for the username.
   Without a token, answer 429.
3. Read the user by username.
   If the user does not exist, verify the password against a fixed dummy hash, so that the time does not reveal the user.
4. If the password does not match, answer 401 with the same message as for an unknown user.
5. Call `sm.RenewToken`, then `sm.Put(ctx, "user_id", u.ID)`.
6. Answer 204.

- Store the hash in the `users` table, column `password_hash text`.
  A user with OIDC only has `NULL` there.
- Give a password change route that needs the current password, and that renews the token.
- Do not add a password reset by email unless the task asks for it.

## Passkeys

```go
import "github.com/go-webauthn/webauthn/webauthn"

func NewWebAuthn(origin, displayName string) (*webauthn.WebAuthn, error) {
	u, err := url.Parse(origin)
	if err != nil {
		return nil, err
	}
	return webauthn.New(&webauthn.Config{
		RPID:          u.Hostname(),
		RPDisplayName: displayName,
		RPOrigins:     []string{origin},
	})
}
```

- `RPID` is the host name of the page, and `RPOrigins` holds the full origin.
  A mismatch is the usual failure at the first test.
- The registration and the login each have two routes: `begin` and `finish`.
  `begin` stores the `SessionData` in the scs session under `webauthn`, and `finish` reads it back.
- Store for each credential the id, the public key, the sign count, the AAGUID, the `BackupEligible` flag, and the `BackupState` flag.
- For a passkey login, call `BeginDiscoverableLogin`.
  The browser then lets the user pick the credential.
- Implement `webauthn.User` on your user type.
  `WebAuthnID` returns a stable random id of 64 bytes, not the username.

## OIDC

```go
import (
	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

type OIDC struct {
	provider *oidc.Provider
	config   oauth2.Config
	verifier *oidc.IDTokenVerifier
}

func NewOIDC(ctx context.Context, issuer, clientID, clientSecret, redirectURL string) (*OIDC, error) {
	p, err := oidc.NewProvider(ctx, issuer)
	if err != nil {
		return nil, err
	}
	return &OIDC{
		provider: p,
		config: oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			Endpoint:     p.Endpoint(),
			RedirectURL:  redirectURL,
			Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
		},
		verifier: p.Verifier(&oidc.Config{ClientID: clientID}),
	}, nil
}
```

The flow:

1. `GET /auth/oidc/start`: make a random `state`, a random `nonce`, and a PKCE verifier with `oauth2.GenerateVerifier`.
   Store the three in the scs session, and redirect to `config.AuthCodeURL(state, oidc.Nonce(nonce), oauth2.S256ChallengeOption(verifier))`.
2. `GET /auth/oidc/callback`: compare the `state` query value with the session value.
   On a mismatch, answer 400.
3. Exchange the code with `config.Exchange(ctx, code, oauth2.VerifierOption(verifier))`.
4. Read `id_token` from the token extras, and call `verifier.Verify`.
   Compare the `nonce` claim with the session value.
5. Read the claims `sub`, `email`, and `name`.
   Find the user by issuer and `sub`, or create the user when the config permits it.
6. Call `sm.RenewToken`, `sm.Put(ctx, "user_id", u.ID)`, and remove the `state`, `nonce`, and verifier.
7. Redirect to the route that the client stored before the login.

- Key the user on `(issuer, sub)`, not on the email.
  An email can change.
- Give `POST /auth/logout`, which calls `sm.Destroy`.
  Add the `end_session_endpoint` redirect of the provider when the task asks for a full logout.
- A redirect URL must be an absolute path on the own host.
  Reject every other value.

## Rate limit

```go
import "golang.org/x/time/rate"

type Limiter struct {
	mu    sync.Mutex
	items map[string]*entry
	r     rate.Limit
	b     int
}

type entry struct {
	lim  *rate.Limiter
	seen time.Time
}

// Allow reports whether key has a token. It keeps one token bucket for each
// key, and a sweeper removes keys idle for ten minutes.
func (l *Limiter) Allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	e, ok := l.items[key]
	if !ok {
		e = &entry{lim: rate.NewLimiter(l.r, l.b)}
		l.items[key] = e
	}
	e.seen = time.Now()
	return e.lim.Allow()
}
```

- Login: five tries per minute for each address, and ten per minute for each username.
- WebSocket connect: ten per minute for each user.
- API: 60 requests per second for each user, for the default.
- Read the client address from `r.RemoteAddr`, or from `X-Forwarded-For` only when the request comes from the proxy address.

## Cookie and header rules

| Item | Value |
|---|---|
| Session cookie | `HttpOnly; Secure; SameSite=Lax; Path=/`, no `Domain` |
| CSRF | `http.CrossOriginProtection` on every route, plus the `SameSite` cookie |
| Body limit | 1 MiB for JSON, 64 KiB for a WebSocket message |
| TLS | 1.3 minimum, `tls.Config{MinVersion: tls.VersionTLS13}` |
| HSTS | `max-age=63072000; includeSubDomains` |
| `Origin` on `/ws` | Allowlist of host names, same host by default |

- A request without `Origin` and without `Sec-Fetch-Site` passes `CrossOriginProtection`.
  Such a request is a non-browser client, and it still needs the session cookie.
- Never put a secret in a URL, a log record, or an error message.
- Read secrets from the environment or from a file in the data directory with mode 0600.
  Do not read them from a flag, because `ps` shows the flags.
