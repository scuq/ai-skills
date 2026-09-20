# Alert Contract

An alert level is a semantic object, not a CSS class.
Each level defines the full contract below.
Do not add a level without defining every field.

## Default levels

| Level | Meaning | Operator response | Visual | Persistence | Ack | Audio |
|---|---|---|---|---|---|---|
| EMERGENCY, WARNING | Immediate action necessary. Harm or an outage is in progress or imminent | Act now | Red. A unique filled shape, for example an octagon or a triangle. The one alert area. A text label | Stays while the condition exists | Required | A distinct tone. Optional in a web interface, and configurable |
| CAUTION | Action necessary soon. Degradation or risk is growing | Act within the defined time | Amber or yellow. A distinct shape. The one alert area. A text label | Stays while the condition exists | Required | Optional. Different from WARNING |
| ADVISORY | For awareness. No immediate action necessary | Note it, plan for it | Neutral or cyan. An outline icon. A text label | Can clear on its own when the condition clears | Optional | None |
| NORMAL | Within the expected range | None | Low salience, for example gray or muted. Never a wall of bright green | Not applicable | Not applicable | None |
| UNKNOWN, STALE | The state cannot be determined | Investigate the data source | A distinct pattern, for example a hatched or dashed outline, with a question mark icon. A text label | Stays until data returns | Not applicable | None |

Notes:

- Collapse EMERGENCY and WARNING into one level, unless the domain truly needs both. Fewer levels, clearly separated, work better than many levels.
- NORMAL stays deliberately quiet. This follows the ISA-101 idea of a high-performance HMI (human-machine interface): color marks only the abnormal states, so an abnormal state stands out. A screen full of green trains operators to stop looking at it.

## Required behavior

1. No silent disappearance. An active alert cannot disappear because the view changed, the list scrolled, or a toast notification timed out.
2. Ack does not mean clear. An acknowledgment means a person saw the alert. The alert stays visible, for example steady instead of flashing, until the condition clears. Track four states: `active-unacked`, `active-acked`, `cleared-unacked`, and `cleared`.
3. A cleared-unacked alert stays visible. A transient fault that already cleared still needs a first view.
4. Direct navigation. Each alert links to the affected object and, where possible, to the relevant control.
5. Attribution. Show the source, the first-seen time, the last-seen time, and a count if the alert repeats.
6. Ordering. Sort by level, then by the age of the unacknowledged condition. Never sort alphabetically.

## Alarm management rules

These rules follow the spirit of ISA-18.2 and EEMUA 191.

- Every alert has a defined operator response. Without a response, downgrade the alert to an event or a log line.
- Every alert has a defined cause and a documented owner.
- Suppress consequential alarms. When an uplink goes down, the 48 downstream "host unreachable" alarms are children of that event, not peers of it. Show the root cause, and collapse the children with a count.
- Use hysteresis or a deadband to prevent flapping. For example, raise the alarm at 90 percent, and clear it at 85 percent.
- Measure the alerts per operator per hour, the standing alerts, and the chattering alerts. A persistent standing alert is a defect, either in the system or in the alert definition.
- Make shelving, which is a temporary suppression, explicit, time-limited, and attributed to a person. Show a shelved alert as "shelved".

## Toasts and notifications

A toast confirms the effect of the action of the operator, for example "Port Gi1/0/12 disabled".
A toast is never the only carrier of a system alert.
