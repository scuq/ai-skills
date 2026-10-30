# Review Checklist

Mark each item pass, fail, or not applicable.
An item marked (C1) is mandatory only for class C1.

## Registry

- [ ] R1 All states, alert levels, icons, and labels come from the registry.
- [ ] R2 The component code has no raw color and no ad hoc icon for a semantic purpose.
- [ ] R3 No synonyms exist. Each term maps to exactly one state.
- [ ] R4 (C1) The registry lists each deviation, with the reason and the approver.

## Perception

- [ ] P1 Every significant state has a text label, a shape or an icon, and a color. This is redundant coding.
- [ ] P2 The interface stays correct in grayscale and under a simulation of deuteranopia or protanopia.
- [ ] P3 The text and icon contrast meet WCAG 2.2 AA at minimum, and AAA for a C1 critical indicator.
- [ ] P4 The alert colors are not reused for branding, links, or decoration.
- [ ] P5 The NORMAL state has low salience. The abnormal states stand out.

## State truth

- [ ] S1 UNKNOWN and STALE are distinct states with distinct coding. Neither state ever shows as OK.
- [ ] S2 The data age is visible where freshness matters. The staleness thresholds are defined.
- [ ] S3 The interface shows the commanded state and the actual state separately until they match.
- [ ] S4 The loading state is distinguishable from the empty state and from the error state.
- [ ] S5 Each timestamp shows a time zone. Each value shows its unit and, where relevant, its threshold.
- [ ] S6 Each count and each trend shows its coverage. An asset without a scan shows as unknown, not as clean.

## Layout and navigation

- [ ] L1 Information that needs operator action is visible without navigation, scrolling, or hovering.
- [ ] L2 The alert area has the same position in every view.
- [ ] L3 Each alert links directly to the affected object.
- [ ] L4 The same element types sit in the same positions across views.
- [ ] L5 The layout follows the CSS width, the zoom, and the pointer type, not the physical resolution. The content reflows at a width of 320 CSS pixels.
- [ ] L6 Density changes only the spacing and the secondary fields. Each layout keeps the state coding and the required fields.
- [ ] L7 No drawer, dialog, or action bar covers the alert area or the focused element. Each panel shows its own data age.
- [ ] L8 Each control works with touch and with a keyboard. Each drag action has a method without a drag.

## View truth

- [ ] V1 The interface shows the active filters. The matched count and the rows agree with the filters.
- [ ] V2 The interface shows the count of action-required records outside the current view.
- [ ] V3 A user cannot hide a required field. A saved view cannot hide the alert area.
- [ ] V4 The interface shows where the filter ran. A client filter runs only on a fully loaded set of records.
- [ ] V5 The loading state and the failure state of a filter are not "no results". During a request or after a failure, the interface tells the operator that the rows show the previous filter.

## Controls

- [ ] K1 A button has one short verb. A control that changes a state has an always-visible action summary line beside it, with the target, the count, and the environment.
- [ ] K2 Each control sits next to what it affects.
- [ ] K3 A destructive action names the target, the environment, and the blast radius.
- [ ] K4 The safe option is the default. Confirmation is reserved for a consequential action.
- [ ] K5 (C1) A high-impact action requires a typed target or a two-step commit, and offers undo or a staged apply where possible.
- [ ] K6 After an action, the interface shows the resulting actual state, not only "success".
- [ ] K7 Each module uses the same bulk flow: select, configure, preview, apply, verify.
- [ ] K8 The preview shows the selection scope, the count for each environment, the change to each record, and the excluded records.
- [ ] K9 Apply reads the state again and reports the changes after the preview. The result shows the commanded state, the actual state, and the status of each record. A partial failure has its own state.
- [ ] K10 A keyboard shortcut and a command palette pass the same gate as a mouse action. Selection follows the record ID, not the row position.
- [ ] K11 A modal dialog holds one focused action. The alert area stays visible above the backdrop.
- [ ] K12 A modal dialog puts the first focus on the safe option and keeps the Tab key inside. It closes with Escape or Cancel, and returns the focus. A click on the backdrop does not close it.
- [ ] K13 A new alert does not open a dialog and does not take the focus.

## Alerting

- [ ] A1 Every alert has a defined operator response.
- [ ] A2 Ack and clear are separate. An acked alert stays visible while it is active.
- [ ] A3 Consequential alarms are grouped under their root cause.
- [ ] A4 Hysteresis or a deadband prevents flapping.
- [ ] A5 A toast never carries a system alert alone.
- [ ] A6 A severity uses an alert color only if the registry maps it to an alert level. Each record of that severity has the operator response. A severity without an alert level uses the muted severity color from the registry.
