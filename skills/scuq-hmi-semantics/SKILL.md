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
- 3b. Unknown is not OK. Give missing, stale, or unreachable data its own explicit state, for example `UNKNOWN` or `STALE`, with its own coding. Never show it as green. Never leave it blank. Never show the last good value without an age indicator. Show the data age, for example "updated 4 min ago", wherever freshness matters. Define a staleness threshold for each data source.
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
- 3e. Critical information is never hidden. In class C1 or C2, show anything that needs operator action without navigation, scrolling, hovering, or expanding. Use one alert area, in the same position in every view. Link each alert directly to the affected object.
- 3f. Control-response mapping. Place a control next to what it affects. State the effect of the control in its label, for example "Disable port Gi1/0/12", not "Apply". Make similar controls behave the same way everywhere.
- 3g. Consequential actions. For a destructive action or an action with a wide impact:

  - Name the exact target: the hostname, the ID, or the environment.
  - State the consequence and the blast radius.
  - Make the safe option the default.
  - Prefer undo or a staged apply over a confirmation question such as "Are you sure?"

  For class C1, require the operator to type the target name, or add a second step. Do not use a confirmation dialog for a routine action. Confirmation fatigue destroys its value.
- 3h. Numbers carry context.

  - Always show the unit.
  - Use a sensible precision.
  - Show the time zone on a timestamp.
  - Show the threshold next to a value when something judges the value against it.
- 3i. Alarm discipline. Every alert must have an operator response. If nobody needs to act on it, it is an event or a log line, not an alert. Avoid alarm floods: group consequential alarms, suppress them under their root cause, and show the root first. See `references/alert-contract.md`.

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
