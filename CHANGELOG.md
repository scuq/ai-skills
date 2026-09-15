# Changelog

All user-visible changes to this project are in this file.
The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).
This project uses [semantic versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- scraibe: The skill and the agent are now in this repository, with the standard at version 0.1.
- `scripts/sync-chezmoi.sh` copies the skills and agents into the chezmoi source directory.
- `VERSION` holds the release version, and the sync script installs it as `~/.claude/.ai-skills-version`.
- nagios-plugin: The skill gives the rules for plugins for Nagios Core, Naemon, and Icinga 2, with a Go template for a static binary.
- nagios-plugin: The template reads credentials from an age file with a passphrase from a file, an option, or `NAGIOS_PLUGIN_PASSPHRASE`.
- nagios-plugin: The agent writes, changes, and reviews plugins with the skill and the scraibe standard.

## [0.0.1] - 2026-09-15

### Added

- `CLAUDE.md` gives the rules for Claude Code in this repository.
- `CHANGELOG.md` records the changes for each version.
- `docs/phase-1.md` records the features of phase 1.
