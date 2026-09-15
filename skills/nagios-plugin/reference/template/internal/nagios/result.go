package nagios

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// A Timeout is the time limit of a check and the status when the time
// limit ends.
type Timeout struct {
	Duration time.Duration
	State    Status
}

// DefaultTimeout is 10 seconds with the status Critical, the same as the
// plugins of the monitoring-plugins project.
var DefaultTimeout = Timeout{Duration: 10 * time.Second, State: Critical}

// ParseTimeout returns the timeout that s describes.
// The format is seconds[:STATE], for example "10" or "30:UNKNOWN".
// Without STATE, the status is the State of [DefaultTimeout].
func ParseTimeout(s string) (Timeout, error) {
	secs, state, hasState := strings.Cut(s, ":")
	n, err := strconv.Atoi(secs)
	if err != nil || n <= 0 {
		return Timeout{}, fmt.Errorf("timeout %q: seconds must be a positive integer", s)
	}
	t := Timeout{Duration: time.Duration(n) * time.Second, State: DefaultTimeout.State}
	if hasState {
		if t.State, err = ParseStatus(state); err != nil {
			return Timeout{}, fmt.Errorf("timeout %q: %w", s, err)
		}
	}
	return t, nil
}

// A Result collects the status, the text, and the performance data of one
// check.
//
// A Result is not safe for use by more than one goroutine at a time.
type Result struct {
	// PerfdataOnFirstLine puts all performance data on the first line,
	// also when there is long output. By default, performance data goes
	// after a "|" at the end of the long output when there is long output.
	PerfdataOnFirstLine bool

	name      string
	status    Status
	summary   string
	lines     []string
	perfdata  []Perfdata
	errs      []string
	verbosity int
	timeout   Timeout
}

// NewResult returns a Result with the status OK for the service name, for
// example "HTTP". The name is the first word of the output.
func NewResult(name string) *Result {
	return &Result{name: name, timeout: DefaultTimeout}
}

// SetVerbosity sets the verbosity level for [Result.Debugf].
// The level is the number of -v options.
func (r *Result) SetVerbosity(level int) {
	r.verbosity = level
}

// Raise changes the status of r to s if s has a higher severity.
// The severity order is OK, UNKNOWN, WARNING, CRITICAL.
func (r *Result) Raise(s Status) {
	r.status = Worst(r.status, s)
}

// Status returns the status of r.
func (r *Result) Status() Status {
	return r.status
}

// Summaryf sets the text of the first line after "NAME STATUS: ".
func (r *Result) Summaryf(format string, args ...any) {
	r.summary = fmt.Sprintf(format, args...)
}

// Linef adds one or more lines to the long output.
// A line break in the text starts a new line.
func (r *Result) Linef(format string, args ...any) {
	r.lines = append(r.lines, strings.Split(fmt.Sprintf(format, args...), "\n")...)
}

// Debugf adds lines to the long output if the verbosity level is level or
// higher. The monitoring-plugins guidelines give these levels:
//
//	1  additional information, for example the items that failed
//	2  configuration debug output, for example the command that ran
//	3  all detail for plugin problem diagnosis
func (r *Result) Debugf(level int, format string, args ...any) {
	if r.verbosity >= level {
		r.Linef(format, args...)
	}
}

// AddPerfdata adds p to the performance data.
// If p is not valid, AddPerfdata calls [Result.Fail] with Unknown, so
// that a plugin defect is visible in the output.
func (r *Result) AddPerfdata(p Perfdata) {
	if err := p.Validate(); err != nil {
		r.Fail(Unknown, err)
		return
	}
	r.perfdata = append(r.perfdata, p)
}

// Fail raises the status to s and puts the text of err in the summary.
// If err is or wraps [context.DeadlineExceeded], Fail uses the timeout
// status in place of s.
func (r *Result) Fail(s Status, err error) {
	if errors.Is(err, context.DeadlineExceeded) {
		s = r.timeout.State
		err = fmt.Errorf("timeout after %s: %w", r.timeout.Duration, err)
	}
	r.Raise(s)
	r.errs = append(r.errs, err.Error())
}

// String returns the plugin output with a line break at the end:
//
//	NAME STATUS: summary | perfdata
//
// or, with long output:
//
//	NAME STATUS: summary
//	long output line 1
//	long output line N | perfdata
//
// The text of Fail calls replaces the summary. String changes each "|" in
// the text to "/", because "|" starts the performance data.
func (r *Result) String() string {
	var b strings.Builder
	if r.name != "" {
		b.WriteString(r.name)
		b.WriteByte(' ')
	}
	b.WriteString(r.status.String())
	summary := r.summary
	if len(r.errs) > 0 {
		summary = strings.Join(r.errs, ", ")
	}
	if summary != "" {
		b.WriteString(": ")
		b.WriteString(cleanText(strings.ReplaceAll(summary, "\n", " ")))
	}

	perf := make([]string, len(r.perfdata))
	for i, p := range r.perfdata {
		perf[i] = p.String()
	}
	perfText := strings.Join(perf, " ")

	onFirst := r.PerfdataOnFirstLine || len(r.lines) == 0
	if onFirst && perfText != "" {
		b.WriteString(" | ")
		b.WriteString(perfText)
	}
	for i, line := range r.lines {
		b.WriteByte('\n')
		b.WriteString(cleanText(line))
		if !onFirst && i == len(r.lines)-1 && perfText != "" {
			b.WriteString(" | ")
			b.WriteString(perfText)
		}
	}
	b.WriteByte('\n')
	return b.String()
}

// Run calls check with a context that ends after t.Duration. Then Run
// writes the output of r to standard output and exits with the status of r.
// Run does not return.
//
// Special cases:
//
//	check panics                    exit with Unknown
//	check does not return in time   exit with t.State one second after t.Duration
func Run(r *Result, t Timeout, check func(ctx context.Context)) {
	out, status := execute(r, t, check)
	os.Stdout.WriteString(out)
	os.Exit(int(status))
}

func execute(r *Result, t Timeout, check func(ctx context.Context)) (string, Status) {
	r.timeout = t
	ctx, cancel := context.WithTimeout(context.Background(), t.Duration)
	defer cancel()

	done := make(chan struct{})
	go func() {
		defer close(done)
		defer func() {
			if v := recover(); v != nil {
				r.Fail(Unknown, fmt.Errorf("plugin panic: %v", v))
			}
		}()
		check(ctx)
	}()

	grace := time.NewTimer(t.Duration + time.Second)
	defer grace.Stop()
	select {
	case <-done:
		return r.String(), r.Status()
	case <-grace.C:
		// The check goroutine can still change r, so use a new Result.
		late := NewResult(r.name)
		late.Raise(t.State)
		late.Summaryf("check did not stop within the timeout of %s", t.Duration)
		return late.String(), late.Status()
	}
}

func cleanText(s string) string {
	return strings.ReplaceAll(strings.ReplaceAll(s, "|", "/"), "\r", "")
}
