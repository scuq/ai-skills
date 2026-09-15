# ai-skills - Claude Code skills and agents, with the ai-skillsctl installer

This repository holds skills and agents for Claude Code.
`ai-skillsctl` installs them from a GitHub release into the Claude configuration directory.

## Usage

```text
ai-skillsctl update [--version TAG] [--force] [--dry-run] [options]
ai-skillsctl status [--offline] [options]
ai-skillsctl self-update [--version TAG] [--force] [options]
ai-skillsctl version
```

To install `ai-skillsctl` on Linux and install the latest release of the skills and agents:

```sh
os=$(uname -s | tr '[:upper:]' '[:lower:]')
arch=$(uname -m | sed 's/x86_64/amd64/; s/aarch64/arm64/')
base=https://github.com/scuq/ai-skills/releases/latest/download
cd "$(mktemp -d)"
curl -fsSL -O "$base/ai-skillsctl-$os-$arch" -O "$base/SHA256SUMS"
sha256sum --ignore-missing -c SHA256SUMS
mkdir -p ~/.local/bin
install -m 0755 "ai-skillsctl-$os-$arch" ~/.local/bin/ai-skillsctl
ai-skillsctl update
```

On macOS, use `shasum -a 256 --ignore-missing -c SHA256SUMS` in place of `sha256sum`.

## Description

The repository holds these items:

| Item | Content |
|---|---|
| `skills/scuq-scraibe/` | The scraibe standard for documentation and code comments |
| `skills/scuq-nagios-plugin/` | The rules and a Go template for plugins for Nagios Core, Naemon, and Icinga 2 |
| `agents/scuq-scraibe.md` | An agent that writes documentation in the scraibe standard |
| `agents/scuq-nagios-plugin.md` | An agent that writes and reviews monitoring plugins |

Each skill and agent has the name prefix `scuq-`.
The prefix prevents a conflict with a skill or an agent of your own that has the same name.

A release holds a bundle of the items, the `ai-skillsctl` binaries, and a `SHA256SUMS` file.

`ai-skillsctl update` downloads the bundle and verifies its SHA-256 sum.
Then it compares each item in the bundle with the item in the Claude configuration directory:

| Action | Condition |
|---|---|
| install | The item does not exist. |
| update | `ai-skillsctl` installed the item, and the release has a different version. |
| adopt | `ai-skillsctl` did not install the item, but the content is the same. |
| replace | `ai-skillsctl` did not install the item, and the content is different. |
| remove | `ai-skillsctl` installed the item, and the release does not have it. |

If an item has local changes, `update` changes no file and exits with 1.
With `--force`, it moves the local version to a backup directory first.
`ai-skillsctl` does not touch skills and agents that are not in the release.

## Options

`--version TAG`
: The release tag, for example `v0.1.0`. The default is the latest release.

`--force`
: For `update`, move the local changes to a backup directory, and install the release. For `self-update`, replace the binary also when it has the same version.

`--dry-run`
: For `update`, show the changes, but do not change files.

`--offline`
: For `status`, do not get the latest release from GitHub.

`--claude-dir DIR`
: The Claude configuration directory. The default is `$CLAUDE_CONFIG_DIR`, or `~/.claude`.

`--repo OWNER/NAME`
: The GitHub repository. The default is `scuq/ai-skills`.

## Exit status

| Code | Condition |
|---|---|
| 0 | The command completed. |
| 1 | The command failed, or `update` found local changes without `--force`. |
| 2 | The command line is not correct. |

## Environment

`CLAUDE_CONFIG_DIR`
: The Claude configuration directory, if `--claude-dir` is not set.

`GITHUB_TOKEN`
: A GitHub token for the API requests. Without a token, GitHub permits 60 API requests per hour for each IP address.

## Files

`~/.claude/skills/NAME/`, `~/.claude/agents/NAME.md`
: The installed skills and agents.

`~/.claude/.ai-skills.json`
: The manifest: the installed release and the hash of each installed item.

`~/.claude/.ai-skills-backup/TIMESTAMP/`
: The local versions that `update --force` replaced.

## Examples

Show the changes of the latest release without a change to files:

```sh
ai-skillsctl update --dry-run
```

Install a specified release:

```sh
ai-skillsctl update --version v0.1.0
```

Replace the binary with the binary of the latest release:

```sh
ai-skillsctl self-update
```

## Release

A push of a tag `vMAJOR.MINOR.PATCH` starts `.github/workflows/release.yml`.
The workflow runs the tests, builds the assets with `scripts/build-release.sh`, and publishes the release.

1. Write the version in `VERSION`, for example `0.1.0`.
2. In `CHANGELOG.md`, move the `[Unreleased]` entries to the heading `## [0.1.0] - DATE`.
3. Run `scripts/build-release.sh v0.1.0` to test the build.
4. Commit the change to `main`, and push it.
5. Create the tag with `git tag -a v0.1.0 -m v0.1.0`, and push it with `git push origin v0.1.0`.

CAUTION: Do not push a tag that does not match `VERSION`. The workflow stops, and you must remove the tag and push it again.

## See also

`CHANGELOG.md`, `CLAUDE.md`, `docs/phase-1.md`
