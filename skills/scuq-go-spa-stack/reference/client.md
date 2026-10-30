# The Svelte client

## BLUF

This file gives the code for sections 3, 9, 10, and 11 of the skill.
It covers the Vite project, the configs, the route module, the API client with valibot, the settings store, the design tokens, and the `Modal.svelte` component.
The client is one Vite project in `web/`, and the Go server embeds its build.
Both client profiles share these parts.

## The project

```text
web/
  package.json
  vite.config.ts
  tsconfig.json
  biome.json
  playwright.config.ts
  index.html
  src/
    main.ts              mount(App), connect the socket
    App.svelte           the route outlet and the login gate
    app.css              the tokens and the base styles
    lib/
      api.ts             fetch wrapper, one function for each route
      protocol.ts        the valibot schemas of the envelope and the rows
      router.svelte.ts   the route module
      ws.svelte.ts       the socket, see realtime.md
      settings.svelte.ts the per-user settings store
      commands.svelte.ts the command registry, workstation only
    components/
      Modal.svelte
      Toast.svelte
    routes/
      Inventory.svelte
      Login.svelte
      NotFound.svelte
    tests/               Vitest files next to the modules, or here
  e2e/                   Playwright tests
```

- `npm create vite@latest web -- --template svelte-ts` gives the start.
- Add `biome`, `valibot`, `vitest`, `@testing-library/svelte`, `jsdom`, and `@playwright/test`.
- `svelte-check` names the TypeScript majors that it supports in its peer dependency.
  Pin `typescript` to the newest of them, or `npm install` fails.
- Put a `popout.html` with an empty body in `public/` for the dockview popout window.
- Set `"type": "module"` and an `engines.node` of the current LTS in `package.json`.

## The configs

`vite.config.ts`:

```ts
import { svelte } from "@sveltejs/vite-plugin-svelte";
import { defineConfig } from "vitest/config";

export default defineConfig({
  plugins: [svelte()],
  build: { outDir: "../internal/web/dist", emptyOutDir: true, sourcemap: true, target: "es2022" },
  server: {
    proxy: {
      "/api": "http://localhost:8080",
      "/auth": "http://localhost:8080",
      "/ws": { target: "ws://localhost:8080", ws: true },
    },
  },
  test: {
    environment: "jsdom",
    include: ["src/**/*.test.ts"],
    setupFiles: ["src/tests/setup.ts"],
  },
});
```

`tsconfig.json` sets `"strict": true`, `"noUncheckedIndexedAccess": true`, `"exactOptionalPropertyTypes": true`, `"moduleResolution": "bundler"`, and `"verbatimModuleSyntax": true`.

`biome.json`:

```json
{
  "$schema": "https://biomejs.dev/schemas/2.0.0/schema.json",
  "formatter": { "indentStyle": "space", "indentWidth": 2, "lineWidth": 100 },
  "linter": {
    "rules": {
      "recommended": true,
      "suspicious": { "noExplicitAny": "error" },
      "correctness": { "noUnusedImports": "error", "noUnusedVariables": "error" }
    }
  },
  "javascript": { "formatter": { "quoteStyle": "double" } }
}
```

`package.json` scripts: `dev`, `build`, `check` (`biome check . && svelte-check && tsc --noEmit`), `test` (`vitest run`), and `e2e` (`playwright test`).

- `defineConfig` comes from `vitest/config`, so that the `test` block has a type.
- The build writes to `../internal/web/dist`, because `embed` reads only below the Go package.
- The Vite proxy sends `/api`, `/auth`, and `/ws` to the Go server in development.
  The Go server then serves no client files, and `index.html` comes from Vite without the nonce.
  Test the CSP with the production build.

## Effects that read and write the same state

An `$effect` that reads a `$state` value and then writes it runs again, and Svelte stops it with `effect_update_depth_exceeded`.
Read the trigger values first, and wrap the work in `untrack` from `svelte`.
The dashboard history of the showcase is the example: the effect tracks the last change, and `untrack` wraps the `record` call that appends to the history.

## TypeScript in a component

The Svelte compiler accepts only the TypeScript features that it can strip.
A constructor parameter property such as `constructor(private name: string)` fails the build.
Write a class with such a feature in a `.ts` module, and import it into the component.

## The route module

```ts
// web/src/lib/router.svelte.ts
import type { Component } from "svelte";

export type Route = { pattern: RegExp; view: Component<{ params: Record<string, string> }>; keys: string[] };

class Router {
  path = $state(location.pathname);

  constructor() {
    window.addEventListener("popstate", () => {
      this.path = location.pathname;
    });
  }

  go(path: string): void {
    if (path === this.path) return;
    history.pushState(null, "", path);
    this.path = path;
  }

  match(routes: Route[]): { view: Route["view"]; params: Record<string, string> } | undefined {
    for (const r of routes) {
      const m = r.pattern.exec(this.path);
      if (!m) continue;
      const params: Record<string, string> = {};
      r.keys.forEach((k, i) => {
        params[k] = decodeURIComponent(m[i + 1] ?? "");
      });
      return { view: r.view, params };
    }
    return undefined;
  }
}

export const router = new Router();

// route turns "/items/:id" into a Route.
export function route(path: string, view: Route["view"]): Route {
  const keys: string[] = [];
  const src = path.replace(/:([a-zA-Z]+)/g, (_, k: string) => {
    keys.push(k);
    return "([^/]+)";
  });
  return { pattern: new RegExp(`^${src}/?$`), view, keys };
}
```

- `App.svelte` holds the route table and renders `match(routes)?.view` or `NotFound`.
- A link is an `<a href>`.
  One click handler on `App.svelte` calls `router.go` for a same-origin link without a modifier key.
- The Go fallback serves `index.html` for every path, so a deep link works.

## The API client

```ts
// web/src/lib/api.ts
import * as v from "valibot";

export class ApiError extends Error {
  constructor(
    public status: number,
    message: string,
    public requestId?: string,
  ) {
    super(message);
  }
}

async function call<T>(method: string, path: string, schema: v.GenericSchema<T>, body?: unknown): Promise<T> {
  const res = await fetch(path, {
    method,
    headers: body === undefined ? {} : { "Content-Type": "application/json" },
    body: body === undefined ? undefined : JSON.stringify(body),
    credentials: "same-origin",
  });
  if (res.status === 401) {
    window.dispatchEvent(new Event("app:unauthorized"));
  }
  if (!res.ok) {
    const err = await res.json().catch(() => ({}));
    throw new ApiError(res.status, err.error ?? res.statusText, res.headers.get("X-Request-Id") ?? undefined);
  }
  if (res.status === 204) return undefined as T;
  return v.parse(schema, await res.json());
}

export const ItemSchema = v.object({
  id: v.pipe(v.string(), v.uuid()),
  name: v.string(),
  status: v.picklist(["active", "retired"]),
  updatedAt: v.pipe(v.string(), v.isoTimestamp()),
});
export type Item = v.InferOutput<typeof ItemSchema>;

export const api = {
  listItems: () => call("GET", "/api/items", v.array(ItemSchema)),
  getItem: (id: string) => call("GET", `/api/items/${encodeURIComponent(id)}`, ItemSchema),
  patchItems: (ids: string[], patch: Partial<Item>) =>
    call("PATCH", "/api/items", v.array(ItemSchema), { ids, patch }),
};
```

- Every response passes a schema.
  A server change that breaks the shape fails at the boundary, not deep in a view.
- `app:unauthorized` makes `App.svelte` show the login view and keep the route.
- No request sends a token.
  The session cookie goes with `credentials: "same-origin"`.

## The settings store

The API carries the settings as a document `{version, settings}`.
The version stays outside the object, and the store keeps it in a field of its own.

```ts
// web/src/lib/settings.svelte.ts
import * as v from "valibot";
import { ApiError, api } from "./api";
import { socket } from "./ws.svelte";

export const SettingsSchema = v.object({
  theme: v.optional(v.picklist(["system", "light", "dark"]), "system"),
  view: v.optional(v.picklist(["dashboard", "workstation"]), "dashboard"),
  inventory: v.optional(
    v.object({
      hiddenColumns: v.optional(v.array(v.string()), []),
      statusFilter: v.optional(v.string(), ""),
    }),
    {},
  ),
  dock: v.optional(v.object({ layoutVersion: v.number(), layout: v.unknown() })),
});
export type Settings = v.InferOutput<typeof SettingsSchema>;

const DocSchema = v.object({ version: v.number(), settings: v.unknown() });

class SettingsStore {
  value = $state<Settings>(v.parse(SettingsSchema, {}));
  version = $state(0);
  saving = $state(false);
  error = $state<string | undefined>();
  loaded = $state(false);
  #timer: ReturnType<typeof setTimeout> | undefined;

  async load(): Promise<void> {
    const doc = await api.getSettings();
    this.#apply(doc.version, doc.settings);
    this.loaded = true;
    socket.subscribe("me.settings", (e) => {
      if (e.type !== "delta" || e.data.op !== "upsert") return;
      const parsed = v.safeParse(DocSchema, e.data.row);
      if (parsed.success && parsed.output.version > this.version) this.#apply(parsed.output.version, parsed.output.settings);
    });
    window.addEventListener("pagehide", () => void this.flush(true));
  }

  // #apply parses the object with defaults. An unknown shape from an old
  // version falls back to the defaults instead of a broken view.
  #apply(version: number, raw: unknown): void {
    const parsed = v.safeParse(SettingsSchema, raw);
    this.value = parsed.success ? parsed.output : v.parse(SettingsSchema, {});
    this.version = version;
  }

  // update changes one part, and schedules the write one second later.
  update(fn: (s: Settings) => void): void {
    fn(this.value);
    clearTimeout(this.#timer);
    this.#timer = setTimeout(() => void this.flush(), 1000);
  }

  async flush(keepalive = false): Promise<void> {
    clearTimeout(this.#timer);
    if (!this.loaded) return;
    this.saving = true;
    try {
      const doc = await api.putSettings(this.version, $state.snapshot(this.value), keepalive);
      this.version = doc.version;
      this.error = undefined;
    } catch (e) {
      if (e instanceof ApiError && e.status === 409) {
        const doc = await api.getSettings();
        this.#apply(doc.version, doc.settings);
        this.error = "Another device changed the settings. The newer settings are shown.";
      } else {
        this.error = "The settings were not saved.";
      }
    } finally {
      this.saving = false;
    }
  }
}

export const settings = new SettingsStore();
```

- `api.putSettings(version, settings, keepalive)` sends `{version, settings}`.
  The server stores `version + 1`, and the response carries it.
- Send `$state.snapshot(this.value)`, not the proxy.
- A view reads `settings.value.inventory.statusFilter` and calls `settings.update` on a change.
- The `me.settings` delta from another device wins only when its version is newer.
- `pagehide` fires on a tab close and on a navigation away.
  `fetch` with `keepalive: true` lets the request finish after the page is gone.
- The open view, the dock layout, the filters, and the theme all live here.
  Nothing of that goes to `localStorage`.

## Design tokens

```css
/* web/src/app.css */
:root {
  --color-bg: #ffffff;
  --color-surface: #f6f7f9;
  --color-text: #1b1f24;
  --color-muted: #5b6470;
  --color-border: #d5dae1;
  --color-accent: #2457d6;
  --color-focus: #2457d6;
  --space-1: 4px;
  --space-2: 8px;
  --space-3: 12px;
  --space-4: 16px;
  --font-size-1: 12px;
  --font-size-2: 13px;
  --font-size-3: 15px;
  --row-height: 28px;
  --radius: 4px;
  color-scheme: light dark;
}

@media (prefers-color-scheme: dark) {
  :root {
    --color-bg: #15181d;
    --color-surface: #1d2127;
    --color-text: #e6e9ee;
    --color-muted: #9aa4b1;
    --color-border: #313843;
    --color-accent: #6b9bff;
    --color-focus: #6b9bff;
  }
}

html, body { height: 100%; margin: 0; background: var(--color-bg); color: var(--color-text); }
body { font: var(--font-size-2)/1.4 system-ui, sans-serif; }
table, .grid { font-variant-numeric: tabular-nums; }
:focus-visible { outline: 2px solid var(--color-focus); outline-offset: 1px; }
```

- Status colors come from the registry of the `scuq-hmi-semantics` skill when the application shows system state.
  Do not define them here.
- A `data-theme` attribute on `html` overrides `prefers-color-scheme` when the user chose a theme in the settings.
- Every library reads these tokens.
  `workstation.md` maps them to the dockview and RevoGrid variables.

## Modal.svelte

```svelte
<script lang="ts">
  import type { Snippet } from "svelte";

  let {
    open = $bindable(false),
    title,
    size = "form",
    closedby = "closerequest",
    dirty = () => false,
    children,
    footer,
  }: {
    open?: boolean;
    title: string;
    size?: "confirm" | "form" | "wide";
    closedby?: "any" | "closerequest" | "none";
    dirty?: () => boolean;
    children: Snippet;
    footer?: Snippet;
  } = $props();

  let dialog: HTMLDialogElement;
  let opener: Element | null = null;
  const id = `dialog-${Math.random().toString(36).slice(2)}`;

  $effect(() => {
    if (open && !dialog.open) {
      opener = document.activeElement;
      dialog.showModal();
    } else if (!open && dialog.open) {
      dialog.close();
    }
  });

  function oncancel(e: Event) {
    if (dirty() && !confirm("Discard the changes?")) e.preventDefault();
  }

  function onclose() {
    open = false;
    if (opener instanceof HTMLElement) opener.focus();
  }
</script>

<dialog bind:this={dialog} class={size} aria-labelledby={id} {closedby} {oncancel} {onclose}>
  <header><h2 id={id} tabindex="-1">{title}</h2></header>
  <div class="body">{@render children()}</div>
  {#if footer}<footer>{@render footer()}</footer>{/if}
</dialog>

<style>
  dialog {
    width: min(90vw, 40rem);
    max-height: 90dvh;
    padding: 0;
    border: 1px solid var(--color-border);
    border-radius: var(--radius);
    background: var(--color-surface);
    color: var(--color-text);
    grid-template-rows: auto 1fr auto;
  }
  dialog[open] { display: grid; }
  dialog.confirm { width: min(90vw, 28rem); }
  dialog.wide { width: min(96vw, 72rem); }
  dialog::backdrop { background: rgb(0 0 0 / 0.4); }
  dialog:modal { overscroll-behavior: contain; }
  header, footer { padding: var(--space-3) var(--space-4); }
  header { border-bottom: 1px solid var(--color-border); }
  footer { border-top: 1px solid var(--color-border); display: flex; justify-content: flex-end; gap: var(--space-2); }
  .body { overflow: auto; padding: var(--space-4); }
  h2 { margin: 0; font-size: var(--font-size-3); }
</style>
```

- The `cancel` event fires on Escape and on a close request.
  The `dirty` guard lives there, so an old browser without `closedby` gets the same behavior.
- The grid layout keeps the header and the footer fixed, and the body scrolls.
- The title gets the first focus through `tabindex="-1"` on `h2` and `autofocus` on it when the body is long.
  Put `autofocus` on the first field for a short form.
- The caller renders the primary action last in `footer`, with one verb: Save, Delete, Apply.
- Replace `confirm()` with a nested confirmation dialog when the application has a design for it.
