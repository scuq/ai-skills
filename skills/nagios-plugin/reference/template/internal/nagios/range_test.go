package nagios

import (
	"math"
	"testing"
)

func TestParseRangeAlert(t *testing.T) {
	// The cases are the example table of the monitoring-plugins guidelines.
	tests := []struct {
		in    string
		alert []float64
		ok    []float64
		str   string
	}{
		{"10", []float64{-1, 10.1, 11}, []float64{0, 5, 10}, "10"},
		{"10:", []float64{9.9, -5}, []float64{10, 1e9}, "10:"},
		{"~:10", []float64{10.1, 11}, []float64{-1e9, 10}, "~:10"},
		{"10:20", []float64{9, 21}, []float64{10, 15, 20}, "10:20"},
		{"@10:20", []float64{10, 15, 20}, []float64{9, 21}, "@10:20"},
		{"-5:-1", []float64{0, -6}, []float64{-5, -1}, "-5:-1"},
		{"0:", []float64{-0.1}, []float64{0, 1e9}, "0:"},
		{"1.5:2.5", []float64{1.4, 2.6}, []float64{1.5, 2.5}, "1.5:2.5"},
	}
	for _, tt := range tests {
		r, err := ParseRange(tt.in)
		if err != nil {
			t.Fatalf("ParseRange(%q): %v", tt.in, err)
		}
		for _, v := range tt.alert {
			if !r.Alert(v) {
				t.Errorf("ParseRange(%q).Alert(%v) = false, want true", tt.in, v)
			}
		}
		for _, v := range tt.ok {
			if r.Alert(v) {
				t.Errorf("ParseRange(%q).Alert(%v) = true, want false", tt.in, v)
			}
		}
		if got := r.String(); got != tt.str {
			t.Errorf("ParseRange(%q).String() = %q, want %q", tt.in, got, tt.str)
		}
	}
}

func TestParseRangeErrors(t *testing.T) {
	for _, in := range []string{"@", "20:10", "abc", "inf", "NaN", "0x10", "1:2:3", "~"} {
		if _, err := ParseRange(in); err == nil {
			t.Errorf("ParseRange(%q): no error, want an error", in)
		}
	}
}

func TestUnsetRange(t *testing.T) {
	r, err := ParseRange("")
	if err != nil {
		t.Fatal(err)
	}
	if r.IsSet() || r.Alert(math.MaxFloat64) || r.String() != "" {
		t.Errorf("unset range: IsSet=%v Alert=%v String=%q", r.IsSet(), r.Alert(math.MaxFloat64), r.String())
	}
}

func TestEvaluate(t *testing.T) {
	th, err := ParseThresholds("5", "10")
	if err != nil {
		t.Fatal(err)
	}
	for v, want := range map[float64]Status{1: OK, 5: OK, 6: Warning, 10: Warning, 11: Critical, -1: Critical} {
		if got := th.Evaluate(v); got != want {
			t.Errorf("Evaluate(%v) = %v, want %v", v, got, want)
		}
	}
}

func TestWorst(t *testing.T) {
	tests := []struct{ a, b, want Status }{
		{OK, Unknown, Unknown},
		{Unknown, Warning, Warning},
		{Critical, Unknown, Critical},
		{Warning, OK, Warning},
	}
	for _, tt := range tests {
		if got := Worst(tt.a, tt.b); got != tt.want {
			t.Errorf("Worst(%v, %v) = %v, want %v", tt.a, tt.b, got, tt.want)
		}
	}
}
