# Tests

## BLUF

This file gives the test patterns for section 13 of the skill.
It covers handler tests with `httptest`, timing tests with `synctest`, a WebSocket test, and store tests against a real database.
It also covers Vitest with a fake socket, and Playwright for the views, the grid, and the dock.
A test needs no network and no database, except the store tests, which skip without the environment variable.

## Handlers

```go
func TestPutSettingsVersionMismatch(t *testing.T) {
	st := fakeSettings{version: 3}
	h := newTestHandler(t, &st)
	body := `{"version":2,"theme":"dark"}`
	req := httptest.NewRequest(http.MethodPut, "/api/me/settings", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Sec-Fetch-Site", "same-origin")
	req.AddCookie(loginAs(t, h, "alice"))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusConflict {
		t.Fatalf("status = %d, want 409", rec.Code)
	}
}
```

- `newTestHandler` builds the full middleware chain with a fake store and `memstore`.
  A test then covers the headers, the CSRF check, and the session together.
- `loginAs` posts to the login route and returns the session cookie.
- Set `Sec-Fetch-Site: same-origin` on a state-changing request, or `CrossOriginProtection` answers 403.
- Add one test that sends `Sec-Fetch-Site: cross-site` and expects 403.
- Add one test that reads every security header of a response.

## Timing with synctest

```go
func TestListenerBackoff(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		var tries atomic.Int32
		l := &Listener{
			wake: make(chan struct{}, 1),
			log:  slog.New(slog.DiscardHandler),
			dial: func(ctx context.Context) error {
				tries.Add(1)
				return errors.New("refused")
			},
		}
		ctx, cancel := context.WithCancel(t.Context())
		go l.Run(ctx)

		synctest.Wait()
		if got := tries.Load(); got != 1 {
			t.Fatalf("tries after start = %d, want 1", got)
		}
		time.Sleep(3 * time.Second) // the fake clock passes the first delay of one to 1.5 seconds
		synctest.Wait()
		if got := tries.Load(); got != 2 {
			t.Fatalf("tries after 3 s = %d, want 2", got)
		}
		cancel()
	})
}
```

- `synctest.Test` runs the function in a bubble with a fake clock.
  `time.Sleep` returns at once when every goroutine in the bubble blocks.
- Give the listener a `dial` function field, so that the test needs no database.
- Test the hub ping in the same way: a client without a pong closes after 60 seconds of fake time.

## WebSocket

```go
func TestSubscribeReplaysHistory(t *testing.T) {
	hub := NewHub(nil)
	srv := httptest.NewServer(withTestUser(hub.ServeWS))
	defer srv.Close()

	hub.Publish("inventory", 0, Delta{Op: "upsert", ID: "a"})
	hub.Publish("inventory", 0, Delta{Op: "upsert", ID: "b"})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	conn, _, err := websocket.Dial(ctx, "ws"+strings.TrimPrefix(srv.URL, "http")+"/ws", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.CloseNow()

	since := int64(1)
	if err := wsjson.Write(ctx, conn, Envelope{V: 1, Type: "subscribe", Topic: "inventory", Since: &since}); err != nil {
		t.Fatal(err)
	}
	var got Envelope
	if err := wsjson.Read(ctx, conn, &got); err != nil {
		t.Fatal(err)
	}
	if got.Type != "delta" || got.Seq != 2 {
		t.Fatalf("got %+v, want delta seq 2", got)
	}
}
```

- `withTestUser` puts a fixed user in the context, so the test skips the login.
- `wsjson` is `github.com/coder/websocket/wsjson`.
- Add tests for a `since` that is too old, which expects `resync`, and for a slow client, which expects a close.

## Store tests against a database

```go
func openTestPool(t *testing.T) *pgxpool.Pool {
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}
	ctx := t.Context()
	if err := Migrate(ctx, url); err != nil {
		t.Fatal(err)
	}
	pool, err := Open(ctx, url)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	return pool
}
```

- Each test runs in a transaction that it rolls back, or it uses a schema with a random name.
- CI starts PostgreSQL and MongoDB as services, and sets the variables.
  Local runs without the variables skip the tests.
- `testcontainers-go` is an option when Docker is on the machine.
  It is pure Go, and it adds no cgo.
- Test the trigger: insert a row, and read one outbox row with the topic and the id.
- Test the listener against the database: `LISTEN`, insert a row, and expect one wake-up within two seconds.

## A WebSocket check client

Give the module a small command, for example `cmd/wscheck`.
It logs in with the HTTP client, dials `/ws` with the cookie jar, subscribes to one topic, and prints the first frames.
It checks the live path after a deploy, and it is the first tool when a client sees no updates.
The showcase of the skill has one.

## Vitest

```ts
// web/src/lib/ws.test.ts
import { beforeEach, expect, test, vi } from "vitest";

class FakeSocket {
  static instances: FakeSocket[] = [];
  sent: string[] = [];
  onopen: (() => void) | null = null;
  onmessage: ((ev: { data: string }) => void) | null = null;
  onclose: (() => void) | null = null;
  constructor(public url: string) {
    FakeSocket.instances.push(this);
  }
  send(data: string) {
    this.sent.push(data);
  }
  close() {
    this.onclose?.();
  }
}

beforeEach(() => {
  FakeSocket.instances = [];
  vi.stubGlobal("WebSocket", FakeSocket);
  vi.useFakeTimers();
});

test("sends the subscriptions again after a reconnect", async () => {
  const { Socket } = await import("./ws.svelte");
  const s = new Socket("ws://test/ws");
  s.subscribe("inventory", () => {});
  s.connect();
  const first = FakeSocket.instances[0];
  first?.onopen?.();
  first?.onmessage?.({ data: JSON.stringify({ v: 1, type: "delta", topic: "inventory", seq: 7, data: { op: "delete", id: "a" } }) });
  first?.close();
  vi.advanceTimersByTime(2000);
  const second = FakeSocket.instances[1];
  second?.onopen?.();
  expect(second?.sent).toContain(JSON.stringify({ v: 1, type: "subscribe", topic: "inventory", since: 7 }));
});
```

- Export the `Socket` class next to the `socket` instance for the tests.
- Test the schemas with one valid and one invalid payload each.
- Test the settings store: `update` writes once after one second, and a 409 reloads.
- Test a component with `@testing-library/svelte`: render `Modal.svelte`, and check `aria-labelledby` and the `cancel` guard.
  `jsdom` has no `showModal`, so stub it on `HTMLDialogElement.prototype`.

## Playwright

```ts
// web/e2e/inventory.spec.ts
import { expect, test } from "@playwright/test";

test("a change on the server shows in the grid", async ({ page, request }) => {
  await page.goto("/");
  await page.getByLabel("Username").fill("alice");
  await page.getByLabel("Password").fill("correct horse battery staple");
  await page.getByRole("button", { name: "Log in" }).click();
  await expect(page.getByText("Inventory")).toBeVisible();

  const res = await request.patch("/api/items", {
    data: { ids: ["0193c2d0-0000-7000-8000-000000000001"], patch: { status: "retired" } },
    headers: { "Sec-Fetch-Site": "same-origin" },
  });
  expect(res.ok()).toBeTruthy();

  await expect(page.locator("revo-grid").getByText("retired").first()).toBeVisible();
});

test("the command palette focuses the inspector", async ({ page }) => {
  await page.goto("/");
  await page.keyboard.press("Control+k");
  await page.getByPlaceholder("Type a command").fill("focus inspector");
  await page.keyboard.press("Enter");
  await expect(page.locator(".panel-host:focus-within")).toContainText("Inspector");
});
```

- `playwright.config.ts` starts the Go binary with `webServer`, against a test database and a seed.
- A saved setting such as the open view survives between tests.
  The login helper sets the view it needs, or a test depends on the order.
- A virtualized grid renders only the visible rows.
  Change a row that the test can see, for example the first row of the API in the sort order of the grid.
- Start the binary on its own line in a shell script, not in a `cd ... && ./app &` chain.
  The chain backgrounds a subshell, and a `kill` of that subshell leaves the server as an orphan on the port.
- Run in Chromium and Firefox.
  `closedby` and `<dialog>` differ between them.
- Test the modal: open, Escape with a dirty form shows the guard, and the focus returns to the opener.
- Test the layout: move a panel, reload, and expect the same layout.
  Then run `Reset layout`, and expect the default.
- Test one popout when the browser permits it, or mark the test as `fixme` with the reason.
