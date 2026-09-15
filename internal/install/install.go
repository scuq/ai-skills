// Package install installs the skills and agents of an ai-skills bundle in
// a Claude configuration directory, and records them in a manifest.
package install

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// ManifestName is the file name of the manifest in the Claude configuration
// directory.
const ManifestName = ".ai-skills.json"

// BackupDirName is the directory in the Claude configuration directory that
// holds the local versions that a forced update replaced.
const BackupDirName = ".ai-skills-backup"

// A Manifest records the installed release and the items that it installed.
type Manifest struct {
	// Version is the tag of the installed release, for example "v0.1.0".
	Version string `json:"version"`
	// Updated is the time of the last installation, in UTC.
	Updated time.Time `json:"updated"`
	// Items holds the hash of each installed item at installation time,
	// by item path, for example "skills/scraibe".
	Items map[string]string `json:"items"`
}

// LoadManifest reads the manifest in claudeDir.
// If the manifest does not exist, LoadManifest returns an empty manifest.
func LoadManifest(claudeDir string) (Manifest, error) {
	m := Manifest{Items: map[string]string{}}
	path := filepath.Join(claudeDir, ManifestName)
	data, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return m, nil
	}
	if err != nil {
		return m, err
	}
	if err := json.Unmarshal(data, &m); err != nil {
		return m, fmt.Errorf("manifest %s is not valid JSON: %w", path, err)
	}
	if m.Items == nil {
		m.Items = map[string]string{}
	}
	for item := range m.Items {
		if !validItem(item) {
			return m, fmt.Errorf("manifest %s has the item %q, which is not skills/NAME or agents/NAME.md", path, item)
		}
	}
	return m, nil
}

// Save writes m to claudeDir. It writes a temporary file and then renames it,
// so that a failure does not leave a partial manifest.
func (m Manifest) Save(claudeDir string) error {
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	f, err := os.CreateTemp(claudeDir, ManifestName+".tmp-")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	if _, err := f.Write(append(data, '\n')); err != nil {
		f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	return os.Rename(f.Name(), filepath.Join(claudeDir, ManifestName))
}

// An Action is the change that a plan makes to one item.
type Action string

// The actions of a plan.
const (
	// Unchanged: the installed item is the same as the item in the release.
	Unchanged Action = "unchanged"
	// Install: the item does not exist in the Claude directory.
	Install Action = "install"
	// Update: ai-skillsctl installed the item, and the release has a different version.
	Update Action = "update"
	// Adopt: ai-skillsctl did not install the item, but it is the same as the item in the release.
	Adopt Action = "adopt"
	// Replace: ai-skillsctl did not install the item, and it is different from the item in the release.
	Replace Action = "replace"
	// Remove: ai-skillsctl installed the item, and the release does not have it.
	Remove Action = "remove"
)

// A Change is the planned change to one item.
type Change struct {
	Item   string
	Action Action
	// Conflict is the reason that the change overwrites local content.
	// It is empty when the change keeps all local content.
	Conflict string
}

// A Plan is the list of changes that installs one release.
type Plan struct {
	// Version is the tag of the release.
	Version string
	// Changes holds one change for each item, sorted by item path.
	Changes []Change

	staged string
	hashes map[string]string
}

// NewPlan compares the items in staged, the directory of the extracted
// bundle of the release version, with the items in claudeDir and m.
func NewPlan(claudeDir, staged, version string, m Manifest) (*Plan, error) {
	items, err := Items(staged)
	if err != nil {
		return nil, err
	}
	p := &Plan{Version: version, staged: staged, hashes: map[string]string{}}

	for _, item := range items {
		newHash, err := Hash(filepath.Join(staged, filepath.FromSlash(item)))
		if err != nil {
			return nil, err
		}
		p.hashes[item] = newHash
		current, err := hashIfExists(filepath.Join(claudeDir, filepath.FromSlash(item)))
		if err != nil {
			return nil, err
		}
		recorded, managed := m.Items[item]

		c := Change{Item: item}
		switch {
		case current == "":
			c.Action = Install
		case current == newHash && managed:
			c.Action = Unchanged
		case current == newHash:
			c.Action = Adopt
		case !managed:
			c.Action, c.Conflict = Replace, "not installed by ai-skillsctl, content is different"
		case current != recorded:
			c.Action, c.Conflict = Update, "local changes"
		default:
			c.Action = Update
		}
		p.Changes = append(p.Changes, c)
	}

	for item, recorded := range m.Items {
		if _, ok := p.hashes[item]; ok {
			continue
		}
		current, err := hashIfExists(filepath.Join(claudeDir, filepath.FromSlash(item)))
		if err != nil {
			return nil, err
		}
		switch {
		case current == "":
			// The item is already gone. The new manifest drops it.
		case current != recorded:
			p.Changes = append(p.Changes, Change{Item: item, Action: Remove, Conflict: "local changes"})
		default:
			p.Changes = append(p.Changes, Change{Item: item, Action: Remove})
		}
	}

	sort.Slice(p.Changes, func(i, j int) bool { return p.Changes[i].Item < p.Changes[j].Item })
	return p, nil
}

// Pending returns the changes with an action other than Unchanged.
func (p *Plan) Pending() []Change {
	var out []Change
	for _, c := range p.Changes {
		if c.Action != Unchanged {
			out = append(out, c)
		}
	}
	return out
}

// Conflicts returns the changes that overwrite local content.
func (p *Plan) Conflicts() []Change {
	var out []Change
	for _, c := range p.Changes {
		if c.Conflict != "" {
			out = append(out, c)
		}
	}
	return out
}

// Apply makes the changes of p in claudeDir and writes the new manifest.
// It returns the backup directory, or an empty string if it made no backup.
//
// If p has conflicts and force is false, Apply changes nothing and returns
// an error. If force is true, Apply first moves each item with a conflict to
// a new directory in [BackupDirName].
func (p *Plan) Apply(claudeDir string, force bool, now time.Time) (string, error) {
	conflicts := p.Conflicts()
	if len(conflicts) > 0 && !force {
		return "", fmt.Errorf("%d items have local changes", len(conflicts))
	}
	backup := ""
	if len(conflicts) > 0 {
		backup = filepath.Join(claudeDir, BackupDirName, now.UTC().Format("20060102T150405Z"))
	}

	for _, c := range p.Changes {
		target := filepath.Join(claudeDir, filepath.FromSlash(c.Item))
		if c.Conflict != "" {
			dst := filepath.Join(backup, filepath.FromSlash(c.Item))
			if err := os.MkdirAll(filepath.Dir(dst), 0o700); err != nil {
				return backup, err
			}
			if err := os.Rename(target, dst); err != nil {
				return backup, fmt.Errorf("move %s to the backup: %w", c.Item, err)
			}
		}
		switch c.Action {
		case Install, Update, Replace:
			if err := replace(filepath.Join(p.staged, filepath.FromSlash(c.Item)), target); err != nil {
				return backup, fmt.Errorf("install %s: %w", c.Item, err)
			}
		case Remove:
			if err := os.RemoveAll(target); err != nil {
				return backup, fmt.Errorf("remove %s: %w", c.Item, err)
			}
		}
	}

	m := Manifest{Version: p.Version, Updated: now.UTC(), Items: p.hashes}
	return backup, m.Save(claudeDir)
}

// replace moves src to dst. It removes an existing dst only after src is in
// place, and it restores dst if the move fails.
func replace(src, dst string) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	old := ""
	if _, err := os.Lstat(dst); err == nil {
		old = dst + ".ai-skills-old"
		if err := os.RemoveAll(old); err != nil {
			return err
		}
		if err := os.Rename(dst, old); err != nil {
			return err
		}
	}
	if err := os.Rename(src, dst); err != nil {
		if old != "" {
			os.Rename(old, dst)
		}
		return err
	}
	if old != "" {
		return os.RemoveAll(old)
	}
	return nil
}

// State returns the state of the installed item: "ok" if its hash is
// recorded, "modified" if it is different, and "missing" if it does not
// exist.
func State(claudeDir, item, recorded string) (string, error) {
	current, err := hashIfExists(filepath.Join(claudeDir, filepath.FromSlash(item)))
	switch {
	case err != nil:
		return "", err
	case current == "":
		return "missing", nil
	case current != recorded:
		return "modified", nil
	}
	return "ok", nil
}
