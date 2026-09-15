package nagios

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestResultSingleLine(t *testing.T) {
	r := NewResult("DISK")
	r.Raise(Warning)
	r.Summaryf("free space: / %d MB (%d%%)", 3326, 56)
	r.AddPerfdata(Perfdata{Label: "/", Value: 2643, UOM: "MB", Min: Float(0), Max: Float(5968)})
	want := "DISK WARNING: free space: / 3326 MB (56%) | /=2643MB;;;0;5968\n"
	if got := r.String(); got != want {
		t.Errorf("String() =\n%q\nwant\n%q", got, want)
	}
	if r.Status() != Warning {
		t.Errorf("Status() = %v, want WARNING", r.Status())
	}
}

func TestResultMultiline(t *testing.T) {
	r := NewResult("DISK")
	r.Summaryf("3 filesystems")
	r.Linef("/ 15272 MB (77%%)\n/boot 68 MB (69%%)")
	r.Linef("/home a|b")
	r.AddPerfdata(Perfdata{Label: "/", Value: 2643, UOM: "MB"})
	r.AddPerfdata(Perfdata{Label: "/boot", Value: 68, UOM: "MB"})
	want := "DISK OK: 3 filesystems\n/ 15272 MB (77%)\n/boot 68 MB (69%)\n/home a/b | /=2643MB /boot=68MB\n"
	if got := r.String(); got != want {
		t.Errorf("String() =\n%q\nwant\n%q", got, want)
	}

	r.PerfdataOnFirstLine = true
	want = "DISK OK: 3 filesystems | /=2643MB /boot=68MB\n/ 15272 MB (77%)\n/boot 68 MB (69%)\n/home a/b\n"
	if got := r.String(); got != want {
		t.Errorf("PerfdataOnFirstLine String() =\n%q\nwant\n%q", got, want)
	}
}

func TestResultFail(t *testing.T) {
	r := NewResult("HTTP")
	r.Summaryf("ignored")
	r.Fail(Critical, errors.New("connection refused"))
	r.AddPerfdata(Perfdata{Label: "bad=label"})
	got := r.String()
	if r.Status() != Critical {
		t.Errorf("Status() = %v, want CRITICAL", r.Status())
	}
	if !strings.HasPrefix(got, "HTTP CRITICAL: connection refused, perfdata label") {
		t.Errorf("String() = %q", got)
	}
}

func TestDebugf(t *testing.T) {
	r := NewResult("")
	r.SetVerbosity(1)
	r.Debugf(1, "shown")
	r.Debugf(2, "hidden")
	if got, want := r.String(), "OK\nshown\n"; got != want {
		t.Errorf("String() = %q, want %q", got, want)
	}
}

func TestParseTimeout(t *testing.T) {
	tests := map[string]Timeout{
		"10":         {10 * time.Second, Critical},
		"30:UNKNOWN": {30 * time.Second, Unknown},
		"5:warning":  {5 * time.Second, Warning},
	}
	for in, want := range tests {
		got, err := ParseTimeout(in)
		if err != nil || got != want {
			t.Errorf("ParseTimeout(%q) = %v, %v, want %v", in, got, err, want)
		}
	}
	for _, in := range []string{"", "0", "-1", "1.5", "10:BAD"} {
		if _, err := ParseTimeout(in); err == nil {
			t.Errorf("ParseTimeout(%q): no error, want an error", in)
		}
	}
}

func TestExecuteDeadline(t *testing.T) {
	r := NewResult("TCP")
	out, status := execute(r, Timeout{Duration: 50 * time.Millisecond, State: Unknown}, func(ctx context.Context) {
		<-ctx.Done()
		r.Fail(Critical, ctx.Err())
	})
	if status != Unknown || !strings.HasPrefix(out, "TCP UNKNOWN: timeout after 50ms") {
		t.Errorf("execute = %q, %v", out, status)
	}
}

func TestExecuteHang(t *testing.T) {
	block := make(chan struct{})
	defer close(block)
	start := time.Now()
	out, status := execute(NewResult("TCP"), Timeout{Duration: 10 * time.Millisecond, State: Critical}, func(context.Context) {
		<-block
	})
	if status != Critical || !strings.Contains(out, "did not stop") {
		t.Errorf("execute = %q, %v", out, status)
	}
	if time.Since(start) > 3*time.Second {
		t.Errorf("execute took %s", time.Since(start))
	}
}

func TestExecutePanic(t *testing.T) {
	out, status := execute(NewResult("X"), DefaultTimeout, func(context.Context) {
		panic("boom")
	})
	if status != Unknown || out != "X UNKNOWN: plugin panic: boom\n" {
		t.Errorf("execute = %q, %v", out, status)
	}
}
