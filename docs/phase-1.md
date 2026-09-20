# Phase 1

Phase 1 sets up the repository for skill and agent development and distribution.

## Features

| Feature | Purpose | Status |
|---|---|---|
| `CLAUDE.md` | Gives the rules for Claude Code in this repository. | done |
| `CHANGELOG.md` | Records the changes for each version. | done |
| scuq-scraibe import | Makes this repository the source of the scuq-scraibe skill and agent. | done |
| `VERSION` | Holds the release version. | done |
| scuq-nagios-plugin skill | Gives the plugin API rules and a Go template for a static plugin binary. | done |
| scuq-nagios-plugin agent | Writes, changes, and reviews plugins with the scuq-nagios-plugin skill. | done |
| scuq-hmi-semantics skill | Gives the rules for the meaning of a state, an alert, and a control in an operator interface. | done |
| scuq-hmi-semantics agent | Builds and reviews operator interfaces with the scuq-hmi-semantics skill. | done |
| `ai-skillsctl` | Installs and updates the skills and agents of a release in `~/.claude`. | done |
| Release workflow | Publishes the `ai-skillsctl` binaries and the skill bundle for a version tag. | done |
| CI workflow | Tests the Go modules and lints the documentation on each branch push. | done |
| `scuq-` name prefix | Prevents a conflict with a skill or an agent of the user with the same name. `scripts/test.sh` checks it. | done |
| `scripts/sync-chezmoi.sh` | Copied the skills and agents into the chezmoi source directory. `ai-skillsctl` replaces it. | done |

## Distribution

This repository is the source.
GitHub releases distribute the skills and agents to each computer.

1. A push of a version tag starts the release workflow.
2. The workflow publishes the `ai-skillsctl` binaries, `ai-skills-bundle.tar.gz`, and `SHA256SUMS`.
3. On each computer, `ai-skillsctl update` installs the bundle in `~/.claude/`.

chezmoi no longer distributes `~/.claude/skills/` and `~/.claude/agents/`.
The chezmoi file `.chezmoiignore` excludes them.
