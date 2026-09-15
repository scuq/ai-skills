# Changelog

All user-visible changes to this project are in this file.
The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).
This project uses [semantic versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- scuq-scraibe: The skill and the agent are now in this repository, with the standard at version 0.1.
- scuq-nagios-plugin: The skill gives the rules for plugins for Nagios Core, Naemon, and Icinga 2, with a Go template for a static binary.
- scuq-nagios-plugin: The template reads credentials from an age file with a passphrase from a file, an option, or `NAGIOS_PLUGIN_PASSPHRASE`.
- scuq-nagios-plugin: The agent writes, changes, and reviews plugins with the skill and the scraibe standard.
- All skills and agents have the name prefix `scuq-`, so that they do not replace items of the user with the same name.
- `ai-skillsctl update` installs, updates, and removes the skills and agents of a release in `~/.claude`.
- If an installed item has local changes, `ai-skillsctl update` stops, and `--force` moves the local version to `~/.claude/.ai-skills-backup/`.
- `ai-skillsctl status` shows the installed release, the state of each item, and the latest release.
- `ai-skillsctl self-update` replaces the binary with the binary of a release.
- A push of a tag `vMAJOR.MINOR.PATCH` publishes a GitHub release with static `ai-skillsctl` binaries for Linux and macOS, the skill bundle, and `SHA256SUMS`.
- `VERSION` holds the release version.

## [0.0.1] - 2026-09-15

### Added

- `CLAUDE.md` gives the rules for Claude Code in this repository.
- `CHANGELOG.md` records the changes for each version.
- `docs/phase-1.md` records the features of phase 1.
