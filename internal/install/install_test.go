package install

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// tarGz returns a gzip-compressed tar archive with the files in files.
func tarGz(t *testing.T, files map[string]string) []byte {
	t.Helper()
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	for name, content := range files {
		if err := tw.WriteHeader(&tar.Header{Name: name, Mode: 0o644, Size: int64(len(content)), Typeflag: tar.TypeReg}); err != nil {
			t.Fatal(err)
		}
		tw.Write([]byte(content))
	}
	tw.Close()
	gz.Close()
	return buf.Bytes()
}

func writeFiles(t *testing.T, root string, files map[string]string) {
	t.Helper()
	for name, content := range files {
		path := filepath.Join(root, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func stage(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	if err := Extract(tarGz(t, files), dir); err != nil {
		t.Fatal(err)
	}
	return dir
}

func TestExtractRejectsUnsafeEntries(t *testing.T) {
	for _, name := range []string{"../evil", "/etc/evil", "skills/../../evil"} {
		if err := Extract(tarGz(t, map[string]string{name: "x"}), t.TempDir()); err == nil {
			t.Errorf("Extract(%q): no error", name)
		}
	}

	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	tw.WriteHeader(&tar.Header{Name: "skills/link", Linkname: "/etc/passwd", Typeflag: tar.TypeSymlink})
	tw.Close()
	gz.Close()
	if err := Extract(buf.Bytes(), t.TempDir()); err == nil {
		t.Error("Extract(symlink): no error")
	}
}

func TestItemsAndHash(t *testing.T) {
	root := t.TempDir()
	writeFiles(t, root, map[string]string{
		"skills/a/SKILL.md":   "a",
		"skills/a/ref/x.txt":  "x",
		"skills/README.md":    "not an item",
		"agents/a.md":         "agent",
		"agents/notes.txt":    "not an item",
		"VERSION":             "1.0.0",
		"skills/b/SKILL.md":   "b",
		"agents/sub/other.md": "not an item",
	})
	items, err := Items(root)
	if err != nil {
		t.Fatal(err)
	}
	if got := strings.Join(items, ","); got != "agents/a.md,skills/a,skills/b" {
		t.Errorf("Items = %s", got)
	}

	path := filepath.Join(root, "skills/a")
	h1, _ := Hash(path)
	os.Chmod(filepath.Join(path, "ref/x.txt"), 0o755)
	h2, _ := Hash(path)
	os.WriteFile(filepath.Join(path, "ref/x.txt"), []byte("y"), 0o755)
	h3, _ := Hash(path)
	if h1 == h2 || h2 == h3 {
		t.Errorf("Hash does not change with mode or content: %s %s %s", h1, h2, h3)
	}
	if _, err := hashIfExists(filepath.Join(root, "missing")); err != nil {
		t.Errorf("hashIfExists(missing) error = %v", err)
	}
}

func TestPlanAndApply(t *testing.T) {
	claude := t.TempDir()
	writeFiles(t, claude, map[string]string{
		"skills/mine/SKILL.md":  "mine",
		"agents/same.md":        "same",
		"skills/other/SKILL.md": "local other",
	})

	// First installation into a directory without a manifest.
	v1 := stage(t, map[string]string{
		"skills/a/SKILL.md":     "a1",
		"skills/gone/SKILL.md":  "gone",
		"agents/same.md":        "same",
		"skills/other/SKILL.md": "release other",
	})
	m, _ := LoadManifest(claude)
	p, err := NewPlan(claude, v1, "v1.0.0", m)
	if err != nil {
		t.Fatal(err)
	}
	want := "agents/same.md adopt,skills/a install,skills/gone install,skills/other replace!"
	if got := describe(p.Changes); got != want {
		t.Errorf("plan v1 = %s, want %s", got, want)
	}
	if _, err := p.Apply(claude, false, time.Now()); err == nil {
		t.Fatal("Apply with a conflict and without force: no error")
	}
	if readFile(t, filepath.Join(claude, "skills/other/SKILL.md")) != "local other" {
		t.Fatal("Apply without force changed a file")
	}
	now := time.Date(2026, 9, 15, 12, 0, 0, 0, time.UTC)
	backup, err := p.Apply(claude, true, now)
	if err != nil {
		t.Fatal(err)
	}
	if got := readFile(t, filepath.Join(backup, "skills/other/SKILL.md")); got != "local other" {
		t.Errorf("backup = %q", got)
	}
	if got := readFile(t, filepath.Join(claude, "skills/other/SKILL.md")); got != "release other" {
		t.Errorf("installed other = %q", got)
	}
	if readFile(t, filepath.Join(claude, "skills/mine/SKILL.md")) != "mine" {
		t.Error("Apply changed an item that is not in the release")
	}

	// Second release: a changes, gone goes away, and a has a local change.
	m, err = LoadManifest(claude)
	if err != nil || m.Version != "v1.0.0" || len(m.Items) != 4 || !m.Updated.Equal(now) {
		t.Fatalf("manifest = %+v, %v", m, err)
	}
	v2 := stage(t, map[string]string{
		"skills/a/SKILL.md":     "a2",
		"agents/same.md":        "same",
		"skills/other/SKILL.md": "release other",
	})
	os.WriteFile(filepath.Join(claude, "skills/a/SKILL.md"), []byte("local a"), 0o644)
	p, err = NewPlan(claude, v2, "v2.0.0", m)
	if err != nil {
		t.Fatal(err)
	}
	want = "agents/same.md unchanged,skills/a update!,skills/gone remove,skills/other unchanged"
	if got := describe(p.Changes); got != want {
		t.Errorf("plan v2 = %s, want %s", got, want)
	}
	if len(p.Pending()) != 2 || len(p.Conflicts()) != 1 {
		t.Errorf("Pending = %v, Conflicts = %v", p.Pending(), p.Conflicts())
	}
	if _, err := p.Apply(claude, true, now.Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	if got := readFile(t, filepath.Join(claude, "skills/a/SKILL.md")); got != "a2" {
		t.Errorf("a = %q", got)
	}
	if _, err := os.Stat(filepath.Join(claude, "skills/gone")); !os.IsNotExist(err) {
		t.Errorf("skills/gone still exists: %v", err)
	}
	if state, _ := State(claude, "skills/a", p.hashes["skills/a"]); state != "ok" {
		t.Errorf("State = %s", state)
	}
}

func TestLoadManifestRejectsUnsafeItems(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, ManifestName), []byte(`{"version":"v1","items":{"skills/../../etc":"x"}}`), 0o644)
	if _, err := LoadManifest(dir); err == nil {
		t.Error("LoadManifest: no error for an unsafe item")
	}
}

func describe(changes []Change) string {
	parts := make([]string, len(changes))
	for i, c := range changes {
		parts[i] = c.Item + " " + string(c.Action)
		if c.Conflict != "" {
			parts[i] += "!"
		}
	}
	return strings.Join(parts, ",")
}
