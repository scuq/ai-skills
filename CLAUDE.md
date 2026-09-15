# CLAUDE.md

This repository holds Claude AI skills and agents.
You develop, test, and document them here.
`ai-skillsctl` installs them from a GitHub release on the computers of the user.

## Layout

- `skills/<name>/`: One skill per directory, with a `SKILL.md` file.
- `agents/<name>.md`: One agent per file.
- `cmd/ai-skillsctl/`: The installer command.
- `internal/`: The Go packages of the installer.
- `scripts/test.sh`: Runs gofmt, go vet, and go test for each Go module.
- `scripts/build-release.sh`: Builds the release assets in `dist/`.
- `scripts/release-notes.sh`: Prints the `CHANGELOG.md` section of one version.
- `.github/workflows/`: The CI workflow and the release workflow.
- `VERSION`: The release version of the repository.
- `docs/phase-<number>.md`: The features of one development phase.
- `README.md`: The user documentation.
- `CHANGELOG.md`: All user-visible changes, by version.

## Git

- Never commit.
- Never push.
- Leave all changes in the working tree.
  The user reviews the changes, commits them, and pushes them.
- If the user asks you to commit, make a branch first.
  Never commit to `main`.
- Never create or push a tag.
  A tag push publishes a release.

## Skills and agents

- Use the directory name of a skill as the `name` in its `SKILL.md` frontmatter.
- Use the file name of an agent as the `name` in its frontmatter.
- In a skill, refer to its own files with a relative path, for example `reference/ste-lint.py`.
- Refer to an installed skill with the path `~/.claude/skills/<name>/`.
- Do not put a secret, a real hostname, or a personal name in a skill or an agent.
  The repository is public.

## Go code

- Run `scripts/test.sh` after each change to Go code.
- Build with `CGO_ENABLED=0`.
  Do not add a dependency that needs cgo.
- `ai-skillsctl` uses only the Go standard library.
  Ask the user before you add a dependency to it.
- Add a test for each change to the behavior of `ai-skillsctl`.
- If you change a command, an option, or an exit code, update `README.md`.

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
- One version applies to all skills, agents, and `ai-skillsctl`.
- For a bug fix, increase the patch number.
- For a new feature, increase the minor number.
- For a change that breaks existing use, increase the minor number while the major number is 0.
- Change the version only when the user tells you to make a release.

## Release

Do these steps only when the user tells you to make a release.

1. Write the new version in `VERSION`.
2. In `CHANGELOG.md`, move the `[Unreleased]` entries to a new heading `## [<version>] - <date>`.
3. Run `scripts/build-release.sh v<version>`, and show the result to the user.
4. Give the user the commands to commit to `main`, push, create the tag `v<version>`, and push the tag.
5. After the release workflow completes, tell the user to run `ai-skillsctl update` on each computer.

## Documentation and comments

- Load the `scraibe` skill before you write documentation or comments.
- Apply the scraibe standard to all Markdown files, `SKILL.md` files, agent files, and code comments.
- Run `ste-lint.py` from the installed scraibe skill on each Markdown file you change.
- In your report, give the `Doc-Draft: scraibe/0.1` trailer for the user to add to the commit.
