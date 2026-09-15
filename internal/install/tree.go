package install

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// BundleName is the name of the release asset with the skills and agents.
const BundleName = "ai-skills-bundle.tar.gz"

const (
	maxBundleEntries = 10000
	maxBundleBytes   = 256 << 20
)

// Extract writes the regular files and the directories of the
// gzip-compressed tar archive data to dir.
// It returns an error for an absolute path, a path outside dir, a link, or
// another special file.
func Extract(data []byte, dir string) error {
	gz, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("bundle: %w", err)
	}
	tr := tar.NewReader(gz)
	var total int64
	for n := 0; ; n++ {
		h, err := tr.Next()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return fmt.Errorf("bundle: %w", err)
		}
		if n >= maxBundleEntries {
			return fmt.Errorf("bundle: more than %d entries", maxBundleEntries)
		}
		name := filepath.FromSlash(h.Name)
		if !filepath.IsLocal(name) {
			return fmt.Errorf("bundle: path %q is outside the target directory", h.Name)
		}
		target := filepath.Join(dir, name)

		switch h.Typeflag {
		case tar.TypeXGlobalHeader:
			continue
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0o755); err != nil {
				return err
			}
		case tar.TypeReg:
			total += h.Size
			if total > maxBundleBytes {
				return fmt.Errorf("bundle: content is larger than %d bytes", maxBundleBytes)
			}
			if err := writeFile(target, tr, h); err != nil {
				return err
			}
		default:
			return fmt.Errorf("bundle: %q is not a regular file or a directory", h.Name)
		}
	}
}

func writeFile(target string, r io.Reader, h *tar.Header) error {
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return err
	}
	mode := os.FileMode(0o644)
	if h.Mode&0o111 != 0 {
		mode = 0o755
	}
	f, err := os.OpenFile(target, os.O_WRONLY|os.O_CREATE|os.O_EXCL, mode)
	if err != nil {
		return fmt.Errorf("bundle: %w", err)
	}
	if _, err := io.CopyN(f, r, h.Size); err != nil {
		f.Close()
		return fmt.Errorf("bundle: %s: %w", h.Name, err)
	}
	return f.Close()
}

// Items returns the items in root as slash-separated paths relative to
// root. An item is a directory in skills/ or a .md file in agents/.
func Items(root string) ([]string, error) {
	var items []string
	skills, err := os.ReadDir(filepath.Join(root, "skills"))
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return nil, err
	}
	for _, e := range skills {
		if e.IsDir() {
			items = append(items, "skills/"+e.Name())
		}
	}
	agents, err := os.ReadDir(filepath.Join(root, "agents"))
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return nil, err
	}
	for _, e := range agents {
		if e.Type().IsRegular() && strings.HasSuffix(e.Name(), ".md") {
			items = append(items, "agents/"+e.Name())
		}
	}
	sort.Strings(items)
	return items, nil
}

// validItem reports whether item has the form skills/NAME or agents/NAME.md,
// so that a changed manifest cannot point outside the Claude directory.
func validItem(item string) bool {
	kind, name, ok := strings.Cut(item, "/")
	if !ok || name == "" || name == "." || name == ".." || strings.ContainsAny(name, `/\`) {
		return false
	}
	return kind == "skills" || (kind == "agents" && strings.HasSuffix(name, ".md"))
}

// Hash returns a SHA-256 hash of the file or the directory tree at path.
// The hash covers the relative path, the type, the executable bit, and the
// content of each entry. The hash of a symbolic link covers its target.
// If path does not exist, the error wraps [fs.ErrNotExist].
func Hash(path string) (string, error) {
	h := sha256.New()
	err := filepath.WalkDir(path, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(path, p)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		switch {
		case d.Type()&fs.ModeSymlink != 0:
			target, err := os.Readlink(p)
			if err != nil {
				return err
			}
			fmt.Fprintf(h, "L %q %q\n", rel, target)
		case d.IsDir():
			fmt.Fprintf(h, "D %q\n", rel)
		case d.Type().IsRegular():
			info, err := d.Info()
			if err != nil {
				return err
			}
			sum, err := fileSum(p)
			if err != nil {
				return err
			}
			fmt.Fprintf(h, "F %q %t %s\n", rel, info.Mode()&0o111 != 0, sum)
		default:
			fmt.Fprintf(h, "O %q\n", rel)
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func fileSum(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// hashIfExists returns an empty hash when path does not exist.
func hashIfExists(path string) (string, error) {
	sum, err := Hash(path)
	if errors.Is(err, fs.ErrNotExist) {
		return "", nil
	}
	return sum, err
}
