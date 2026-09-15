package nagios

import (
	"errors"
	"fmt"
	"math"
	"strings"
)

// UOMs holds the units of measurement that [Perfdata.Validate] accepts.
// The set is the list from the monitoring-plugins development guidelines.
// A plugin can add a unit before the first check, for example "packets"
// for Icinga 2.
var UOMs = map[string]bool{
	"":   true, // a number of things
	"s":  true,
	"ms": true,
	"us": true,
	"%":  true,
	"B":  true,
	"KB": true,
	"MB": true,
	"GB": true,
	"TB": true,
	"c":  true, // a continuous counter
}

// A Perfdata is one performance data metric in the format
// 'label'=value[UOM];[warn];[crit];[min];[max].
type Perfdata struct {
	// Label is the name of the metric. It cannot contain "=", "'", "|",
	// or a line break. Use [SanitizeLabel] for a label from external data.
	Label string
	// Value is the measured value.
	Value float64
	// Undetermined is true when the plugin cannot get the value.
	// The output then shows "U" in place of the value.
	Undetermined bool
	// UOM is the unit of Value, Min, Max, and the thresholds.
	UOM string
	// Warning and Critical are the threshold ranges. Unset ranges give
	// empty fields.
	Warning  Range
	Critical Range
	// Min and Max are the limits of Value. Nil gives an empty field.
	Min *float64
	Max *float64
}

// Float returns a pointer to v, for [Perfdata.Min] and [Perfdata.Max].
func Float(v float64) *float64 {
	return &v
}

// SanitizeLabel returns s with each "=", "'", "|", and line break
// changed to "_".
func SanitizeLabel(s string) string {
	return strings.NewReplacer("=", "_", "'", "_", "|", "_", "\n", "_", "\r", "_").Replace(s)
}

// Validate returns an error if p cannot give valid performance data.
func (p Perfdata) Validate() error {
	switch {
	case p.Label == "":
		return errors.New("perfdata label is empty")
	case strings.ContainsAny(p.Label, "='|\r\n"):
		return fmt.Errorf("perfdata label %q contains \"=\", \"'\", \"|\", or a line break", p.Label)
	case !UOMs[p.UOM]:
		return fmt.Errorf("perfdata %q: unit %q is not in UOMs", p.Label, p.UOM)
	case !p.Undetermined && !isFinite(p.Value):
		return fmt.Errorf("perfdata %q: value is not a finite number, set Undetermined", p.Label)
	case p.Min != nil && !isFinite(*p.Min):
		return fmt.Errorf("perfdata %q: min is not a finite number", p.Label)
	case p.Max != nil && !isFinite(*p.Max):
		return fmt.Errorf("perfdata %q: max is not a finite number", p.Label)
	}
	return nil
}

// String returns p in the performance data format.
// It drops the empty fields at the end.
func (p Perfdata) String() string {
	label := p.Label
	if strings.Trim(label, "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_./-") != "" {
		label = "'" + label + "'"
	}
	fields := make([]string, 5)
	if p.Undetermined {
		fields[0] = "U"
	} else {
		fields[0] = formatNumber(p.Value) + p.UOM
	}
	fields[1] = p.Warning.String()
	fields[2] = p.Critical.String()
	if p.Min != nil {
		fields[3] = formatNumber(*p.Min)
	}
	if p.Max != nil {
		fields[4] = formatNumber(*p.Max)
	}
	return label + "=" + strings.TrimRight(strings.Join(fields, ";"), ";")
}

func isFinite(v float64) bool {
	return !math.IsNaN(v) && !math.IsInf(v, 0)
}
