# Realtime push

## BLUF

This file gives the code for section 5 of the skill.
It covers the PostgreSQL trigger and listener, the outbox, the MongoDB watcher, the in-process bus, the hub, the envelope, and the client module.
The data store wakes the server, the server reads the changed rows, and the hub pushes them to the subscribed clients.
The client never polls.

## The envelope

Every message in both directions is one JSON object.

```ts
export type Envelope =
  | { v: 1; type: "subscribe"; topic: string; since?: number }
  | { v: 1; type: "unsubscribe"; topic: string }
  | { v: 1; type: "snapshot"; topic: string; seq: number; data: unknown[] }
  | { v: 1; type: "delta"; topic: string; seq: number; data: Delta }
  | { v: 1; type: "resync"; topic: string }
  | { v: 1; type: "error"; code: string; message: string };

export type Delta = {
  op: "upsert" | "delete";
  id: string;
  row?: unknown;
};
```

```go
package realtime

type Envelope struct {
	V     int             `json:"v"`
	Type  string          `json:"type"`
	Topic string          `json:"topic,omitempty"`
	Seq   int64           `json:"seq,omitempty"`
	Since *int64          `json:"since,omitempty"`
	Data  json.RawMessage `json:"data,omitempty"`
	Code  string          `json:"code,omitempty"`
	Msg   string          `json:"message,omitempty"`
}
```

- `v` is the protocol version.
  Change it only when an old client cannot read the new messages.
- A topic names a view or a resource, for example `inventory`, `item.7f3a`, or `me.settings`.
- `seq` grows by one for each delta of a topic.
  The client stores the last `seq` and sends it as `since` when it subscribes again.
- If the server cannot replay from `since`, it answers `resync`, and the client requests a snapshot.
- A `delta` with `op: "upsert"` carries the full row.
  The client replaces the row with the same id, or adds it.

## The snapshot

```go
// ListItems answers the rows with the sequence number of the topic. It reads
// the number first, so that a change between the two reads repeats as a delta
// instead of a gap. The client ignores a repeated delta.
func (s *Service) ListItems(w http.ResponseWriter, r *http.Request) {
	seq := s.Hub.Seq("inventory")
	items, err := s.Items.List(r.Context())
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "the inventory could not be read")
		return
	}
	WriteJSON(w, http.StatusOK, struct {
		Seq  int64 `json:"seq"`
		Data any   `json:"data"`
	}{seq, items})
}
```

- The client calls `socket.setSince(topic, seq)` with the snapshot, and then subscribes.
- A `resync` envelope makes the view call the list endpoint again.

## PostgreSQL: trigger and notify

```sql
-- migrations/0002_changes.sql
-- +goose Up
CREATE TABLE outbox (
    seq        bigserial PRIMARY KEY,
    topic      text        NOT NULL,
    op         text        NOT NULL CHECK (op IN ('upsert', 'delete')),
    row_id     text        NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX outbox_topic_seq ON outbox (topic, seq);

CREATE FUNCTION notify_change() RETURNS trigger
LANGUAGE plpgsql AS $$
DECLARE
    v_op  text;
    v_id  text;
BEGIN
    IF TG_OP = 'DELETE' THEN
        v_op := 'delete';
        v_id := OLD.id::text;
    ELSE
        v_op := 'upsert';
        v_id := NEW.id::text;
    END IF;
    INSERT INTO outbox (topic, op, row_id) VALUES (TG_ARGV[0], v_op, v_id);
    PERFORM pg_notify('app_changes', TG_ARGV[0]);
    RETURN NULL;
END;
$$;

CREATE TRIGGER items_notify
AFTER INSERT OR UPDATE OR DELETE ON items
FOR EACH ROW EXECUTE FUNCTION notify_change('inventory');

-- +goose Down
DROP TRIGGER items_notify ON items;
DROP FUNCTION notify_change();
DROP TABLE outbox;
```

- The payload is the topic name only.
  The listener reads the outbox for the rows.
- PostgreSQL sends the notification at commit.
  A rollback sends nothing.
- Without an at-least-once requirement, drop the outbox, and send `json_build_object('op', v_op, 'id', v_id)` as the payload.
  The listener then reads the row by id.
- Delete outbox rows older than one hour with a periodic `DELETE`.
  Keep one hour, so that a client that reconnects after a short outage gets the deltas.

## PostgreSQL: the listener

```go
package realtime

import (
	"context"
	"log/slog"
	"math/rand/v2"
	"time"

	"github.com/jackc/pgx/v5"
)

// Listener holds one dedicated connection in LISTEN mode and wakes the
// relay. It reconnects with backoff when the connection fails.
type Listener struct {
	connString string
	wake       chan struct{}
	log        *slog.Logger
}

func NewListener(connString string, log *slog.Logger) *Listener {
	return &Listener{connString: connString, wake: make(chan struct{}, 1), log: log}
}

// Wake returns a channel that gets one value for a burst of notifications.
func (l *Listener) Wake() <-chan struct{} { return l.wake }

// Run blocks until ctx ends.
func (l *Listener) Run(ctx context.Context) {
	delay := time.Second
	for ctx.Err() == nil {
		err := l.listen(ctx)
		if ctx.Err() != nil {
			return
		}
		l.log.Warn("listener stopped, reconnect follows", "err", err, "delay", delay)
		select {
		case <-time.After(delay + rand.N(delay/2)):
		case <-ctx.Done():
			return
		}
		delay = min(delay*2, 30*time.Second)
	}
}

func (l *Listener) listen(ctx context.Context) error {
	conn, err := pgx.Connect(ctx, l.connString)
	if err != nil {
		return err
	}
	defer conn.Close(context.Background())
	if _, err := conn.Exec(ctx, "LISTEN app_changes"); err != nil {
		return err
	}
	// A change can have happened while the connection was down.
	l.signal()
	for {
		if _, err := conn.WaitForNotification(ctx); err != nil {
			return err
		}
		l.signal()
	}
}

// signal coalesces: a full channel means the relay has a wake-up pending.
func (l *Listener) signal() {
	select {
	case l.wake <- struct{}{}:
	default:
	}
}
```

- `pgx.Connect` gives a dedicated connection.
  Do not use `pool.Acquire` for the listener, because the pool health check interrupts `WaitForNotification`.
- Connect the listener to PostgreSQL itself, not to a transaction-mode pooler.
- The relay drains the outbox on each wake-up and on a timer of 30 seconds.
  The timer covers a lost notification.

## PostgreSQL: the relay

```go
// Relay reads the outbox after a wake-up and publishes the rows to the hub.
func (r *Relay) Run(ctx context.Context, wake <-chan struct{}) {
	tick := time.NewTicker(30 * time.Second)
	defer tick.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-wake:
		case <-tick.C:
		}
		if err := r.drain(ctx); err != nil {
			r.log.Warn("outbox drain failed", "err", err)
		}
	}
}

func (r *Relay) drain(ctx context.Context) error {
	rows, err := r.q.OutboxAfter(ctx, r.lastSeq)
	if err != nil {
		return err
	}
	for _, o := range rows {
		var row json.RawMessage
		if o.Op == "upsert" {
			row, err = r.q.ItemJSON(ctx, o.RowID)
			if err != nil {
				return err
			}
		}
		r.hub.Publish(o.Topic, o.Seq, Delta{Op: o.Op, ID: o.RowID, Row: row})
		r.lastSeq = o.Seq
	}
	return nil
}
```

- `OutboxAfter` is `SELECT seq, topic, op, row_id FROM outbox WHERE seq > $1 ORDER BY seq LIMIT 1000`.
- `ItemJSON` is `SELECT to_jsonb(i) FROM items i WHERE id = $1`.
  A row that is gone between the outbox read and the row read becomes a `delete`.
- At start, set `lastSeq` to `SELECT coalesce(max(seq), 0) FROM outbox`.
  The hub keeps the last hour of deltas for each topic, so that a `since` in that window replays.

## MongoDB: the watcher

```go
package realtime

import (
	"context"
	"log/slog"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type changeEvent struct {
	ID            bson.Raw `bson:"_id"`
	OperationType string   `bson:"operationType"`
	DocumentKey   struct {
		ID bson.ObjectID `bson:"_id"`
	} `bson:"documentKey"`
	FullDocument bson.Raw `bson:"fullDocument"`
}

// Watch follows the change stream of coll and publishes each event. It
// resumes from the stored token after a failure.
func (w *Watcher) Watch(ctx context.Context, coll *mongo.Collection, topic string) {
	for ctx.Err() == nil {
		opts := options.ChangeStream().SetFullDocument(options.UpdateLookup)
		if tok := w.tokens.Load(topic); tok != nil {
			opts.SetResumeAfter(tok)
		}
		cs, err := coll.Watch(ctx, mongo.Pipeline{}, opts)
		if err != nil {
			w.log.Warn("change stream open failed", "err", err)
			w.sleep(ctx)
			continue
		}
		w.consume(ctx, cs, topic)
		cs.Close(context.Background())
	}
}

func (w *Watcher) consume(ctx context.Context, cs *mongo.ChangeStream, topic string) {
	for cs.Next(ctx) {
		var ev changeEvent
		if err := cs.Decode(&ev); err != nil {
			w.log.Warn("change event decode failed", "err", err)
			continue
		}
		d := Delta{ID: ev.DocumentKey.ID.Hex()}
		switch ev.OperationType {
		case "insert", "update", "replace":
			d.Op = "upsert"
			d.Row = toJSON(ev.FullDocument)
		case "delete":
			d.Op = "delete"
		default:
			continue
		}
		seq := w.hub.Publish(topic, 0, d)
		w.tokens.Store(topic, cs.ResumeToken(), seq)
	}
	if err := cs.Err(); err != nil {
		w.log.Warn("change stream failed", "err", err)
	}
}
```

- Store the resume token after `Publish`, not before.
  A crash between the two repeats one event, and the client ignores a repeated delta.
- Keep the `_id` field in every pipeline stage.
  The resume token lives in it.
- `fullDocument: updateLookup` reads the document after the update.
  Enable `changeStreamPreAndPostImages` on the collection when the client needs the old values.
- A change stream needs a replica set.
  Start a one-member replica set with `mongod --replSet rs0` and `rs.initiate()` for a small install.
- MongoDB has no sequence.
  The hub assigns `seq` when `Publish` gets `0`.

## The in-process bus

In the `none` variant, the store calls `hub.Publish` after each change under its mutex.
There is no listener and no watcher.
Give the store a `Publisher` interface with one method, `Publish(topic, op, id string, row any)`, so that a test passes a fake.
The hub implements it and encodes the row.

## The hub

```go
package realtime

import (
	"context"
	"sync"
	"time"

	"github.com/coder/websocket"
)

type Hub struct {
	mu      sync.Mutex
	topics  map[string]*topic
	clients map[*Client]struct{}
}

type topic struct {
	seq     int64
	history []message // the last hour, oldest first
	subs    map[*Client]struct{}
}

type message struct {
	seq  int64
	at   time.Time
	data []byte // the encoded envelope
}

type Client struct {
	userID string
	conn   *websocket.Conn
	send   chan []byte
	topics map[string]struct{}
}

// Publish encodes one delta and sends it to every subscriber of the topic.
// With seq 0, the hub assigns the next sequence number. It returns the
// sequence number that it used.
func (h *Hub) Publish(name string, seq int64, d Delta) int64 {
	h.mu.Lock()
	defer h.mu.Unlock()
	t := h.topic(name)
	if seq == 0 {
		seq = t.seq + 1
	}
	t.seq = seq
	data := encode(Envelope{V: 1, Type: "delta", Topic: name, Seq: seq, Data: mustJSON(d)})
	t.history = append(t.history, message{seq: seq, at: time.Now(), data: data})
	t.trim()
	for c := range t.subs {
		select {
		case c.send <- data:
		default:
			// The client does not read. Close it, and let it reconnect.
			go c.conn.Close(websocket.StatusPolicyViolation, "slow consumer")
			delete(t.subs, c)
		}
	}
	return seq
}

// Subscribe adds the client and replays from since. It returns false when
// the history does not reach since, and the client must take a snapshot.
func (h *Hub) Subscribe(c *Client, name string, since int64) bool {
	h.mu.Lock()
	defer h.mu.Unlock()
	t := h.topic(name)
	t.subs[c] = struct{}{}
	c.topics[name] = struct{}{}
	if since == 0 || len(t.history) == 0 || t.history[0].seq > since+1 {
		return since == 0 && t.seq == 0 || since == t.seq
	}
	for _, m := range t.history {
		if m.seq > since {
			c.send <- m.data
		}
	}
	return true
}
```

- A send buffer of 256 messages for each client is enough for a burst.
- `Publish` holds the mutex while it writes to the channels.
  The writes are non-blocking, so the hold is short.
- The `topic` method creates the topic on first use, and `trim` drops history older than one hour.
- When `Subscribe` returns false, the handler reads a snapshot from the store, sends it with the current `seq`, and the deltas after it follow from the history.

## The WebSocket handler

```go
func (h *Hub) ServeWS(w http.ResponseWriter, r *http.Request) {
	user := userFrom(r.Context())
	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		OriginPatterns: h.origins, // host names, for example "app.cyberdyne.example"
	})
	if err != nil {
		return
	}
	conn.SetReadLimit(64 << 10)
	c := &Client{userID: user.ID, conn: conn, send: make(chan []byte, 256), topics: map[string]struct{}{}}
	h.add(c)
	defer h.remove(c)

	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()
	go h.writePump(ctx, c)
	h.readPump(ctx, c) // returns when the client closes or an error occurs
}

func (h *Hub) writePump(ctx context.Context, c *Client) {
	ping := time.NewTicker(30 * time.Second)
	defer ping.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case data := <-c.send:
			wctx, cancel := context.WithTimeout(ctx, 10*time.Second)
			err := c.conn.Write(wctx, websocket.MessageText, data)
			cancel()
			if err != nil {
				return
			}
		case <-ping.C:
			pctx, cancel := context.WithTimeout(ctx, 30*time.Second)
			err := c.conn.Ping(pctx)
			cancel()
			if err != nil {
				c.conn.Close(websocket.StatusPolicyViolation, "no pong")
				return
			}
		}
	}
}
```

- `websocket.Accept` checks the `Origin` header against `OriginPatterns`.
  Without a pattern, it accepts only the same host.
- `requireUser` runs before `ServeWS`, so an upgrade without a session gets 401.
- `readPump` decodes each `Envelope`, and calls `Subscribe` or `Unsubscribe`.
  It rejects a type that it does not know with an `error` envelope.
- Authorize each topic.
  A user subscribes to `me.settings` for the own user id only.

## The client module

```ts
// web/src/lib/ws.svelte.ts
import * as v from "valibot";
import { EnvelopeSchema, type Envelope } from "./protocol";

type Handler = (e: Envelope) => void;

class Socket {
  status = $state<"open" | "closed" | "connecting">("closed");
  #ws: WebSocket | undefined;
  #handlers = new Map<string, Set<Handler>>();
  #since = new Map<string, number>();
  #attempt = 0;
  #timer: number | undefined;

  constructor(private url: string) {
    document.addEventListener("visibilitychange", () => {
      if (document.visibilityState === "visible" && this.status === "closed") this.connect();
    });
  }

  connect(): void {
    if (this.#ws) return;
    this.status = "connecting";
    const ws = new WebSocket(this.url);
    this.#ws = ws;
    ws.onopen = () => {
      this.status = "open";
      this.#attempt = 0;
      for (const topic of this.#handlers.keys()) this.#send({ v: 1, type: "subscribe", topic, since: this.#since.get(topic) });
    };
    ws.onmessage = (ev) => {
      const parsed = v.safeParse(EnvelopeSchema, JSON.parse(ev.data));
      if (!parsed.success) return;
      const e = parsed.output;
      if ("seq" in e && e.topic) this.#since.set(e.topic, e.seq);
      for (const h of this.#handlers.get("topic" in e ? e.topic : "") ?? []) h(e);
    };
    ws.onclose = () => {
      this.#ws = undefined;
      this.status = "closed";
      this.#scheduleReconnect();
    };
  }

  subscribe(topic: string, handler: Handler): () => void {
    let set = this.#handlers.get(topic);
    if (!set) {
      set = new Set();
      this.#handlers.set(topic, set);
      if (this.status === "open") this.#send({ v: 1, type: "subscribe", topic, since: this.#since.get(topic) });
    }
    set.add(handler);
    return () => {
      set?.delete(handler);
      if (set?.size === 0) {
        this.#handlers.delete(topic);
        if (this.status === "open") this.#send({ v: 1, type: "unsubscribe", topic });
      }
    };
  }

  #send(e: Envelope): void {
    this.#ws?.send(JSON.stringify(e));
  }

  #scheduleReconnect(): void {
    if (document.visibilityState !== "visible") return;
    const base = Math.min(1000 * 2 ** this.#attempt, 30_000);
    const delay = base / 2 + Math.random() * (base / 2);
    this.#attempt += 1;
    clearTimeout(this.#timer);
    this.#timer = window.setTimeout(() => this.connect(), delay);
  }
}

export const socket = new Socket(`${location.protocol === "https:" ? "wss" : "ws"}://${location.host}/ws`);
```

- The browser sends the session cookie with the upgrade.
  No token goes in the URL.
- A view calls `subscribe` in `$effect` and returns the unsubscribe function.
- A `resync` envelope makes the view get a snapshot through the API, and reset `since` for the topic to the `seq` of the snapshot.
- A hidden tab does not reconnect.
  The `visibilitychange` listener reconnects when the tab shows again.
- A unit test replaces `WebSocket` with a fake class, see `testing.md`.
