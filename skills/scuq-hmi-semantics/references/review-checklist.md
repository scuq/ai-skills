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

## Layout and navigation

- [ ] L1 Information that needs operator action is visible without navigation, scrolling, or hovering.
- [ ] L2 The alert area has the same position in every view.
- [ ] L3 Each alert links directly to the affected object.
- [ ] L4 The same element types sit in the same positions across views.

## Controls

- [ ] K1 A control label names the action and the target, for example "Disable port Gi1/0/12".
- [ ] K2 Each control sits next to what it affects.
- [ ] K3 A destructive action names the target, the environment, and the blast radius.
- [ ] K4 The safe option is the default. Confirmation is reserved for a consequential action.
- [ ] K5 (C1) A high-impact action requires a typed target or a two-step commit, and offers undo or a staged apply where possible.
- [ ] K6 After an action, the interface shows the resulting actual state, not only "success".

## Alerting

- [ ] A1 Every alert has a defined operator response.
- [ ] A2 Ack and clear are separate. An acked alert stays visible while it is active.
- [ ] A3 Consequential alarms are grouped under their root cause.
- [ ] A4 Hysteresis or a deadband prevents flapping.
- [ ] A5 A toast never carries a system alert alone.
