# Phase 1

Phase 1 sets up the repository for skill and agent development.

## Features

| Feature | Purpose | Status |
|---|---|---|
| `CLAUDE.md` | Gives the rules for Claude Code in this repository. | done |
| `CHANGELOG.md` | Records the changes for each version. | done |
| scraibe import | Makes this repository the source of the scraibe skill and agent. | done |
| `scripts/sync-chezmoi.sh` | Copies the skills and agents into the chezmoi source directory. | done |
| `VERSION` | Holds the release version and shows the installed version on each computer. | done |
| nagios-plugin skill | Gives the plugin API rules and a Go template for a static plugin binary. | done |
| nagios-plugin agent | Writes, changes, and reviews plugins with the nagios-plugin skill. | done |

## Distribution

This repository is the source.
chezmoi distributes the skills and agents to each computer.

1. `scripts/sync-chezmoi.sh` copies the files to the chezmoi source directory.
2. `chezmoi apply` installs the files in `~/.claude/` on this computer.
3. A commit and a push in the chezmoi repository publish the files.
4. `chezmoi update` installs the files on the other computers.

Each computer shows the installed version in `~/.claude/.ai-skills-version`.
