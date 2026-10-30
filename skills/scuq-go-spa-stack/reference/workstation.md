# The workstation profile

## BLUF

This file gives the code for section 12 of the skill.
It covers the dockview adapter for Svelte 5, the panel registry, the layout save and restore, and the RevoGrid wrapper.
It also covers the command palette and the bulk-edit panel.
The profile makes the application behave like a desktop program on a large monitor.
Check the API names against the installed versions before you copy a snippet.

## Versions

The snippets were built and run with dockview 8.8, RevoGrid 4.30, Bits UI 2.20, Svelte 5.57, and Vite 8.3.

## Layout

```text
+------------------------------------------------------------------+
| Command bar: palette trigger, search, user menu                   |
+---------------------------------------+--------------------------+
| Inventory grid (dock group, 30 cols)  | Inspector (dock group)   |
|                                       |  Tabs: Asset | Vulns     |
|                                       +--------------------------+
|                                       | Bulk edit (dock group)   |
+---------------------------------------+--------------------------+
| Status line: connection, row count, selection, last update       |
+------------------------------------------------------------------+
```

- The command bar and the status line are fixed.
  The dock fills the rest with `height: 100dvh` minus the two bars.
- The default layout has three groups.
  The user moves, floats, or pops out a panel, and the layout saves.
- The status line shows the socket status from `socket.status`, so a stale grid is visible.

## The panel registry

```ts
// web/src/lib/panels.ts
import type { Component } from "svelte";
import BulkEdit from "../panels/BulkEdit.svelte";
import Inspector from "../panels/Inspector.svelte";
import InventoryGrid from "../panels/InventoryGrid.svelte";

export type PanelParams = Record<string, unknown>;

export const panels: Record<string, { title: string; view: Component<{ params: PanelParams }> }> = {
  inventory: { title: "Inventory", view: InventoryGrid },
  inspector: { title: "Inspector", view: Inspector },
  bulkEdit: { title: "Bulk edit", view: BulkEdit },
};
```

- The registry key is the dockview component name.
  The layout JSON stores it, so do not rename a key without a layout migration.

## The dockview adapter

Install `dockview`.
It re-exports `dockview-core`, and it is the package that ships the CSS.
`dockview-core` alone has no stylesheet in version 8.

```ts
// web/src/lib/dock-adapter.ts
import type { GroupPanelPartInitParameters, IContentRenderer } from "dockview";
import { mount, unmount } from "svelte";
import { type PanelParams, panels } from "./panels";

// SveltePanel renders one registry component into a dockview panel. dockview
// gives the element, and Svelte mounts the component into it.
export class SveltePanel implements IContentRenderer {
  readonly element = document.createElement("div");
  #name: string;
  #instance: Record<string, unknown> | undefined;

  constructor(name: string) {
    this.#name = name;
    this.element.className = "panel-host";
    this.element.tabIndex = -1;
  }

  init(params: GroupPanelPartInitParameters): void {
    const def = panels[this.#name];
    if (!def) {
      this.element.textContent = `Unknown panel ${this.#name}`;
      return;
    }
    this.#instance = mount(def.view, {
      target: this.element,
      props: { params: (params.params ?? {}) as PanelParams },
    });
  }

  dispose(): void {
    if (this.#instance) unmount(this.#instance);
  }
}
```

```svelte
<!-- web/src/components/Dock.svelte -->
<script lang="ts">
  import { createDockview, type DockviewApi, type SerializedDockview, themeDark, themeLight } from "dockview";
  import "dockview/dist/styles/dockview.css";
  import { untrack } from "svelte";
  import { SveltePanel } from "../lib/dock-adapter";
  import { panels } from "../lib/panels";

  let {
    layout,
    dark,
    onLayoutChange,
    api = $bindable(),
  }: {
    layout: SerializedDockview | undefined;
    dark: boolean;
    onLayoutChange: (layout: SerializedDockview) => void;
    api?: DockviewApi;
  } = $props();

  let host: HTMLDivElement;

  function addDefaultPanels(dock: DockviewApi): void {
    dock.addPanel({ id: "inventory", component: "inventory", title: panels.inventory?.title });
    dock.addPanel({
      id: "inspector",
      component: "inspector",
      title: panels.inspector?.title,
      position: { referencePanel: "inventory", direction: "right" },
      initialWidth: 420,
    });
    dock.addPanel({
      id: "bulkEdit",
      component: "bulkEdit",
      title: panels.bulkEdit?.title,
      position: { referencePanel: "inspector", direction: "below" },
      initialHeight: 260,
    });
  }

  // Create the dock once. The layout is read untracked, because every layout
  // change writes it back to the settings, and a tracked read would dispose
  // and re-create the dock, with every panel, on each change.
  $effect(() => {
    const initial = untrack(() => layout);
    const dock = createDockview(host, {
      theme: untrack(() => dark) ? themeDark : themeLight,
      createComponent: (options) => new SveltePanel(options.name),
    });
    api = dock;
    let restored = false;
    if (initial) {
      try {
        dock.fromJSON(initial);
        restored = dock.panels.length > 0;
      } catch {
        restored = false;
      }
    }
    if (!restored) {
      dock.clear();
      addDefaultPanels(dock);
    }
    const sub = dock.onDidLayoutChange(() => onLayoutChange(dock.toJSON()));
    return () => {
      sub.dispose();
      dock.dispose();
    };
  });

  // A theme change updates the options, it does not re-create the dock.
  $effect(() => {
    api?.updateOptions({ theme: dark ? themeDark : themeLight });
  });
</script>

<div bind:this={host} class="dock"></div>

<style>
  .dock { height: 100%; }
</style>
```

- `mount` and `unmount` are the Svelte 5 API for a component outside the tree.
  Each panel is its own root, and it reads the shared stores from `lib/`.
- The class lives in a `.ts` module.
  The Svelte compiler rejects a constructor parameter property inside a component.
- `fromJSON` throws on a layout from an old version.
  The adapter then builds the default layout.
- Read the `layout` prop with `untrack`.
  The dock writes each layout change to the settings.
  A tracked read then re-creates the dock and every panel on each change, and that drops the focus and the panel state.
- Change the theme with `api.updateOptions`, not with a re-create.
- `onDidLayoutChange` fires often while the user drags.
  The caller debounces the write, see the settings store.
- A float comes from `dock.addFloatingGroup(panel, { position: { left: 80, top: 80 }, width: 640, height: 420 })`.
  A popout comes from `dock.addPopoutGroup(panel)`, and it loads `/popout.html`.
  Bind both to commands.
- The popout window copies the stylesheets of the page.
  Put a `popout.html` with an empty body in `web/public/`, or the SPA fallback loads the whole application there.
- Give the panel element `tabindex="-1"`, so that a focus command can focus it.

## Layout save and restore

```ts
// in App.svelte or a layout store
const LAYOUT_VERSION = 1;

function onLayoutChange(layout: SerializedDockview): void {
  settings.update((s) => {
    s.dock = { layoutVersion: LAYOUT_VERSION, layout };
  });
}

function storedLayout(): SerializedDockview | undefined {
  const d = settings.value.dock;
  if (!d || d.layoutVersion !== LAYOUT_VERSION) return undefined;
  return d.layout as SerializedDockview;
}
```

- Raise `LAYOUT_VERSION` when a panel key changes or a panel goes away.
  An old layout is then ignored, and the user gets the default.
- The reset layout command sets `s.dock = undefined` and recreates the dock.

## Dock theme

```css
/* the dock element carries the theme class, so the selector needs the parent */
.dock .dockview-theme-light, .dock .dockview-theme-dark {
  --dv-group-view-background-color: var(--color-bg);
  --dv-tabs-and-actions-container-background-color: var(--color-surface);
  --dv-tabs-and-actions-container-height: 28px;
  --dv-tabs-and-actions-container-font-size: var(--font-size-1);
  --dv-activegroup-visiblepanel-tab-background-color: var(--color-bg);
  --dv-activegroup-visiblepanel-tab-color: var(--color-text);
  --dv-activegroup-hiddenpanel-tab-background-color: var(--color-surface);
  --dv-activegroup-hiddenpanel-tab-color: var(--color-muted);
  --dv-inactivegroup-visiblepanel-tab-background-color: var(--color-bg);
  --dv-inactivegroup-visiblepanel-tab-color: var(--color-muted);
  --dv-inactivegroup-hiddenpanel-tab-background-color: var(--color-surface);
  --dv-inactivegroup-hiddenpanel-tab-color: var(--color-muted);
  --dv-tab-divider-color: var(--color-border);
  --dv-separator-border: var(--color-border);
  --dv-paneview-active-outline-color: var(--color-focus);
  --dv-active-sash-color: var(--color-accent);
  --dv-edge-dock-indicator-color: var(--color-accent);
  --dv-floating-border: 1px solid var(--color-border);
}
.panel-host { height: 100%; overflow: auto; container-type: size; }
```

- Pass a theme object from `dockview` in the `theme` option.
  The object sets the class on the dock element, and the page CSS overrides its variables with a more specific selector.
- The variable names come from `dockview/dist/styles/dockview.css`.
  Read the `.dockview-theme-light` block of the installed version for the full list.
- No panel, tab, or group has a color of its own.
  Every value is a token.

## The grid

```svelte
<!-- web/src/panels/InventoryGrid.svelte -->
<script lang="ts">
  import type { ColumnRegular, Components } from "@revolist/revogrid";
  import { defineCustomElements } from "@revolist/revogrid/loader";
  import { inventory } from "../lib/inventory.svelte";
  import { selection } from "../lib/selection.svelte";

  defineCustomElements();

  let grid: HTMLElement & Components.RevoGrid;

  const columns: ColumnRegular[] = [
    { prop: "id", name: "Id", size: 120, pin: "colPinStart", readonly: true },
    { prop: "name", name: "Name", size: 220 },
    { prop: "status", name: "Status", size: 100 },
    // 27 more columns: owner, site, rack, model, serial, os, ip, ...
  ];

  $effect(() => {
    grid.columns = columns;
  });

  $effect(() => {
    grid.source = inventory.rows;
  });

  $effect(() => {
    grid.theme = dark ? "darkCompact" : "compact";
  });

  type EditDetail =
    | { prop: string; val?: unknown; rowIndex: number }
    | { data: Record<number, Record<string, unknown>> };

  function onafteredit(e: CustomEvent<EditDetail>) {
    const d = e.detail;
    if ("data" in d) return; // a range paste or fill, see the showcase for the loop
    const row = inventory.rows[d.rowIndex];
    if (row) inventory.patch([row.id], { [d.prop]: d.val });
  }

  async function onafterfocus(e: CustomEvent<{ rowIndex: number }>) {
    const row = inventory.rows[e.detail.rowIndex];
    if (row) selection.focusedId = row.id;
    const range = await grid.getSelectedRange();
    if (range) selection.set(inventory.rows.slice(range.y, range.y1 + 1).map((r) => r.id));
    else if (row) selection.set([row.id]);
  }
</script>

<revo-grid
  bind:this={grid}
  row-size={28}
  resize={true}
  range={true}
  can-focus={true}
  row-headers={true}
  {onafteredit}
  {onafterfocus}
  class="grid"
></revo-grid>

<style>
  .grid { height: 100%; width: 100%; }
</style>
```

```ts
// web/src/lib/inventory.svelte.ts
import { type Item, api } from "./api";
import { socket } from "./ws.svelte";

class Inventory {
  rows = $state<Item[]>([]);
  #byId = new Map<string, Item>();
  #pending: Array<{ op: "upsert" | "delete"; id: string; row?: Item }> = [];
  #frame: number | undefined;

  async load(): Promise<void> {
    const items = await api.listItems();
    this.#byId = new Map(items.map((i) => [i.id, i]));
    this.rows = items;
    socket.subscribe("inventory", (e) => {
      if (e.type === "delta") this.#queue(e.data as (typeof this.#pending)[number]);
      if (e.type === "resync") void this.load();
    });
  }

  // #queue collects the deltas of one frame, and rebuilds the array once.
  #queue(d: (typeof this.#pending)[number]): void {
    this.#pending.push(d);
    this.#frame ??= requestAnimationFrame(() => {
      for (const p of this.#pending) {
        if (p.op === "delete") this.#byId.delete(p.id);
        else if (p.row) this.#byId.set(p.id, p.row);
      }
      this.#pending = [];
      this.#frame = undefined;
      this.rows = [...this.#byId.values()];
    });
  }

  async patch(ids: string[], patch: Partial<Item>): Promise<void> {
    await api.patchItems(ids, patch);
    // The rows come back over the socket. No local write here.
  }
}

export const inventory = new Inventory();
```

- `defineCustomElements` registers `<revo-grid>` once.
  Call it in `main.ts` when more than one panel uses the grid.
- The package exports `Components.RevoGrid` for the element type, not `HTMLRevoGridElement`.
  Declare `"revo-grid"` in `svelteHTML.IntrinsicElements` in `src/app.d.ts`, or `svelte-check` flags the tag.
- Set `theme` as a property.
  The built-in themes are `default`, `material`, `compact`, `darkMaterial`, and `darkCompact`.
- `afteredit` carries one cell, or a `data` map of row index to changed fields after a range paste.
  Handle both shapes.
- `afterfocus` fires on each cell focus.
  `getSelectedRange()` gives the rows of a range selection.
- A delta changes the map, and one new array goes to the grid once per frame.
  A burst of 1000 deltas is one render.
- The grid shows the server state.
  An edit goes to the server, and the result comes back over the socket.
  Show a pending marker on the cell until it does.
- Pin the id column, and let the user hide and order the other columns.
  Store the column list in `settings.value.inventory.columns`.
- Check `e.detail.val` against the schema of the column before `patch`.
- Set `readonly` on a column that the user cannot edit.
- Read the event names and the `ColumnRegular` fields from the installed RevoGrid version.

Grid theme:

```css
revo-grid {
  --revo-grid-bg: var(--color-bg);
  font-size: var(--font-size-1);
  font-variant-numeric: tabular-nums;
}
revo-grid .rgHeaderCell { background: var(--color-surface); color: var(--color-muted); border-color: var(--color-border); }
revo-grid .rgCell { border-color: var(--color-border); }
revo-grid .rgRow.focused-rgRow { background: color-mix(in srgb, var(--color-accent) 12%, transparent); }
```

- RevoGrid renders in the light DOM, so page CSS reaches the cells.
  The class names `rgHeaderCell`, `rgCell`, `rgRow`, and `focused-rgRow` are in version 4.30.
- RevoGrid and dockview set style attributes, so the CSP needs `style-src 'self' 'unsafe-inline'`.
  Scripts keep the nonce.

## The command palette

```svelte
<!-- web/src/components/CommandPalette.svelte -->
<script lang="ts">
  import { Command, Dialog } from "bits-ui";
  import { commands } from "../lib/commands.svelte";

  let open = $state(false);

  function onkeydown(e: KeyboardEvent) {
    if ((e.ctrlKey || e.metaKey) && e.key === "k") {
      e.preventDefault();
      open = !open;
    }
  }

  // The dialog restores the focus to its opener when it closes, which undoes
  // a command that moves the focus. run stores the command, and the close
  // auto-focus handler runs it in place of the restore.
  let pending: string | undefined;

  function run(id: string) {
    pending = id;
    open = false;
  }

  function onCloseAutoFocus(e: Event) {
    if (!pending) return;
    e.preventDefault();
    const id = pending;
    pending = undefined;
    commands.run(id);
  }
</script>

<svelte:window {onkeydown} />

<Dialog.Root bind:open>
  <Dialog.Portal>
    <Dialog.Overlay class="palette-overlay" />
    <Dialog.Content class="palette" aria-label="Command palette" {onCloseAutoFocus}>
      <Command.Root>
        <Command.Input placeholder="Type a command" />
        <Command.List>
          <Command.Viewport>
            <Command.Empty>No command matches.</Command.Empty>
            {#each commands.groups() as group (group.name)}
              <Command.Group value={group.name}>
                <Command.GroupHeading>{group.name}</Command.GroupHeading>
                <Command.GroupItems>
                  {#each group.items as c (c.id)}
                    <Command.Item value={c.id} keywords={[c.title, ...(c.keywords ?? [])]} onSelect={() => run(c.id)}>
                      <span>{c.title}</span>
                      {#if c.shortcut}<kbd>{c.shortcut}</kbd>{/if}
                    </Command.Item>
                  {/each}
                </Command.GroupItems>
              </Command.Group>
            {/each}
          </Command.Viewport>
        </Command.List>
      </Command.Root>
    </Dialog.Content>
  </Dialog.Portal>
</Dialog.Root>
```

```ts
// web/src/lib/commands.svelte.ts
import { SvelteMap } from "svelte/reactivity";

export type CommandDef = {
  id: string;
  title: string;
  group: string;
  keywords?: string[];
  shortcut?: string;
  enabled?: () => boolean;
  run: () => void | Promise<void>;
};

class Commands {
  #defs = new SvelteMap<string, CommandDef>();

  register(def: CommandDef): () => void {
    this.#defs.set(def.id, def);
    return () => {
      this.#defs.delete(def.id);
    };
  }

  groups(): Array<{ name: string; items: CommandDef[] }> {
    const byGroup = new Map<string, CommandDef[]>();
    for (const d of this.#defs.values()) {
      if (d.enabled && !d.enabled()) continue;
      (byGroup.get(d.group) ?? byGroup.set(d.group, []).get(d.group))?.push(d);
    }
    return [...byGroup].map(([name, items]) => ({ name, items: items.sort((a, b) => a.title.localeCompare(b.title)) }));
  }

  run(id: string): void {
    void this.#defs.get(id)?.run();
  }
}

export const commands = new Commands();
```

- Each panel registers its commands in `$effect` and returns the unregister function.
- Default commands: `layout.reset`, `panel.open.<name>`, `panel.close.<name>`, `panel.focus.<name>`, `panel.float`, `panel.popout`, `grid.columns`, `selection.clear`, `theme.toggle`.
- A command title is one verb and an object: Reset layout, Open inspector, Clear selection.
- Use `SvelteMap` from `svelte/reactivity` for the registry, so that the palette sees a new command.
- Give each `Command.Item` the title in `keywords`.
  The filter matches on `value` and `keywords`, and `value` is the id.
- Style the parts through their data attributes: `[data-command-input]`, `[data-command-list]`, `[data-command-item]`, and `[data-command-item][data-selected]`.
- Bits UI renders the palette in a portal with the dialog semantics.
  Style `.palette` with the tokens, `width: min(90vw, 36rem)`, and a `max-height: 60dvh` list.

## The bulk-edit panel

```svelte
<!-- web/src/panels/BulkEdit.svelte -->
<script lang="ts">
  import { inventory } from "../lib/inventory.svelte";
  import { selection } from "../lib/selection.svelte";

  let status = $state<"" | "active" | "retired">("");
  let busy = $state(false);
  let result = $state<string | undefined>();

  async function apply() {
    busy = true;
    try {
      await inventory.patch([...selection.ids], { status });
      result = `Changed ${selection.ids.size} rows.`;
    } catch {
      result = "The change failed. No row changed.";
    } finally {
      busy = false;
    }
  }
</script>

<section>
  <p class="summary">{selection.ids.size} rows selected</p>
  <label>Status
    <select bind:value={status}>
      <option value="">Keep</option>
      <option value="active">active</option>
      <option value="retired">retired</option>
    </select>
  </label>
  <button onclick={apply} disabled={busy || selection.ids.size === 0 || status === ""}>Apply</button>
  {#if result}<p role="status">{result}</p>{/if}
</section>
```

- The panel stays open.
  The selection in the grid drives it.
- The button holds one verb.
  The count and the target stand in the summary line beside it.
- One `PATCH` changes all rows in one transaction.
  The server pushes the changed rows, and the grid shows them.
- Above 500 rows, the server answers 202 and reports the progress over the topic `jobs.<id>`.
- When the application shows system state, the `scuq-hmi-semantics` skill gives the confirmation rule for a bulk change.

## Keyboard and focus

- Ctrl+K opens the palette from every panel.
- Each panel has a `focus panel` command, and the dock gives the panel element `tabindex="-1"`.
  The command calls `panel.api.setActive()`, and then focuses `panel.view.content.element` in a `requestAnimationFrame` callback, because `setActive` moves the element in the DOM.
- Run a palette command in the `onCloseAutoFocus` handler of the Bits UI `Dialog.Content`, and call `preventDefault` there.
  Bits UI restores the focus to the opener on close, and that undoes a focus command that ran before.
- Inside the grid, the arrow keys move the cell focus, Enter edits, Escape cancels, and Space toggles the selection.
- A floating or popout panel keeps its commands.
- Test the path with Playwright: open the palette, run `panel.focus.inspector`, and expect `.panel-host:focus-within` to hold the inspector.
- RevoGrid adds an attribution link to the grid under its MIT license terms.
  Keep it, or read the attribution page of RevoGrid for the conditions.
