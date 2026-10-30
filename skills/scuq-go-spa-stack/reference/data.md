# Data access

## BLUF

This file gives the code for sections 8 and 9 of the skill.
It covers the store interface, the PostgreSQL schema with sqlc and goose, and the MongoDB patterns.
It also covers the file store of the `none` variant, and the user settings in each variant.
One Go interface for each entity hides the variant from the handlers.

## The store interface

```go
package store

type Item struct {
	ID        string    `json:"id" bson:"_id"`
	Name      string    `json:"name" bson:"name"`
	Status    string    `json:"status" bson:"status"`
	UpdatedAt time.Time `json:"updatedAt" bson:"updatedAt"`
}

type Items interface {
	List(ctx context.Context) ([]Item, error)
	Get(ctx context.Context, id string) (Item, error)
	Patch(ctx context.Context, ids []string, patch ItemPatch) ([]Item, error)
}

type Settings interface {
	Get(ctx context.Context, userID string) (json.RawMessage, int, error)
	// Put stores the object when version matches the stored version. It
	// returns ErrVersion otherwise.
	Put(ctx context.Context, userID string, settings json.RawMessage, version int) (int, error)
}

var ErrNotFound = errors.New("not found")
var ErrVersion = errors.New("version mismatch")
```

- A handler maps `ErrNotFound` to 404 and `ErrVersion` to 409.
- The JSON tags give the wire format.
  The bson tags give the MongoDB format.
  Both use lower camel case.

## PostgreSQL

### Schema

```sql
-- migrations/0001_init.sql
-- +goose Up
CREATE TABLE users (
    id            uuid PRIMARY KEY DEFAULT uuidv7(),
    username      text UNIQUE,
    password_hash text,
    oidc_issuer   text,
    oidc_subject  text,
    created_at    timestamptz NOT NULL DEFAULT now(),
    UNIQUE (oidc_issuer, oidc_subject)
);

CREATE TABLE user_settings (
    user_id  uuid PRIMARY KEY REFERENCES users ON DELETE CASCADE,
    version  integer NOT NULL DEFAULT 0,
    settings jsonb NOT NULL DEFAULT '{}'::jsonb CHECK (jsonb_typeof(settings) = 'object'),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE items (
    id         uuid PRIMARY KEY DEFAULT uuidv7(),
    name       text NOT NULL,
    status     text NOT NULL DEFAULT 'active',
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE sessions (
    token  text PRIMARY KEY,
    data   bytea NOT NULL,
    expiry timestamptz NOT NULL
);
CREATE INDEX sessions_expiry ON sessions (expiry);

-- +goose Down
DROP TABLE sessions, items, user_settings, users;
```

- `uuidv7()` is a function of PostgreSQL 18.
  Time-ordered ids keep the primary key index compact.
- Every time column is `timestamptz`.
- The `sessions` table is the one that `pgxstore` expects.

### Queries with sqlc

```yaml
# sqlc.yaml
version: "2"
sql:
  - engine: postgresql
    schema: migrations
    queries: internal/store/postgres/queries
    gen:
      go:
        package: pgq
        out: internal/store/postgres/pgq
        sql_package: pgx/v5
        emit_json_tags: true
        emit_pointers_for_null_types: true
```

```sql
-- internal/store/postgres/queries/settings.sql

-- name: GetSettings :one
SELECT settings, version FROM user_settings WHERE user_id = $1;

-- name: PutSettings :one
INSERT INTO user_settings (user_id, version, settings, updated_at)
VALUES ($1, 1, $2, now())
ON CONFLICT (user_id) DO UPDATE
SET settings = EXCLUDED.settings,
    version = user_settings.version + 1,
    updated_at = now()
WHERE user_settings.version = $3
RETURNING version;

-- name: PatchItems :many
UPDATE items
SET name = coalesce(sqlc.narg('name'), name),
    status = coalesce(sqlc.narg('status'), status),
    updated_at = now()
WHERE id = ANY(sqlc.arg('ids')::uuid[])
RETURNING *;
```

- `PutSettings` returns no row when the version does not match.
  The store maps `pgx.ErrNoRows` to `ErrVersion`.
- Run `sqlc generate` in CI, and fail when the output differs from the commit.

### Pool and migrations

```go
import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pressly/goose/v3"
)

//go:embed migrations/*.sql
var migrations embed.FS

func Open(ctx context.Context, url string) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(url)
	if err != nil {
		return nil, err
	}
	cfg.MaxConns = 10
	cfg.MaxConnLifetime = time.Hour
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, err
	}
	return pool, pool.Ping(ctx)
}

func Migrate(ctx context.Context, url string) error {
	db, err := sql.Open("pgx", url)
	if err != nil {
		return err
	}
	defer db.Close()
	goose.SetBaseFS(migrations)
	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}
	return goose.UpContext(ctx, db, "migrations")
}
```

- goose uses `database/sql`.
  Import `github.com/jackc/pgx/v5/stdlib` for the `pgx` driver name.
- Run `Migrate` before `Open` at start.
  Stop the process when it fails.
- Give the application role `INSERT`, `UPDATE`, `DELETE`, and `SELECT`.
  Give the DDL right to the migration role only when the operator separates them.

## MongoDB

```go
import (
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func Open(ctx context.Context, uri string) (*mongo.Client, error) {
	client, err := mongo.Connect(options.Client().ApplyURI(uri).SetMaxPoolSize(20))
	if err != nil {
		return nil, err
	}
	return client, client.Ping(ctx, nil)
}

// Put stores settings when version matches. It returns ErrVersion otherwise.
func (s *Settings) Put(ctx context.Context, userID string, settings json.RawMessage, version int) (int, error) {
	var doc bson.D
	if err := bson.UnmarshalExtJSON(settings, true, &doc); err != nil {
		return 0, err
	}
	res := s.users.FindOneAndUpdate(ctx,
		bson.D{{Key: "_id", Value: userID}, {Key: "settingsVersion", Value: version}},
		bson.D{{Key: "$set", Value: bson.D{{Key: "settings", Value: doc}}}, {Key: "$inc", Value: bson.D{{Key: "settingsVersion", Value: 1}}}},
		options.FindOneAndUpdate().SetReturnDocument(options.After).SetProjection(bson.D{{Key: "settingsVersion", Value: 1}}),
	)
	var out struct {
		Version int `bson:"settingsVersion"`
	}
	if err := res.Decode(&out); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return 0, store.ErrVersion
		}
		return 0, err
	}
	return out.Version, nil
}
```

- Use `bson.D` for every filter, update, and pipeline.
  `bson.M` loses the key order, and the order matters in a pipeline.
- Create the indexes at start with `Indexes().CreateMany`.
  An index that exists is not an error.
- Create a unique index on `(oidcIssuer, oidcSubject)` and on `username`.
- Use a schema validator on the collection with `$jsonSchema`, so that a stray document does not break the decode.
- The connection string names the replica set, for example `mongodb://db.cyberdyne.example:27017/app?replicaSet=rs0`.

## The none variant

```go
package file

// Store holds the state in memory and writes each change to one JSON file
// for each entity type. It is safe for use from several goroutines.
type Store struct {
	mu    sync.Mutex
	root  *os.Root
	items map[string]store.Item
	hub   Publisher
}

func Open(dir string) (*Store, error) {
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, err
	}
	root, err := os.OpenRoot(dir)
	if err != nil {
		return nil, err
	}
	s := &Store{root: root, items: map[string]store.Item{}}
	return s, s.load("items.json", &s.items)
}

// save writes v to name through a temporary file and a rename, so that a
// crash leaves the old file or the new file, never a partial file.
func (s *Store) save(name string, v any) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	tmp := name + ".tmp"
	f, err := s.root.OpenFile(tmp, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)
	if err != nil {
		return err
	}
	if _, err := f.Write(data); err != nil {
		f.Close()
		return err
	}
	if err := f.Sync(); err != nil {
		f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	return s.root.Rename(tmp, name)
}
```

- `os.Root` keeps every path inside the data directory, also a path with `..` or a symbolic link.
- User settings live in `settings/<user id>.json`, with the version in the object.
- The store calls `hub.Publish` after `save` succeeds, under the mutex.
- The README of the application says that the data directory is the only state, and that a backup is a copy of it.
