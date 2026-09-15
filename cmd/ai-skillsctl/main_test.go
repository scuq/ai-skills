package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/scuq/ai-skills/internal/install"
)

// fakeGitHub serves one release with a bundle of files and a SHA256SUMS asset.
func fakeGitHub(t *testing.T, tag string, files map[string]string, damage bool) *httptest.Server {
	t.Helper()
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	for name, content := range files {
		tw.WriteHeader(&tar.Header{Name: name, Mode: 0o644, Size: int64(len(content)), Typeflag: tar.TypeReg})
		tw.Write([]byte(content))
	}
	tw.Close()
	gz.Close()
	bundle := buf.Bytes()
	sum := sha256.Sum256(bundle)
	if damage {
		sum[0] ^= 0xff
	}
	sums := fmt.Sprintf("%s  %s\n", hex.EncodeToString(sum[:]), install.BundleName)

	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/repos/o/r/releases/latest", "/repos/o/r/releases/tags/" + tag:
			fmt.Fprintf(w, `{"tag_name":%q,"assets":[{"name":%q,"browser_download_url":"%s/dl/bundle"},{"name":"SHA256SUMS","browser_download_url":"%s/dl/sums"}]}`,
				tag, install.BundleName, srv.URL, srv.URL)
		case "/dl/bundle":
			w.Write(bundle)
		case "/dl/sums":
			fmt.Fprint(w, sums)
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)
	return srv
}

func runCmd(t *testing.T, args ...string) (int, string, string) {
	t.Helper()
	var stdout, stderr bytes.Buffer
	code := run(context.Background(), args, &stdout, &stderr)
	return code, stdout.String(), stderr.String()
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func TestUpdateFlow(t *testing.T) {
	claude := t.TempDir()
	os.MkdirAll(filepath.Join(claude, "skills/mine"), 0o755)
	os.WriteFile(filepath.Join(claude, "skills/mine/SKILL.md"), []byte("mine"), 0o644)
	os.MkdirAll(filepath.Join(claude, "agents"), 0o755)
	os.WriteFile(filepath.Join(claude, "agents/alpha.md"), []byte("agent"), 0o644)

	v1 := fakeGitHub(t, "v0.1.0", map[string]string{
		"VERSION":                  "0.1.0\n",
		"skills/alpha/SKILL.md":    "alpha 1",
		"skills/alpha/ref/x.txt":   "x",
		"skills/beta/SKILL.md":     "beta",
		"agents/alpha.md":          "agent",
		"skills/alpha/ref/y/z.txt": "z",
	}, false)
	common := []string{"--claude-dir", claude, "--repo", "o/r", "--api-url", v1.URL}

	code, out, errOut := runCmd(t, append([]string{"update"}, common...)...)
	if code != 0 {
		t.Fatalf("update v0.1.0: exit %d\n%s%s", code, out, errOut)
	}
	for _, want := range []string{"none -> v0.1.0", "adopt    agents/alpha.md", "install  skills/alpha", "install  skills/beta", "Installed ai-skills v0.1.0."} {
		if !strings.Contains(out, want) {
			t.Errorf("update output has no %q:\n%s", want, out)
		}
	}
	if readFile(t, filepath.Join(claude, "skills/alpha/ref/y/z.txt")) != "z" || readFile(t, filepath.Join(claude, "skills/mine/SKILL.md")) != "mine" {
		t.Error("files after update are not correct")
	}
	if matches, _ := filepath.Glob(filepath.Join(claude, ".ai-skills-tmp-*")); len(matches) > 0 {
		t.Errorf("staging directory is still there: %v", matches)
	}

	code, out, _ = runCmd(t, append([]string{"update"}, common...)...)
	if code != 0 || !strings.Contains(out, "v0.1.0 is up to date") {
		t.Errorf("second update: exit %d\n%s", code, out)
	}

	code, out, _ = runCmd(t, append([]string{"status"}, common...)...)
	if code != 0 || !strings.Contains(out, "Installed: v0.1.0") || !strings.Contains(out, "ok       skills/alpha") || !strings.Contains(out, "Latest release: v0.1.0") {
		t.Errorf("status: exit %d\n%s", code, out)
	}

	// Release v0.2.0 changes alpha and removes beta. alpha has a local change.
	v2 := fakeGitHub(t, "v0.2.0", map[string]string{
		"VERSION":               "0.2.0\n",
		"skills/alpha/SKILL.md": "alpha 2",
		"agents/alpha.md":       "agent",
	}, false)
	common[len(common)-1] = v2.URL
	os.WriteFile(filepath.Join(claude, "skills/alpha/SKILL.md"), []byte("local"), 0o644)

	code, out, _ = runCmd(t, append([]string{"status", "--offline"}, common...)...)
	if code != 0 || !strings.Contains(out, "modified skills/alpha") || strings.Contains(out, "Latest") {
		t.Errorf("status --offline: exit %d\n%s", code, out)
	}

	code, out, _ = runCmd(t, append([]string{"update", "--dry-run"}, common...)...)
	if code != 0 || !strings.Contains(out, "update   skills/alpha (local changes)") || !strings.Contains(out, "remove   skills/beta") || !strings.Contains(out, "Dry run") {
		t.Errorf("update --dry-run: exit %d\n%s", code, out)
	}

	code, _, errOut = runCmd(t, append([]string{"update"}, common...)...)
	if code != 1 || !strings.Contains(errOut, "--force") || readFile(t, filepath.Join(claude, "skills/alpha/SKILL.md")) != "local" {
		t.Errorf("update with a conflict: exit %d\n%s", code, errOut)
	}

	code, out, errOut = runCmd(t, append([]string{"update", "--force"}, common...)...)
	if code != 0 {
		t.Fatalf("update --force: exit %d\n%s%s", code, out, errOut)
	}
	if readFile(t, filepath.Join(claude, "skills/alpha/SKILL.md")) != "alpha 2" {
		t.Error("alpha is not version 2")
	}
	if _, err := os.Stat(filepath.Join(claude, "skills/beta")); !os.IsNotExist(err) {
		t.Error("beta still exists")
	}
	backups, _ := filepath.Glob(filepath.Join(claude, install.BackupDirName, "*", "skills/alpha/SKILL.md"))
	if len(backups) != 1 || readFile(t, backups[0]) != "local" {
		t.Errorf("backup = %v", backups)
	}
}

func TestUpdateRejectsDamagedBundle(t *testing.T) {
	srv := fakeGitHub(t, "v0.1.0", map[string]string{"skills/a/SKILL.md": "a"}, true)
	claude := t.TempDir()
	code, _, errOut := runCmd(t, "update", "--claude-dir", claude, "--repo", "o/r", "--api-url", srv.URL)
	if code != 1 || !strings.Contains(errOut, "SHA-256") {
		t.Errorf("exit %d\n%s", code, errOut)
	}
	if _, err := os.Stat(filepath.Join(claude, "skills")); !os.IsNotExist(err) {
		t.Error("update with a damaged bundle wrote files")
	}
}

func TestUsage(t *testing.T) {
	tests := []struct {
		args []string
		code int
	}{
		{nil, exitUsage},
		{[]string{"frobnicate"}, exitUsage},
		{[]string{"update", "extra"}, exitUsage},
		{[]string{"update", "--bogus"}, exitUsage},
		{[]string{"version"}, exitOK},
		{[]string{"--help"}, exitOK},
		{[]string{"status", "-h"}, exitOK},
	}
	for _, tt := range tests {
		if code, _, _ := runCmd(t, tt.args...); code != tt.code {
			t.Errorf("run(%q) = %d, want %d", tt.args, code, tt.code)
		}
	}
	for i, line := range strings.Split(usage, "\n") {
		if len(line) > 80 {
			t.Errorf("usage line %d is longer than 80 characters", i+1)
		}
	}
}

func TestReplaceExecutable(t *testing.T) {
	path := filepath.Join(t.TempDir(), "ai-skillsctl")
	os.WriteFile(path, []byte("old"), 0o755)
	if err := replaceExecutable(path, []byte("new")); err != nil {
		t.Fatal(err)
	}
	info, _ := os.Stat(path)
	if readFile(t, path) != "new" || info.Mode().Perm() != 0o755 {
		t.Errorf("content %q, mode %v", readFile(t, path), info.Mode())
	}
	if entries, _ := os.ReadDir(filepath.Dir(path)); len(entries) != 1 {
		t.Errorf("directory has %d entries, want 1", len(entries))
	}
}
