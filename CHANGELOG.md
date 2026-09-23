# Changelog

All user-visible changes to this project are in this file.
The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).
This project uses [semantic versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.3.0] - 2026-09-24

### Added

- scuq-scraibe: Section 2b of the standard gives 12 patterns for the choice of a word without the dictionary. The patterns come from an analysis of the ASD-STE100 Issue 9 dictionary.
- scuq-scraibe: `ste-lint.py` gives an advisory finding for a verb in -ize, -ise, or -ify.

### Changed

- scuq-scraibe: The scraibe standard is now version 0.2. The commit trailer is `Doc-Draft: scraibe/0.2`.
- scuq-scraibe: The skill, the agent, and `ste-lint.py` give the copyright and trade mark statement of ASD for ASD-STE100.
- scuq-scraibe: The attribution section states that ASD and the STEMG do not endorse or approve this standard.
- scuq-scraibe: In check mode, the agent gives the copyright and trade mark statement with its result.

## [0.2.0] - 2026-09-20

### Added

- scuq-hmi-semantics: The skill gives the rules for the meaning of a state, an alert level, and a control in an operator interface. It has a template for a semantic registry, an alert contract, and a review checklist.
- scuq-hmi-semantics: The agent builds and reviews operator interfaces with the skill and the scraibe standard.

## [0.1.0] - 2026-09-16

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
