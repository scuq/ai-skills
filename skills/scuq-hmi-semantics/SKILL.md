---
name: scuq-hmi-semantics
description: >
  Rules for the meaning of a state, an alert, and a control in an
  operator interface. Load this skill when you build, extend, or
  review an interface that shows a system state or lets an operator
  change one: a dashboard, a monitoring view, a NOC console, a SOC
  console, an admin panel, a network tool, a firewall tool, a NAC
  tool, alerting, a status page, CLI status output, TUI status
  output, or a control panel. Load it also when you define a design
  token, a status color, an icon, an alert level, or a confirmation
  dialog. Load it even when the user asks only for "a dashboard" or
  "a status page" and does not mention safety.
---

# HMI Semantics

Treat the user interface as part of the control loop.
It is not decoration around the loop.

```
system state → semantic meaning → representation → perception → decision → action → system state
```

Every step in this loop can fail.
This skill helps the operator find the correct system state fast.
It also makes the consequence of each action clear before the action runs.
Aesthetics come second.
For visual polish, combine this skill with a frontend skill or a design skill.
When the two skills disagree, semantics win.

## Workflow

Follow these steps in order.
For a small interface, keep each step short, but do not skip a step.

### 1. Classify the interface

Ask the user what happens if the operator misreads the interface.
If the user does not say, infer it from the context.

| Class | Misreading causes | Rigor |
|---|---|---|
| C1 critical | An outage, a security breach, data loss, or harm to a person | The full workflow, the full checklist, and documented deviations |
| C2 operational | Degraded service, a wrong ticket, or wasted hours | The full workflow, and the checklist without a formal deviation log |
| C3 informational | Minor confusion | The core rules only, sections 3a to 3d |

State the class you chose in your response.
The user can then correct it.

### 2. Define the semantic registry before any component code

Create a single source of truth, for example a file named `semantics.yaml`.
Extend it if one already exists.
See `assets/semantics.template.yaml` for a starting point.
It holds:

- The state vocabulary: every state an entity can show, with its meaning.
- The alert levels, each with its full contract. See `references/alert-contract.md`.
- The icon library: one icon for each meaning, one meaning for each icon.
- The label plan: the canonical term for each concept. Do not use two labels for one state. For example, "down", "offline", and "unreachable" must map to distinct states, or to one state.

Generate CSS tokens, enums, and internationalization keys from the registry.
A component reads a semantic value, for example `status="warning"`.
A component never reads a raw style value, for example `color="red"`.
If a developer needs a new icon or a new state, add it to the registry first.

### 3. Apply the core rules

- 3a. Redundant coding. Color never carries meaning alone. Every state with an operational effect uses at least two more channels: a text label, a shape or an icon, and position. Test the interface in grayscale and for red-green color blindness. The interface must stay correct in both.
- 3b. Unknown is not OK. Give missing, stale, or unreachable data its own explicit state, for example `UNKNOWN` or `STALE`, with its own coding. Never show it as green. Never leave it blank. Never show the last good value without an age indicator. Show the data age, for example "updated 4 min ago", wherever freshness matters. Define a staleness threshold for each data source. Zero is a claim: show the coverage next to a count or a trend. An asset without a scan is unknown, not clean, and a falling count can mean that a scanner stopped.
- 3c. Commanded is not actual. When the interface can change a state, show the intended state and the observed state separately until they match:

  ```
  COMMAND   shutdown port Gi1/0/12
  ACTUAL    up (last polled 00:00:08 ago)
  STATUS    PENDING — mismatch
  ```

  Never show a requested state as the confirmed state. Apply the same rule to:

  - Configuration intent and the running configuration.
  - The desired replica count and the current count.
  - A planned change and a deployed change.
- 3d. Consistency. Use the same color, icon, word, and position for the same concept. Keep this consistency within one screen, across the application, and across other applications of the user that share the registry. Never reuse a color reserved for an alert level for branding, links, or decoration.
- 3e. Critical information is never hidden. In class C1 or C2, show anything that needs operator action without navigation, scrolling, hovering, or expanding. Use one alert area, in the same position in every view. Link each alert directly to the affected object. You can hide navigation labels, help, secondary fields, and rare actions. Never hide a state, a consequence, a target, or a data age. A tooltip is never the only path to information, because touch input has no hover.
- 3f. Control-response mapping. Place a control next to what it affects. Give a button one short verb. Never put the target, the count, or the scope on the button. On a control that changes a state, use a specific verb, for example "Disable" or "Patch", not "Apply" or "OK". Make similar controls behave the same way everywhere. Use a drawer to inspect a record, a dialog for one focused action, and a page for a complex workflow.

  Show the effect of a control that changes a state in an action summary line beside the control:

  ```
  Disable port Gi1/0/12 on sw01 (prod)                [Cancel] [Disable]
  24 assets: 18 prod, 6 dev. Install at 02:00 CET.    [Cancel] [Patch]
  ```

  - The line gives the target, the count, the environment, and the time of a scheduled action.
  - The line updates with the selection and is always visible.
  - The button points to the line with `aria-describedby`, so a screen reader gives the full effect.
  - A control that only opens a view or moves to a step changes no state, so it needs no line. Examples are "New scan" and "Next".
  - A row control gets its target from its row. Its "..." menu shows the row name as the heading.
- 3g. Consequential actions. For a destructive action or an action with a wide impact:

  - Name the exact target: the hostname, the ID, or the environment.
  - State the consequence and the blast radius.
  - Make the safe option the default.
  - Prefer undo or a staged apply over a confirmation question such as "Are you sure?"

  For class C1, require the operator to type the target name, or add a second step. Do not use a confirmation dialog for a routine action. Confirmation fatigue destroys its value.

  A keyboard shortcut and a command palette pass the same gate as a mouse action. Selection follows the record ID, not the row position. A live update does not move a row under the pointer during a selection.
- 3h. Numbers carry context.

  - Always show the unit.
  - Use a sensible precision.
  - Show the time zone on a timestamp.
  - Show the threshold next to a value when something judges the value against it.
- 3i. Alarm discipline. Every alert must have an operator response. If nobody needs to act on it, it is an event or a log line, not an alert. Avoid alarm floods: group consequential alarms, suppress them under their root cause, and show the root first. See `references/alert-contract.md`.
- 3j. Severity is not an alert. A severity classifies a record, for example a finding with a CVSS score. An alert asks a person to act. Map a severity to an alert level in the registry only if each record of that severity has the operator response. Then the severity uses the alert color. Otherwise, show the severity with text and a neutral color. A record with an active alert shows the level of that alert in its row.
- 3k. View truth. A filter, a search, a page, a saved view, or a hidden column never hides the alert area.

  - Show the active filters.
  - Show the matched count next to the total, and make the count and the rows agree with the filters.
  - Show the count of action-required records outside the view.
  - Mark in the registry the fields of an entity that a user cannot hide.
  - Show where the filter ran. A client filter shows the loaded count next to the total. A server filter shows the time of the result.
  - If a client filter holds only a part of the records, do not show its count as the result. Send the filter to the server.
  - While a request runs, show a loading state, and tell the operator that the rows show the previous filter.
  - If a request fails, show the failure and a retry control. Never show a failure or a loading state as "no results".
<!-- ste-lint: disable-next-line synonym-rotation -->
- 3l. Bulk changes. Use the same flow in every module: select, configure, preview, apply, verify.

  - Select: show the selection scope, with the records outside the current page and the count for each environment.
  - Configure: show the effect in the action summary line of rule 3f.
  - Preview: show the current value and the new value of each record, and show the excluded records with the reason.
  - Apply: read the state again, and report each record that changed after the preview.
  - Verify: show the commanded state, the actual state, and the status of each record. A partial failure is a state of its own, not a success.
- 3m. Adaptive layout. Select the layout from the CSS width, the zoom, and the pointer type, not from the physical resolution.

  - Density changes the spacing and the secondary fields, never the state coding.
  - Each layout keeps the required fields of the registry, in the registry order.
  - The alert area stays visible in each layout. A drawer, a dialog, or an action bar never covers the alert area or the focused element.
  - Each panel shows its own data age.
  - Each control works with touch and with a keyboard. Each drag action has a method without a drag.
- 3n. Modal dialogs. Use a modal dialog only for one focused action. Keep the alert area visible and above the backdrop, so do not use the browser top layer of `showModal()`.

  - Put the first focus on the safe option.
  - Keep the Tab key in the dialog, and make the content behind the dialog inert.
  - Close the dialog with Escape or Cancel. A click on the backdrop does not close it, so a stray click cannot lose input.
  - After the dialog closes, return the focus to the control that opened it.
  - A new alert shows in the alert area. It does not open a dialog, and it does not take the focus.

### 4. Review

Run `references/review-checklist.md` against the result.
Report the findings in a short table with the columns rule, status, and change.
Status is pass, fail, or not applicable.

### 5. Document deviations (C1)

If you break a rule on purpose, record it in the registry under `deviations:`, with the reason and the approver.
An undocumented deviation is a bug.

## Output expectations

When you produce interface code with this skill, also give:

1. The criticality class, with a one-line reason.
2. The registry, new or as a difference from the existing registry.
3. The checklist result table.

Keep explanations short.
The registry and the code are the deliverables.

## References

- `references/alert-contract.md`: The alert levels, their contracts, and the alarm management rules. Read it when the interface has alerting or a status severity.
- `references/review-checklist.md`: The review checklist. Read it in step 4.
- `references/sources.md`: The standards behind these rules. Read it when the user asks for justification, or wants more detail.
- `assets/semantics.template.yaml`: The starting point for the registry.
- `assets/micro-parts.html`: Small HTML and CSS parts that hold the semantics, with registry tokens only. Read it when you build an interface. Do not read it for a review.
