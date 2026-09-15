package nagios

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"filippo.io/age"
	"filippo.io/age/armor"
)

// PassphraseEnv is the environment variable that [ResolvePassphrase] reads
// when no passphrase option is set.
const PassphraseEnv = "NAGIOS_PLUGIN_PASSPHRASE"

// DefaultWorkFactor is the scrypt work factor for [EncryptCredentials].
// A work factor of 15 uses 32 MiB of memory for each decryption. The age
// command uses 18, which uses 256 MiB for each decryption.
const DefaultWorkFactor = 15

// maxFileSize is the size limit of a credentials file and a passphrase file.
const maxFileSize = 1 << 20

// ResolvePassphrase returns the passphrase from the first source that is
// set, in this order:
//
//  1. The file at file, read with [ReadPassphraseFile]
//  2. The value of flag
//  3. The environment variable [PassphraseEnv]
func ResolvePassphrase(flag, file string) (string, error) {
	switch {
	case file != "":
		return ReadPassphraseFile(file)
	case flag != "":
		return flag, nil
	}
	if v := os.Getenv(PassphraseEnv); v != "" {
		return v, nil
	}
	return "", fmt.Errorf("no passphrase: set --passphrase-file, --passphrase, or %s", PassphraseEnv)
}

// ReadPassphraseFile returns the first line of the file at path, without
// the line break. The file must be a regular file, not a symbolic link.
// Group and others must have no permissions on the file.
func ReadPassphraseFile(path string) (string, error) {
	data, err := readRegularFile(path, true)
	if err != nil {
		return "", err
	}
	line, _, _ := strings.Cut(string(data), "\n")
	line = strings.TrimSuffix(line, "\r")
	if line == "" {
		return "", fmt.Errorf("passphrase file %s: first line is empty", path)
	}
	return line, nil
}

// LoadCredentials decrypts the age file at path with passphrase and
// returns the JSON object in it, for example
// {"username": "monitor", "password": "secret"}.
// The file can be binary or ASCII-armored. The file must be a regular
// file, not a symbolic link.
func LoadCredentials(path, passphrase string) (map[string]string, error) {
	data, err := readRegularFile(path, false)
	if err != nil {
		return nil, err
	}
	id, err := age.NewScryptIdentity(passphrase)
	if err != nil {
		return nil, fmt.Errorf("credentials file %s: %w", path, err)
	}
	var src io.Reader = bytes.NewReader(data)
	if trimmed := bytes.TrimLeft(data, " \t\r\n"); bytes.HasPrefix(trimmed, []byte(armor.Header)) {
		src = armor.NewReader(bytes.NewReader(trimmed))
	}
	plain, err := age.Decrypt(src, id)
	if err != nil {
		return nil, fmt.Errorf("credentials file %s: decryption failed, the passphrase is possibly not correct: %w", path, err)
	}
	body, err := io.ReadAll(io.LimitReader(plain, maxFileSize))
	if err != nil {
		return nil, fmt.Errorf("credentials file %s: %w", path, err)
	}
	var creds map[string]string
	if err := json.Unmarshal(body, &creds); err != nil {
		return nil, fmt.Errorf("credentials file %s: content is not a JSON object of strings: %w", path, err)
	}
	return creds, nil
}

// EncryptCredentials writes creds to w as an ASCII-armored age file,
// encrypted with passphrase and the scrypt work factor workFactor.
func EncryptCredentials(w io.Writer, creds map[string]string, passphrase string, workFactor int) error {
	rcpt, err := age.NewScryptRecipient(passphrase)
	if err != nil {
		return err
	}
	rcpt.SetWorkFactor(workFactor)
	aw := armor.NewWriter(w)
	ew, err := age.Encrypt(aw, rcpt)
	if err != nil {
		return err
	}
	if err := json.NewEncoder(ew).Encode(creds); err != nil {
		return err
	}
	if err := ew.Close(); err != nil {
		return err
	}
	return aw.Close()
}

// readRegularFile does not follow a symbolic link, so that a link cannot
// point the plugin to a different file. If private is true, it also
// rejects a file that group or others have permissions on.
func readRegularFile(path string, private bool) ([]byte, error) {
	before, err := os.Lstat(path)
	if err != nil {
		return nil, err
	}
	if !before.Mode().IsRegular() {
		return nil, fmt.Errorf("%s is not a regular file", path)
	}
	if private && before.Mode().Perm()&0o077 != 0 {
		return nil, fmt.Errorf("%s has mode %04o, set mode 0600 or 0400", path, before.Mode().Perm())
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	after, err := f.Stat()
	if err != nil {
		return nil, err
	}
	if !os.SameFile(before, after) {
		return nil, errors.New(path + " changed while the plugin opened it")
	}
	return io.ReadAll(io.LimitReader(f, maxFileSize))
}
