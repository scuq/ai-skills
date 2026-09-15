package main

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"check_http_example/internal/nagios"
)

func TestParseOptions(t *testing.T) {
	var out bytes.Buffer
	for _, args := range [][]string{{"-h"}, {"-V"}, {}, {"-u", "http://x", "-w", "20:10"}, {"-u", "http://x", "extra"}} {
		out.Reset()
		_, code, stop := parseOptions(args, &out)
		if !stop || code != int(nagios.Unknown) {
			t.Errorf("parseOptions(%q) = code %d, stop %v, want 3, true", args, code, stop)
		}
	}
	out.Reset()
	parseOptions([]string{"-h"}, &out)
	for i, line := range strings.Split(out.String(), "\n") {
		if len(line) > 80 {
			t.Errorf("help line %d has %d characters, the limit is 80: %q", i+1, len(line), line)
		}
	}

	o, _, stop := parseOptions([]string{"-u", "http://x", "-w", "1", "-c", "2", "-t", "5:UNKNOWN", "-vv", "-e", "200,301"}, &out)
	if stop || o.verbose != 2 || o.timeout.State != nagios.Unknown || !o.expect[301] {
		t.Errorf("parseOptions = %+v, stop %v", o, stop)
	}
}

func TestCheck(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		user, pass, ok := req.BasicAuth()
		if !ok || user != "monitor" || pass != "secret" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.Write([]byte("hello"))
	}))
	defer srv.Close()

	dir := t.TempDir()
	credPath := filepath.Join(dir, "creds.age")
	passPath := filepath.Join(dir, "pass")
	if err := os.WriteFile(passPath, []byte("pw\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	o, _, _ := parseOptions([]string{"--encrypt-credentials", credPath, "--passphrase-file", passPath, "--work-factor", "10"}, &out)
	if code := encryptCredentials(o, strings.NewReader(`{"username":"monitor","password":"secret"}`), &out); code != 0 {
		t.Fatalf("encryptCredentials = %d: %s", code, out.String())
	}

	o, _, stop := parseOptions([]string{"-u", srv.URL, "-w", "5", "-c", "10",
		"--credentials-file", credPath, "--passphrase-file", passPath}, &out)
	if stop {
		t.Fatalf("parseOptions stopped: %s", out.String())
	}
	r := nagios.NewResult(serviceName)
	check(context.Background(), r, o)

	got := r.String()
	re := regexp.MustCompile(`^HTTP OK: 200 OK, [0-9.]+ s, 5 B\nURL: .*\nProtocol: HTTP/1.1 \| time=[0-9.]+s;5;10;0 size=5B;;;0\n$`)
	if !re.MatchString(got) {
		t.Errorf("output =\n%s", got)
	}

	o.credentials = ""
	r = nagios.NewResult(serviceName)
	check(context.Background(), r, o)
	if r.Status() != nagios.Critical {
		t.Errorf("without credentials: status %v, output %q", r.Status(), r.String())
	}
}
