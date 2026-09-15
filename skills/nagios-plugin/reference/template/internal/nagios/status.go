// Package nagios implements the plugin API of Nagios and its derivatives:
// exit codes, threshold ranges, performance data, multiline output,
// timeouts, and encrypted credentials.
package nagios

import (
	"fmt"
	"strings"
)

// A Status is the exit code of a plugin.
type Status int

// The exit codes of the plugin API.
const (
	OK       Status = 0
	Warning  Status = 1
	Critical Status = 2
	Unknown  Status = 3
)

// String returns the name of s in uppercase, for example "WARNING".
// A value that is not an exit code of the plugin API gives "UNKNOWN".
func (s Status) String() string {
	switch s {
	case OK:
		return "OK"
	case Warning:
		return "WARNING"
	case Critical:
		return "CRITICAL"
	default:
		return "UNKNOWN"
	}
}

// ParseStatus returns the status with the name s.
// The name is not case sensitive, for example "unknown" or "CRITICAL".
func ParseStatus(s string) (Status, error) {
	switch strings.ToUpper(s) {
	case "OK":
		return OK, nil
	case "WARNING":
		return Warning, nil
	case "CRITICAL":
		return Critical, nil
	case "UNKNOWN":
		return Unknown, nil
	}
	return Unknown, fmt.Errorf("status %q is not OK, WARNING, CRITICAL, or UNKNOWN", s)
}

// rank orders the statuses by severity, from low to high:
// OK, UNKNOWN, WARNING, CRITICAL.
// This is the order of max_state_alt in the monitoring-plugins project.
var rank = map[Status]int{OK: 0, Unknown: 1, Warning: 2, Critical: 3}

// Worst returns the status of a and b with the higher severity.
func Worst(a, b Status) Status {
	if rank[b] > rank[a] {
		return b
	}
	return a
}
