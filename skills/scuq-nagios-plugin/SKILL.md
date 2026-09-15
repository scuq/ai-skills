---
name: scuq-nagios-plugin
description: >
  Rules and a Go template for monitoring plugins for Nagios Core, Naemon,
  Icinga 2, and other forks of Nagios. Load when you write, review, or
  change a check plugin: exit codes, output, long output, performance data,
  warning and critical ranges, timeouts, encrypted credentials, and a
  static Go build.
---

# scuq-nagios-plugin

This skill gives the rules for a monitoring plugin and a Go template that obeys them.
A plugin is a program that the monitoring core starts, reads from standard output, and judges by its exit code.

The rules come from these sources:

- The monitoring-plugins development guidelines (https://www.monitoring-plugins.org/doc/guidelines.html)
- The Nagios Core plugin API (https://assets.nagios.com/downloads/nagioscore/docs/nagioscore/4/en/pluginapi.html)
- The Naemon plugin API (https://www.naemon.io/documentation/usersguide/pluginapi.html)
- The Icinga 2 plugin API (https://icinga.com/docs/icinga-2/latest/doc/05-service-monitoring/)

When a source and this file disagree, this file decides.

## 1. Exit codes

| Code | Status | Use |
|---|---|---|
| 0 | OK | The check ran, and the service works. |
| 1 | WARNING | The check ran, and a value is outside the warning range, or the service shows a problem. |
| 2 | CRITICAL | The check ran, and the service does not work, or a value is outside the critical range. |
| 3 | UNKNOWN | The check did not run: options are not correct, or an internal failure stopped it. |

- Exit with 3 after `--help` and `--version`.
- If the plugin cannot connect to the service, exit with 2, not 3.
- If the plugin cannot parse its options, cannot read its credentials, or has a defect, exit with 3.
- To combine statuses, use the severity order OK, UNKNOWN, WARNING, CRITICAL. This is `max_state_alt` in the monitoring-plugins project.

## 2. Output

Write only to standard output.
The monitoring core does not read standard error.

The output has this shape:

```text
NAME STATUS: summary | perfdata
long output line 1
long output line N | more perfdata
more perfdata
```

- Line 1 starts with the service name and the status, for example `HTTP OK: 200 OK, 0.042 s`.
- Keep line 1 shorter than 80 characters.
  The web interfaces show line 1 in the service list.
- Put details in the long output, one fact per line.
- If there is long output, write the performance data after a `|` at the end of the last long output line.
  Naemon and Nagios Core add it to the performance data of line 1.
- If there is no long output, write the performance data after a `|` on line 1.
- Never write a `|` in the text.
  The template changes `|` in the text to `/`.
- Write numbers in the C locale: a period as the decimal separator, no thousands separator.
- Nagios Core reads only the first 4 KB of output.
  Naemon has no limit.
  Keep the output under 4 KB.
- Never write a password, a passphrase, a token, or the content of a credentials file in the output.

## 3. Performance data

The format of one metric:

```text
'label'=value[UOM];[warn];[crit];[min];[max]
```

One space separates two metrics.

- The label cannot contain `=`, `'`, `|`, or a line break.
  Use `nagios.SanitizeLabel` for a label from external data, for example a mount point.
- Put single quotes around a label with a space or a special character.
  The template does this.
- Make the first 19 characters of each label unique.
  RRD-based graphing tools use only 19 characters.
- Value, min, and max use only the characters `-0123456789.`.
- If the plugin cannot get a value, write `U` as the value.
- Value, min, max, warn, and crit use the same unit.
- Warn and crit use the range format of section 4.
- Leave a field empty if it does not apply.
  Drop the empty fields at the end.
- For `%`, min and max are not necessary.
- Use one unit from this list:

| UOM | Meaning |
|---|---|
| (none) | A number of things, for example users, processes, or a load average |
| `s`, `ms`, `us` | Seconds, milliseconds, microseconds |
| `%` | Percent |
| `B`, `KB`, `MB`, `GB`, `TB` | Bytes |
| `c` | A continuous counter, for example bytes sent on an interface |

Icinga 2 accepts more units, for example `packets`, `W`, and `C`.
Icinga 2 discards a unit that it does not know.
Nagios Core and Naemon do not check the unit.
To use a unit from the Icinga 2 list, add it to `nagios.UOMs` in `main`.

## 4. Threshold ranges

The range format is `[@]start:end`.
The ends are inclusive.

| Range | Alert if x is |
|---|---|
| `10` | < 0 or > 10 |
| `10:` | < 10 |
| `~:10` | > 10 |
| `10:20` | < 10 or > 20 |
| `@10:20` | ≥ 10 and ≤ 20 |

- `~` is negative infinity.
- Without `start:`, start is 0.
- Without end after the colon, end is infinity.
- If start is greater than end, exit with 3 and name the range.
- Evaluate the critical range first, then the warning range.
- Write the ranges from the options into the warn and crit fields of the metric that they apply to.

## 5. Options

Use these options for the same purpose in every plugin:

| Option | Purpose |
|---|---|
| `-h`, `--help` | Show help and exit with 3. |
| `-V`, `--version` | Show the version and exit with 3. |
| `-v`, `--verbose` | Add detail. The option can repeat, up to `-vvv`. |
| `-t`, `--timeout` | Time limit as `SECONDS[:STATE]`. The default is `10`, with the status CRITICAL. |
| `-w`, `--warning` | Warning range. |
| `-c`, `--critical` | Critical range. |
| `-H`, `--hostname` | Host name or address. |
| `-p`, `--port` | Port. |
| `-u`, `--url` or `--username` | URL or user name. |
| `-l`, `--logname` | Login name. |
| `-C`, `--community` | SNMP community. |

- Do not use these letters for other purposes.
- Show help that fits in 80 columns.
- If an option is not correct, write `NAME UNKNOWN: <reason>` and a usage line, and exit with 3.
- For credentials, use `--credentials-file` and the passphrase options of section 7.
  Do not add a password option.

The verbosity levels:

| Level | Content |
|---|---|
| 0 | One line, the summary |
| 1 | Additional information, for example the items that failed |
| 2 | Configuration debug output, for example the request or the command |
| 3 | All detail for problem diagnosis |

## 6. Timeouts

- Give the context from `nagios.Run` to every network call, command, and query.
- Do not start work that ignores the context.
- If the context ends, call `r.Fail` with the error.
  `r.Fail` uses the status of the `-t` option.
- If the check does not stop one second after the time limit, `nagios.Run` writes the timeout status and exits.

## 7. Credentials

The plugin reads credentials from an encrypted file.
The operator gives the file path and the passphrase.

- The file is an age file (https://age-encryption.org) with a passphrase recipient.
  It can be binary or ASCII-armored.
- The decrypted content is a JSON object of strings, for example `{"username": "monitor", "password": "secret"}`.
- The plugin gets the passphrase from the first source that has a value:
  1. `--passphrase-file PATH`: The first line of the file. Group and others must have no permissions on the file.
  2. `--passphrase TEXT`: The option value.
  3. `NAGIOS_PLUGIN_PASSPHRASE`: The environment variable.
- The plugin does not follow a symbolic link to the credentials file or the passphrase file.
- If the plugin cannot read or decrypt the credentials, it exits with 3.

CAUTION: Do not use `--passphrase` on a shared host. All local users can read the options of a process with `ps`.

CAUTION: Do not encrypt a credentials file with `age -p` for a check that runs often. It uses the scrypt work factor 18, and each decryption then uses 256 MiB of memory. Use `--encrypt-credentials`. Its default work factor 15 uses 32 MiB.

To make a credentials file with the template:

```sh
umask 077
printf '%s\n' 'the passphrase' > /etc/naemon/credentials/web.pass
printf '%s' '{"username":"monitor","password":"secret"}' \
  | check_http_example --encrypt-credentials /etc/naemon/credentials/web.age \
      --passphrase-file /etc/naemon/credentials/web.pass
```

## 8. Security

- Prefer a Go library to an external command.
- If the plugin must run a command, use `exec.CommandContext` with an absolute path.
  Never start a shell.
- Never put user input in a command line, a query, or a path without validation.
- Never follow a symbolic link to a file that the plugin reads.
- Limit the size of each response that the plugin reads.

## 9. Static build

The plugin is one static binary.
It needs no C library, no interpreter, and no shared library at run time.

- Build with `CGO_ENABLED=0`.
- Do not add a dependency that needs cgo, for example `github.com/mattn/go-sqlite3`.
  Use a pure Go package, for example `modernc.org/sqlite`.
- Build with `-trimpath` and `-ldflags "-s -w -X main.version=VERSION"`.
- Build for `linux/amd64` and `linux/arm64`.
- Make sure that `go version -m BINARY` shows `CGO_ENABLED=0`.
- With `CGO_ENABLED=0`, Go uses its own DNS resolver.
  It reads `/etc/resolv.conf` and `/etc/hosts`, but not NSS modules.

## 10. The template

`reference/template/` is a complete plugin, `check_http_example`.
It checks a URL, reads credentials, and uses all rules in this file.

| File | Content |
|---|---|
| `go.mod` | The module `check_http_example`, with age and pflag |
| `main.go` | Options, the check function, and `--encrypt-credentials` |
| `main_test.go` | Tests for options, a check against a test server, and credentials |
| `internal/nagios/status.go` | `Status`, `ParseStatus`, `Worst` |
| `internal/nagios/range.go` | `Range`, `ParseRange`, `Thresholds` |
| `internal/nagios/perfdata.go` | `Perfdata`, `UOMs`, `SanitizeLabel` |
| `internal/nagios/result.go` | `Result`, `Timeout`, `Run` |
| `internal/nagios/credentials.go` | `LoadCredentials`, `EncryptCredentials`, `ResolvePassphrase` |
| `build.sh` | Vet, test, static build, static test |

To start a new plugin:

1. Copy `reference/template/` to the new plugin directory.
2. Change the module name in `go.mod` to the plugin name, for example `check_postgres`.
3. Change the import path `check_http_example/internal/nagios` in `main.go` and `main_test.go`.
4. Change `progName` and `serviceName` in `main.go`.
5. Replace the options, the `check` function, and the tests.
6. Keep `internal/nagios/` without change, unless a rule in this file is missing from it.
7. Run `go mod tidy`.
8. Run `bash build.sh`.

Use the `internal/nagios` API like this:

```go
r := nagios.NewResult("PGSQL")
nagios.Run(r, o.timeout, func(ctx context.Context) {
	n, err := countConnections(ctx, o)
	if err != nil {
		r.Fail(nagios.Critical, err)
		return
	}
	r.Raise(o.thresholds.Evaluate(float64(n)))
	r.Summaryf("%d connections", n)
	r.AddPerfdata(nagios.Perfdata{
		Label:    "connections",
		Value:    float64(n),
		Warning:  o.thresholds.Warning,
		Critical: o.thresholds.Critical,
		Min:      nagios.Float(0),
	})
})
```

## 11. Tests

Each plugin has tests for:

- Each option error, with exit code 3
- `--help` and `--version`, with exit code 3
- Each status that the check can give: OK, WARNING, CRITICAL, and UNKNOWN
- The full output text of one check, with a regular expression for the variable values
- The performance data: labels, units, ranges, min, and max
- The timeout: a check against a server that does not answer
- The credentials: a correct passphrase, an incorrect passphrase, and a missing file

Use `net/http/httptest`, `net.Listen` on `127.0.0.1:0`, or a fake of the client interface.
Tests do not need a network or a real service.

## 12. Documentation

Each plugin has a `README.md` in the man-page structure of the scraibe standard.
Load the scuq-scraibe skill before you write it.

The README gives:

- All options, with the default values
- The exit codes and the conditions for each
- The performance data: each label, its unit, and the range options that apply to it
- A Naemon command definition
- The steps to make the credentials file, if the plugin uses credentials

A Naemon command definition example:

```text
define command {
    command_name  check_http_example
    command_line  $USER1$/check_http_example -u $ARG1$ -w $ARG2$ -c $ARG3$ --credentials-file /etc/naemon/credentials/web.age --passphrase-file /etc/naemon/credentials/web.pass
}
```

## 13. Review checklist

Check every item before you finish.

- `gofmt -l .` shows no file.
- `go vet ./...` passes.
- `go test ./...` passes.
- `bash build.sh` passes, and each binary shows `CGO_ENABLED=0`.
- `--help` and `--version` exit with 3.
- An option error writes `NAME UNKNOWN:` on line 1 and exits with 3.
- Line 1 is shorter than 80 characters for the usual case.
- No text contains `|`.
- Each metric has a valid label and a unit from `nagios.UOMs`.
- Each metric with a range option has the range in its warn or crit field.
- Every network call, command, and query uses the context.
- No output contains a secret.
- The README lists every option, exit code, and metric.
