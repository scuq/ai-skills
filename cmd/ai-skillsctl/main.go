// ai-skillsctl installs and updates the skills and agents of the ai-skills
// repository in the Claude configuration directory, from a GitHub release.
//
// Usage:
//
//	ai-skillsctl update [--version TAG] [--force] [--dry-run] [options]
//	ai-skillsctl status [--offline] [options]
//	ai-skillsctl self-update [--version TAG] [--force] [options]
//	ai-skillsctl version
//
// Options:
//
//	--claude-dir DIR   Claude configuration directory
//	                   (default: $CLAUDE_CONFIG_DIR, or ~/.claude)
//	--repo OWNER/NAME  GitHub repository (default: scuq/ai-skills)
//
// If GITHUB_TOKEN is set, ai-skillsctl sends it with the GitHub API requests.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"time"

	"github.com/scuq/ai-skills/internal/install"
	"github.com/scuq/ai-skills/internal/release"
)

// version is set at build time with -ldflags "-X main.version=VERSION".
var version = "0.0.0-dev"

const (
	defaultRepo   = "scuq/ai-skills"
	defaultAPIURL = "https://api.github.com"
	maxBundleSize = 256 << 20
	maxBinarySize = 64 << 20
)

// The exit codes.
const (
	exitOK    = 0
	exitError = 1
	exitUsage = 2
)

const usage = `Usage:
  ai-skillsctl update [--version TAG] [--force] [--dry-run] [options]
  ai-skillsctl status [--offline] [options]
  ai-skillsctl self-update [--version TAG] [--force] [options]
  ai-skillsctl version

Commands:
  update       Install the skills and agents of a release. The default
               is the latest release.
  status       Show the installed release, the state of each item, and
               the latest release.
  self-update  Replace this binary with the binary of a release.
  version      Show the version of this binary.

Options:
  --version TAG      Release tag, for example v0.1.0
  --force            Move local changes to a backup and continue
  --dry-run          Show the changes, but do not change files
  --offline          Do not look for the latest release
  --claude-dir DIR   Claude configuration directory
                     (default: $CLAUDE_CONFIG_DIR, or ~/.claude)
  --repo OWNER/NAME  GitHub repository (default: scuq/ai-skills)

Set GITHUB_TOKEN to raise the GitHub API rate limit.
`

func main() {
	os.Exit(run(context.Background(), os.Args[1:], os.Stdout, os.Stderr))
}

// usageError is an error in the command line.
type usageError struct{ error }

// run runs the command in args and returns the exit code.
func run(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprint(stderr, usage)
		return exitUsage
	}
	var err error
	switch cmd, rest := args[0], args[1:]; cmd {
	case "update":
		err = cmdUpdate(ctx, rest, stdout)
	case "status":
		err = cmdStatus(ctx, rest, stdout)
	case "self-update":
		err = cmdSelfUpdate(ctx, rest, stdout)
	case "version", "--version", "-V":
		fmt.Fprintf(stdout, "ai-skillsctl %s\n", version)
		return exitOK
	case "help", "--help", "-h":
		fmt.Fprint(stdout, usage)
		return exitOK
	default:
		fmt.Fprintf(stderr, "Command %q is not known.\n\n%s", cmd, usage)
		return exitUsage
	}

	var ue usageError
	switch {
	case err == nil:
		return exitOK
	case errors.Is(err, flag.ErrHelp):
		fmt.Fprint(stdout, usage)
		return exitOK
	case errors.As(err, &ue):
		fmt.Fprintf(stderr, "%v\n\n%s", err, usage)
		return exitUsage
	}
	fmt.Fprintf(stderr, "ai-skillsctl: %v\n", err)
	return exitError
}

// globals holds the options that all commands accept.
type globals struct {
	claudeDir string
	repo      string
	apiURL    string
}

func newFlagSet(name string, g *globals) *flag.FlagSet {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	fs.StringVar(&g.claudeDir, "claude-dir", defaultClaudeDir(), "")
	fs.StringVar(&g.repo, "repo", defaultRepo, "")
	// api-url is for tests. The usage text does not show it.
	fs.StringVar(&g.apiURL, "api-url", defaultAPIURL, "")
	return fs
}

func parse(fs *flag.FlagSet, g *globals, args []string) error {
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return err
		}
		return usageError{err}
	}
	if fs.NArg() > 0 {
		return usageError{fmt.Errorf("argument %q is not an option", fs.Arg(0))}
	}
	if g.claudeDir == "" {
		return usageError{errors.New("the home directory is unknown, set --claude-dir")}
	}
	return nil
}

func defaultClaudeDir() string {
	if dir := os.Getenv("CLAUDE_CONFIG_DIR"); dir != "" {
		return dir
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".claude")
}

func (g globals) client() *release.Client {
	return &release.Client{APIURL: g.apiURL, Repo: g.repo, Token: os.Getenv("GITHUB_TOKEN")}
}

func cmdUpdate(ctx context.Context, args []string, stdout io.Writer) error {
	var g globals
	var tag string
	var force, dryRun bool
	fs := newFlagSet("update", &g)
	fs.StringVar(&tag, "version", "", "")
	fs.BoolVar(&force, "force", false, "")
	fs.BoolVar(&dryRun, "dry-run", false, "")
	if err := parse(fs, &g, args); err != nil {
		return err
	}

	c := g.client()
	rel, err := c.Find(ctx, tag)
	if err != nil {
		return err
	}
	m, err := install.LoadManifest(g.claudeDir)
	if err != nil {
		return err
	}
	bundle, err := c.DownloadVerified(ctx, rel, install.BundleName, maxBundleSize)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(g.claudeDir, 0o700); err != nil {
		return err
	}
	// The staging directory is in the Claude directory, so that a rename
	// moves each item on the same filesystem.
	staged, err := os.MkdirTemp(g.claudeDir, ".ai-skills-tmp-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(staged)
	if err := install.Extract(bundle, staged); err != nil {
		return err
	}

	plan, err := install.NewPlan(g.claudeDir, staged, rel.Tag, m)
	if err != nil {
		return err
	}
	pending := plan.Pending()
	if len(pending) == 0 && m.Version == rel.Tag {
		fmt.Fprintf(stdout, "ai-skills %s is up to date in %s.\n", rel.Tag, g.claudeDir)
		return nil
	}

	from := m.Version
	if from == "" {
		from = "none"
	}
	fmt.Fprintf(stdout, "ai-skills %s -> %s in %s\n", from, rel.Tag, g.claudeDir)
	for _, ch := range pending {
		if ch.Conflict != "" {
			fmt.Fprintf(stdout, "  %-8s %s (%s)\n", ch.Action, ch.Item, ch.Conflict)
		} else {
			fmt.Fprintf(stdout, "  %-8s %s\n", ch.Action, ch.Item)
		}
	}

	conflicts := plan.Conflicts()
	if dryRun {
		fmt.Fprintln(stdout, "Dry run. No file changed.")
		return nil
	}
	if len(conflicts) > 0 && !force {
		return fmt.Errorf("%d items have local changes, and no file changed. "+
			"To move the local versions to %s and install the release, run the command again with --force",
			len(conflicts), filepath.Join(g.claudeDir, install.BackupDirName))
	}
	backup, err := plan.Apply(g.claudeDir, force, time.Now())
	if backup != "" {
		fmt.Fprintf(stdout, "The local versions are in %s.\n", backup)
	}
	if err != nil {
		return err
	}
	fmt.Fprintf(stdout, "Installed ai-skills %s.\n", rel.Tag)
	return nil
}

func cmdStatus(ctx context.Context, args []string, stdout io.Writer) error {
	var g globals
	var offline bool
	fs := newFlagSet("status", &g)
	fs.BoolVar(&offline, "offline", false, "")
	if err := parse(fs, &g, args); err != nil {
		return err
	}

	m, err := install.LoadManifest(g.claudeDir)
	if err != nil {
		return err
	}
	if m.Version == "" {
		fmt.Fprintf(stdout, "Installed: none in %s\n", g.claudeDir)
	} else {
		fmt.Fprintf(stdout, "Installed: %s in %s, updated %s\n", m.Version, g.claudeDir, m.Updated.Format(time.RFC3339))
	}
	items := make([]string, 0, len(m.Items))
	for item := range m.Items {
		items = append(items, item)
	}
	sort.Strings(items)
	for _, item := range items {
		state, err := install.State(g.claudeDir, item, m.Items[item])
		if err != nil {
			return err
		}
		fmt.Fprintf(stdout, "  %-8s %s\n", state, item)
	}

	if offline {
		return nil
	}
	rel, err := g.client().Find(ctx, "")
	if err != nil {
		return err
	}
	fmt.Fprintf(stdout, "Latest release: %s\n", rel.Tag)
	if rel.Tag != m.Version {
		fmt.Fprintln(stdout, "To install it, run: ai-skillsctl update")
	}
	return nil
}

func cmdSelfUpdate(ctx context.Context, args []string, stdout io.Writer) error {
	var g globals
	var tag string
	var force bool
	fs := newFlagSet("self-update", &g)
	fs.StringVar(&tag, "version", "", "")
	fs.BoolVar(&force, "force", false, "")
	if err := parse(fs, &g, args); err != nil {
		return err
	}

	c := g.client()
	rel, err := c.Find(ctx, tag)
	if err != nil {
		return err
	}
	if rel.Tag == "v"+version && !force {
		fmt.Fprintf(stdout, "ai-skillsctl %s is up to date.\n", rel.Tag)
		return nil
	}
	name := fmt.Sprintf("ai-skillsctl-%s-%s", runtime.GOOS, runtime.GOARCH)
	bin, err := c.DownloadVerified(ctx, rel, name, maxBinarySize)
	if err != nil {
		return err
	}
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	if exe, err = filepath.EvalSymlinks(exe); err != nil {
		return err
	}
	if err := replaceExecutable(exe, bin); err != nil {
		return fmt.Errorf("replace %s: %w", exe, err)
	}
	fmt.Fprintf(stdout, "Replaced %s: ai-skillsctl %s -> %s\n", exe, version, rel.Tag)
	return nil
}

// replaceExecutable writes data to a temporary file in the directory of path
// and renames it to path, so that a running process keeps the old file.
func replaceExecutable(path string, data []byte) error {
	f, err := os.CreateTemp(filepath.Dir(path), ".ai-skillsctl-new-")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	if _, err := f.Write(data); err != nil {
		f.Close()
		return err
	}
	if err := f.Chmod(0o755); err != nil {
		f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	return os.Rename(f.Name(), path)
}
