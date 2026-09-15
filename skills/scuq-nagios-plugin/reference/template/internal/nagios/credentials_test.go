package nagios

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCredentialsRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "creds.age")
	want := map[string]string{"username": "monitor", "password": "p;a|s=s'"}

	var buf bytes.Buffer
	if err := EncryptCredentials(&buf, want, "correct horse", 10); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, buf.Bytes(), 0o600); err != nil {
		t.Fatal(err)
	}

	got, err := LoadCredentials(path, "correct horse")
	if err != nil {
		t.Fatal(err)
	}
	if got["username"] != want["username"] || got["password"] != want["password"] {
		t.Errorf("LoadCredentials = %v, want %v", got, want)
	}

	if _, err := LoadCredentials(path, "wrong"); err == nil || !strings.Contains(err.Error(), "decryption failed") {
		t.Errorf("wrong passphrase: err = %v", err)
	}

	link := filepath.Join(dir, "link.age")
	if err := os.Symlink(path, link); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadCredentials(link, "correct horse"); err == nil {
		t.Error("symbolic link: no error, want an error")
	}
}

func TestReadPassphraseFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "pass")
	if err := os.WriteFile(path, []byte("secret\r\nsecond line\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	got, err := ReadPassphraseFile(path)
	if err != nil || got != "secret" {
		t.Errorf("ReadPassphraseFile = %q, %v", got, err)
	}

	if err := os.Chmod(path, 0o640); err != nil {
		t.Fatal(err)
	}
	if _, err := ReadPassphraseFile(path); err == nil {
		t.Error("mode 0640: no error, want an error")
	}
}

func TestResolvePassphrase(t *testing.T) {
	t.Setenv(PassphraseEnv, "from-env")
	if got, _ := ResolvePassphrase("from-flag", ""); got != "from-flag" {
		t.Errorf("flag: got %q", got)
	}
	if got, _ := ResolvePassphrase("", ""); got != "from-env" {
		t.Errorf("env: got %q", got)
	}
	t.Setenv(PassphraseEnv, "")
	if _, err := ResolvePassphrase("", ""); err == nil {
		t.Error("no source: no error, want an error")
	}
}
