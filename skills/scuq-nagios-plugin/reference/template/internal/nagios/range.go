package nagios

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

// A Range is a threshold range in the format [@]start:end.
//
// The zero value is an unset range. An unset range never raises an alert.
type Range struct {
	// Start is the lower end of the range. It is negative infinity for "~".
	Start float64
	// End is the upper end of the range. It is positive infinity when the
	// range has no end.
	End float64
	// Inside is true when the range starts with "@". Then a value inside
	// the range raises an alert. Otherwise a value outside the range
	// raises an alert.
	Inside bool

	set bool
}

// ParseRange returns the range that s describes.
// If s is empty, ParseRange returns an unset range.
// The ends of the range are inclusive:
//
//	10      alert if x < 0 or x > 10
//	10:     alert if x < 10
//	~:10    alert if x > 10
//	10:20   alert if x < 10 or x > 20
//	@10:20  alert if 10 <= x <= 20
func ParseRange(s string) (Range, error) {
	if s == "" {
		return Range{}, nil
	}
	r := Range{set: true, End: math.Inf(1)}
	body := s
	if strings.HasPrefix(body, "@") {
		r.Inside = true
		body = body[1:]
	}
	start, end, hasColon := strings.Cut(body, ":")
	if !hasColon {
		start, end = "", start
		if end == "" {
			return Range{}, fmt.Errorf("range %q has no end", s)
		}
	}
	switch start {
	case "":
		r.Start = 0
	case "~":
		r.Start = math.Inf(-1)
	default:
		v, err := parseNumber(start)
		if err != nil {
			return Range{}, fmt.Errorf("range %q: start: %w", s, err)
		}
		r.Start = v
	}
	if end != "" {
		v, err := parseNumber(end)
		if err != nil {
			return Range{}, fmt.Errorf("range %q: end: %w", s, err)
		}
		r.End = v
	}
	if r.Start > r.End {
		return Range{}, fmt.Errorf("range %q: start is greater than end", s)
	}
	return r, nil
}

// IsSet reports whether r came from a range that was not empty.
func (r Range) IsSet() bool {
	return r.set
}

// Alert reports whether v raises an alert for r.
func (r Range) Alert(v float64) bool {
	if !r.set {
		return false
	}
	inside := v >= r.Start && v <= r.End
	return inside == r.Inside
}

// String returns r in the shortest range format, for performance data.
// An unset range gives an empty string.
func (r Range) String() string {
	if !r.set {
		return ""
	}
	var b strings.Builder
	if r.Inside {
		b.WriteByte('@')
	}
	endInf := math.IsInf(r.End, 1)
	switch {
	case math.IsInf(r.Start, -1):
		b.WriteString("~:")
	case r.Start != 0 || endInf:
		b.WriteString(formatNumber(r.Start))
		b.WriteByte(':')
	}
	if !endInf {
		b.WriteString(formatNumber(r.End))
	}
	return b.String()
}

// Thresholds holds the warning range and the critical range of one metric.
type Thresholds struct {
	Warning  Range
	Critical Range
}

// ParseThresholds returns the thresholds for the warning range and the
// critical range. An empty string gives an unset range.
func ParseThresholds(warning, critical string) (Thresholds, error) {
	w, err := ParseRange(warning)
	if err != nil {
		return Thresholds{}, fmt.Errorf("warning: %w", err)
	}
	c, err := ParseRange(critical)
	if err != nil {
		return Thresholds{}, fmt.Errorf("critical: %w", err)
	}
	return Thresholds{Warning: w, Critical: c}, nil
}

// Evaluate returns the status of v.
// The critical range comes first, then the warning range:
//
//	Critical  if v raises an alert for the critical range
//	Warning   if v raises an alert for the warning range
//	OK        in all other cases
func (t Thresholds) Evaluate(v float64) Status {
	switch {
	case t.Critical.Alert(v):
		return Critical
	case t.Warning.Alert(v):
		return Warning
	}
	return OK
}

// parseNumber accepts only the characters of the class [-+0-9.], so that
// "inf", "NaN", and hexadecimal floats are errors.
func parseNumber(s string) (float64, error) {
	if strings.Trim(s, "-+0123456789.") != "" {
		return 0, fmt.Errorf("%q is not a decimal number", s)
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, fmt.Errorf("%q is not a decimal number", s)
	}
	return v, nil
}

// formatNumber writes v in decimal notation without an exponent, in the
// C locale.
func formatNumber(v float64) string {
	return strconv.FormatFloat(v, 'f', -1, 64)
}
