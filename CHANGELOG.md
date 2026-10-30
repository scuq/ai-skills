# Changelog

All user-visible changes to this project are in this file.
The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).
This project uses [semantic versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- scuq-hmi-semantics: Rule 3j separates a severity from an alert. Only a severity that the registry maps to an alert level uses an alert color.
- scuq-hmi-semantics: Rule 3k adds view truth. A filter, a saved view, or a hidden column cannot hide the alert area, and the interface shows the action-required records outside the view.
- scuq-hmi-semantics: Rule 3k tells a client filter from a server filter, and gives the loading state and the failure state of a filter.
- scuq-hmi-semantics: Rule 3n gives the behavior of a modal dialog: focus, the Tab key, closure, and the alert area above the backdrop.
- scuq-hmi-semantics: Rule 3l gives one bulk flow for every module, from the selection to the result for each record.
- scuq-hmi-semantics: Rule 3m adds adaptive layout. The layout follows the CSS width, the zoom, and the pointer type, and density never changes the state coding.
- scuq-hmi-semantics: `assets/micro-parts.html` gives small HTML and CSS parts that hold the semantics, with registry tokens only.
- scuq-hmi-semantics: The registry template has `severities`, `short_label`, `entity_fields`, `layouts`, and the levels `normal` and `stale`.
- scuq-hmi-semantics: The review checklist has a View truth group and the items S6, L5 to L8, K7 to K13, and A6.

### Changed

- scuq-hmi-semantics: Rule 3f gives a button one short verb. The target, the count, and the environment now go in an always-visible action summary line beside the button.
- scuq-hmi-semantics: Rule 3b asks for the coverage next to a count or a trend. Rule 3e limits what a tooltip or a progressive disclosure can hide. Rule 3g applies to keyboard shortcuts and the command palette.
- scuq-hmi-semantics: The alert contract defines UNKNOWN and STALE as two levels.

## [0.5.0] - 2026-10-25

### Added

- scuq-scraibe: `ste-lint.py` has the rule `ghost-command`. It flags a command in the prose or in a shell example that the Usage or Synopsis section does not have.
- scuq-scraibe: `ste-lint.py --fix` changes a semicolon between clauses to a period, removes a marketing adjective, and replaces a soft phrasal verb. It changes nothing else.
- scuq-scraibe: Section 6 of the standard gives three questions for each comment, and asks for evergreen comments and a source for each constant.
- scuq-scraibe: Section 12 of the standard has two more review habits. A sentence that fits any project carries no information, and a claim of significance needs the mechanism.
- scuq-scraibe: `ste-lint.py` tests more rules of the standard. Hard: Latin abbreviations, contractions, parenthetical plurals, em and en dashes, the filler words of section 2, and a numeral at the start of a sentence. Advisory: bold, a condition after the instruction, a sentence that starts with "This", "-ing" verb forms, and title case in a heading.
- scuq-scraibe: A directive comment such as `<!-- ste-lint: disable-next-line semicolon -->` disables a rule for a line, a block, or a file. A quotation of a bad example can then pass the linter.
- scuq-scraibe: `ste-lint.py --fix` also writes a contraction in -n't as two words. It removes two filler adverbs, and it replaces five filler phrases and two Latin abbreviations with the plain words.
- scuq-scraibe: `evals/evals.json` holds four evaluation tasks with fixtures: a README without BLUF, a README with a stale BLUF, Go doc comments, and a CHANGELOG from commits. `evals/run.sh` runs them with `claude -p` in a temporary configuration directory and tests the result with shell checks. `--without-skill` gives the baseline.

- scuq-scraibe: `ste-lint.py` has the rule `ghost-flag`. It flags a flag in the prose or in a shell example that the Usage, Synopsis, Options, and Configuration sections do not have.
- scuq-scraibe: `ste-lint.py` reads a glossary `.scraibe.json` from the repository root. The rule `glossary` flags the words that the glossary replaces, and the glossary can disable rules and set the baseline.
- scuq-scraibe: `reference/scraibe-init.sh` installs the linter in a repository. It writes a copy in `.scraibe/`, an empty glossary, a pre-commit hook, and a GitHub Actions job. The hook lints the staged Markdown files, and the job lints every Markdown file.

### Changed

- scuq-scraibe: The scraibe standard is now version 0.4. The commit trailer is `Doc-Draft: scraibe/0.4`.
- scuq-scraibe: The agent file is shorter. It has the same steps, and it points at the reference files of the skill instead of a copy of their text.
- scuq-scraibe: When a repository has no `.scraibe.json`, the agent and the hook suggest `scraibe-init.sh` once. They do not run it.
- scuq-scraibe: `SKILL.md` is under 500 lines and about 5000 tokens. Seven files under `reference/` hold the detail: names, word choice, interface instructions, text types, comments by language, lint details, and attribution. The standard keeps a short rule and a pointer for each.
- scuq-scraibe: The grep in the review checklist of the standard is shorter. The linter tests the patterns that the grep tested.
- The CI workflow runs the linter self-test and lints the skills and the agents too.
- The release bundle no longer has the `evals/` directory of a skill.

### Fixed

- scuq-scraibe: `ste-lint.py` and `--fix` skip an indented code block, as they skip a fenced code block. Before, the lines of a Usage section got sentence findings.

## [0.4.0] - 2026-10-25

### Added

- scuq-scraibe: Section 4c of the standard gives the rules for the BLUF section (bottom line up front) at the top of every prose document. The BLUF is one paragraph of at most 80 words. It states what the document describes, its state, and the first action.
- scuq-scraibe: `ste-lint.py` flags a `README.md` without a BLUF section. It also flags a BLUF that is not the first section, or that has more than 80 words or six sentences. A list, a code block, or a link in the BLUF is a finding too.

### Changed

- scuq-scraibe: The scraibe standard is now version 0.3. The commit trailer is `Doc-Draft: scraibe/0.3`.
- scuq-scraibe: On every task, the agent compares the BLUF of the top-level `README.md` with the code, the commits since the last tag, and the `[Unreleased]` entries. If a sentence is not true now, the agent corrects it in the same task. The report has a `BLUF:` line.
- scuq-scraibe: The README structure in section 4 has the section BLUF after the title.

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
