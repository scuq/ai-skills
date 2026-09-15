// check_http_example checks an HTTP or HTTPS URL. It reports the status
// code, the response time, and the response size.
//
// This program is the template for a plugin. Copy the template directory,
// change the module name in go.mod, and replace the check function.
//
// Usage:
//
//	check_http_example -u URL [-w SECONDS] [-c SECONDS] [-e CODES] [-k]
//	    [-t SECONDS[:STATE]] [-v]...
//	    [--credentials-file PATH [--passphrase-file PATH | --passphrase TEXT]]
//	check_http_example --encrypt-credentials PATH [--passphrase-file PATH]
//	    [--work-factor N] < credentials.json
//	check_http_example -h | -V
package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	flag "github.com/spf13/pflag"

	"check_http_example/internal/nagios"
)

// version is set at build time with -ldflags "-X main.version=VERSION".
var version = "0.0.0"

const (
	progName    = "check_http_example"
	serviceName = "HTTP"
	maxBodySize = 16 << 20
)

type options struct {
	url            string
	thresholds     nagios.Thresholds
	expect         map[int]bool
	insecure       bool
	timeout        nagios.Timeout
	verbose        int
	credentials    string
	passphrase     string
	passphraseFile string
	encryptTo      string
	workFactor     int
}

func main() {
	o, code, stop := parseOptions(os.Args[1:], os.Stdout)
	if stop {
		os.Exit(code)
	}
	if o.encryptTo != "" {
		os.Exit(encryptCredentials(o, os.Stdin, os.Stdout))
	}
	r := nagios.NewResult(serviceName)
	r.SetVerbosity(o.verbose)
	nagios.Run(r, o.timeout, func(ctx context.Context) {
		check(ctx, r, o)
	})
}

// parseOptions returns the options in args. If stop is true, the caller
// must exit with code, because parseOptions wrote help, the version, or
// an error to out.
func parseOptions(args []string, out io.Writer) (o options, code int, stop bool) {
	fs := flag.NewFlagSet(progName, flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	fs.SortFlags = false

	var warning, critical, expect, timeout string
	var help, showVersion bool
	// The monitoring-plugins guidelines limit help to 80 columns. pflag
	// pads the option column, so keep each text under 30 characters.
	fs.StringVarP(&o.url, "url", "u", "", "URL to check")
	fs.StringVarP(&warning, "warning", "w", "", "WARNING range, seconds")
	fs.StringVarP(&critical, "critical", "c", "", "CRITICAL range, seconds")
	fs.StringVarP(&expect, "expect", "e", "200", "status codes for OK")
	fs.BoolVarP(&o.insecure, "insecure", "k", false, "do not verify TLS certificate")
	fs.StringVarP(&timeout, "timeout", "t", "10", "limit as SECONDS[:STATE]")
	fs.CountVarP(&o.verbose, "verbose", "v", "more detail, up to -vvv")
	fs.StringVar(&o.credentials, "credentials-file", "", "age file with credentials")
	fs.StringVar(&o.passphraseFile, "passphrase-file", "", "passphrase file, mode 0600")
	fs.StringVar(&o.passphrase, "passphrase", "", "passphrase, visible in ps")
	fs.StringVar(&o.encryptTo, "encrypt-credentials", "", "encrypt stdin JSON to file")
	fs.IntVar(&o.workFactor, "work-factor", nagios.DefaultWorkFactor, "scrypt work factor")
	fs.BoolVarP(&help, "help", "h", false, "show this help and exit")
	fs.BoolVarP(&showVersion, "version", "V", false, "show the version and exit")

	fail := func(err error) (options, int, bool) {
		fmt.Fprintf(out, "%s UNKNOWN: %v\nUsage: %s -u URL [options], see %s --help\n",
			serviceName, err, progName, progName)
		return o, int(nagios.Unknown), true
	}

	if err := fs.Parse(args); err != nil {
		return fail(err)
	}
	switch {
	case help:
		fmt.Fprintf(out, "%s %s\n\nCheck an HTTP or HTTPS URL.\n\nOptions:\n%s", progName, version, fs.FlagUsages())
		return o, int(nagios.Unknown), true
	case showVersion:
		fmt.Fprintf(out, "%s %s\n", progName, version)
		return o, int(nagios.Unknown), true
	case fs.NArg() > 0:
		return fail(fmt.Errorf("argument %q is not an option", fs.Arg(0)))
	case o.encryptTo != "":
		return o, 0, false
	case o.url == "":
		return fail(errors.New("option -u is necessary"))
	}

	var err error
	if o.thresholds, err = nagios.ParseThresholds(warning, critical); err != nil {
		return fail(err)
	}
	if o.timeout, err = nagios.ParseTimeout(timeout); err != nil {
		return fail(err)
	}
	o.expect = map[int]bool{}
	for _, s := range strings.Split(expect, ",") {
		n, err := strconv.Atoi(strings.TrimSpace(s))
		if err != nil || n < 100 || n > 599 {
			return fail(fmt.Errorf("status code %q is not a number from 100 to 599", s))
		}
		o.expect[n] = true
	}
	return o, 0, false
}

// check does one HTTP request and records the status, the text, and the
// performance data in r.
func check(ctx context.Context, r *nagios.Result, o options) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, o.url, nil)
	if err != nil {
		r.Fail(nagios.Unknown, err)
		return
	}
	if o.credentials != "" {
		passphrase, err := nagios.ResolvePassphrase(o.passphrase, o.passphraseFile)
		if err != nil {
			r.Fail(nagios.Unknown, err)
			return
		}
		creds, err := nagios.LoadCredentials(o.credentials, passphrase)
		if err != nil {
			r.Fail(nagios.Unknown, err)
			return
		}
		if creds["username"] == "" {
			r.Fail(nagios.Unknown, fmt.Errorf("credentials file %s has no username", o.credentials))
			return
		}
		req.SetBasicAuth(creds["username"], creds["password"])
	}
	r.Debugf(2, "GET %s", o.url)

	client := &http.Client{Transport: &http.Transport{
		Proxy:           http.ProxyFromEnvironment,
		TLSClientConfig: &tls.Config{InsecureSkipVerify: o.insecure},
	}}
	start := time.Now()
	resp, err := client.Do(req)
	if err != nil {
		r.Fail(nagios.Critical, err)
		return
	}
	defer resp.Body.Close()
	size, err := io.Copy(io.Discard, io.LimitReader(resp.Body, maxBodySize))
	elapsed := time.Since(start).Seconds()
	if err != nil {
		r.Fail(nagios.Critical, fmt.Errorf("read response body: %w", err))
		return
	}

	r.Raise(o.thresholds.Evaluate(elapsed))
	codeText := resp.Status
	if !o.expect[resp.StatusCode] {
		r.Raise(nagios.Critical)
		codeText += " (unexpected)"
	}
	r.Summaryf("%s, %.3f s, %d B", codeText, elapsed, size)
	r.Linef("URL: %s", o.url)
	r.Linef("Protocol: %s", resp.Proto)
	r.Debugf(1, "Server: %s", resp.Header.Get("Server"))

	r.AddPerfdata(nagios.Perfdata{
		Label:    "time",
		Value:    elapsed,
		UOM:      "s",
		Warning:  o.thresholds.Warning,
		Critical: o.thresholds.Critical,
		Min:      nagios.Float(0),
	})
	r.AddPerfdata(nagios.Perfdata{
		Label: "size",
		Value: float64(size),
		UOM:   "B",
		Min:   nagios.Float(0),
	})
}

// encryptCredentials reads a JSON object from in and writes it encrypted
// to the file o.encryptTo with mode 0600. It returns the exit code.
func encryptCredentials(o options, in io.Reader, out io.Writer) int {
	var creds map[string]string
	if err := json.NewDecoder(in).Decode(&creds); err != nil {
		fmt.Fprintf(out, "Standard input is not a JSON object of strings: %v\n", err)
		return 1
	}
	passphrase, err := nagios.ResolvePassphrase(o.passphrase, o.passphraseFile)
	if err != nil {
		fmt.Fprintln(out, err)
		return 1
	}
	f, err := os.OpenFile(o.encryptTo, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		fmt.Fprintf(out, "Cannot create %s: %v. Remove the file or give a new path.\n", o.encryptTo, err)
		return 1
	}
	if err := nagios.EncryptCredentials(f, creds, passphrase, o.workFactor); err != nil {
		f.Close()
		os.Remove(o.encryptTo)
		fmt.Fprintf(out, "Encryption failed: %v\n", err)
		return 1
	}
	if err := f.Close(); err != nil {
		fmt.Fprintf(out, "Cannot write %s: %v\n", o.encryptTo, err)
		return 1
	}
	fmt.Fprintf(out, "Wrote %s\n", o.encryptTo)
	return 0
}
