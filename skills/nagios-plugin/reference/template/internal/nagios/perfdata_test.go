package nagios

import (
	"math"
	"testing"
)

func mustRange(t *testing.T, s string) Range {
	t.Helper()
	r, err := ParseRange(s)
	if err != nil {
		t.Fatal(err)
	}
	return r
}

func TestPerfdataString(t *testing.T) {
	tests := []struct {
		p    Perfdata
		want string
	}{
		{Perfdata{Label: "users", Value: 3}, "users=3"},
		{Perfdata{Label: "time", Value: 0.25, UOM: "s", Warning: mustRange(t, "1"), Critical: mustRange(t, "2"), Min: Float(0)},
			"time=0.25s;1;2;0"},
		{Perfdata{Label: "/boot", Value: 68, UOM: "MB", Warning: mustRange(t, "88"), Critical: mustRange(t, "93"), Min: Float(0), Max: Float(98)},
			"/boot=68MB;88;93;0;98"},
		{Perfdata{Label: "free space", Value: 56, UOM: "%"}, "'free space'=56%"},
		{Perfdata{Label: "temp", Undetermined: true, Max: Float(90)}, "temp=U;;;;90"},
		{Perfdata{Label: "load", Value: 1e-7, Critical: mustRange(t, "@10:20")}, "load=0.0000001;;@10:20"},
	}
	for _, tt := range tests {
		if err := tt.p.Validate(); err != nil {
			t.Errorf("Validate(%+v): %v", tt.p, err)
		}
		if got := tt.p.String(); got != tt.want {
			t.Errorf("String() = %q, want %q", got, tt.want)
		}
	}
}

func TestPerfdataValidate(t *testing.T) {
	bad := []Perfdata{
		{Label: ""},
		{Label: "a=b"},
		{Label: "it's"},
		{Label: "a|b"},
		{Label: "x", UOM: "furlongs"},
		{Label: "x", Value: math.NaN()},
		{Label: "x", Value: math.Inf(1)},
		{Label: "x", Min: Float(math.Inf(-1))},
	}
	for _, p := range bad {
		if err := p.Validate(); err == nil {
			t.Errorf("Validate(%+v): no error, want an error", p)
		}
	}
}

func TestSanitizeLabel(t *testing.T) {
	if got := SanitizeLabel("a=b'c|d\ne"); got != "a_b_c_d_e" {
		t.Errorf("SanitizeLabel = %q", got)
	}
}
