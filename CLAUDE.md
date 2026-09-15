# CLAUDE.md

This repository holds Claude AI skills and agents.
You develop, test, and document them here.
chezmoi distributes them to the computers of the user.

## Layout

- `skills/<name>/`: One skill per directory, with a `SKILL.md` file.
- `agents/<name>.md`: One agent per file.
- `scripts/sync-chezmoi.sh`: Copies the skills and agents into the chezmoi source directory.
- `VERSION`: The release version of the repository.
- `docs/phase-<number>.md`: The features of one development phase.
- `CHANGELOG.md`: All user-visible changes, by version.

## Git

- Never commit.
- Never push.
- Leave all changes in the working tree.
  The user reviews the changes, commits them, and pushes them.
- If the user asks you to commit, make a branch first.
  Never commit to `main`.
- These rules also apply to the chezmoi repository.
  Never run `chezmoi git` commands that commit or push.

## Skills and agents

- Use the directory name of a skill as the `name` in its `SKILL.md` frontmatter.
- Use the file name of an agent as the `name` in its frontmatter.
- Do not give a file a leading dot or the executable bit.
  chezmoi reads both as attributes, and `scripts/sync-chezmoi.sh` stops.
- In a skill, refer to its own files with a relative path, for example `reference/ste-lint.py`.

## The scraibe skill and agent

The source of the scraibe skill is `skills/scraibe/`.
The source of the scraibe agent is `agents/scraibe.md`.
Claude Code loads the installed copies from `~/.claude/`, not the copies in this repository.

- To write documentation in this repository, use the installed scraibe skill and agent.
- To change the scraibe standard, edit `skills/scraibe/SKILL.md`.
- The standard has its own version, for example `scraibe v0.1`.
  If you change the rules of the standard, tell the user.
  The user decides the new standard version.

## Features and phases

- Record each new feature in `docs/phase-<number>.md`.
- Use the phase file with the highest number.
- Start a new phase file only when the user tells you to.
- For each feature, write the name, the purpose, and the status.
- Status is one of: planned, in progress, done.

## Changelog

- Update `CHANGELOG.md` with each user-visible change.
- Use the Keep a Changelog format.
- Add new entries under `## [Unreleased]`.
- Use only these groups: Added, Changed, Deprecated, Removed, Fixed, Security.
- Name the skill or agent at the start of each entry, for example "scraibe: ".

## Versions

- Use semantic versioning (semver).
- The first version is 0.0.1.
- One version applies to all skills and agents in the repository.
- For a bug fix, increase the patch number.
- For a new feature, increase the minor number.
- For a change that breaks existing use, increase the minor number while the major number is 0.
- Change the version only when the user tells you to make a release.

## Release

Do these steps only when the user tells you to make a release.

1. Write the new version in `VERSION`.
2. In `CHANGELOG.md`, move the `[Unreleased]` entries to a new heading `## [<version>] - <date>`.
3. Run `scripts/sync-chezmoi.sh --dry-run`, and show the output to the user.
4. If the user agrees, run `scripts/sync-chezmoi.sh`.
5. Give the user the commands to commit, tag, and push this repository.
6. Give the user the next steps that the script shows.

## Documentation and comments

- Load the `scraibe` skill before you write documentation or comments.
- Apply the scraibe standard to all Markdown files, `SKILL.md` files, agent files, and code comments.
- Run `ste-lint.py` from the installed scraibe skill on each Markdown file you change.
- In your report, give the `Doc-Draft: scraibe/0.1` trailer for the user to add to the commit.
