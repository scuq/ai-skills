---
name: scuq-scraibe
description: >
  scraibe v0.4, the writing standard for all documentation and code
  comments. Load when you write or review a README, docs/, CHANGELOG,
  man page, Go doc comment, Python docstring, .NET XML doc, Bash
  header, or inline comment.
---

<!-- ste-lint: disable-file filler-word,marketing-adjective,nominalization,phrasal-verb,parenthetical-plural,synonym-rotation -->

# scraibe v0.4

## BLUF

scraibe is one standard for all documentation and comments: short
text, simple words, no missing facts. It comes from ASD-STE100 Issue
9 (2025), man-pages(7), and go.dev/doc/comment, and this file decides
when they disagree. Every prose document starts with a BLUF section
(section 4c). On every task in a repository, make sure the BLUF of
the top-level README.md is true now.

## 1. Priority

Apply these in order. A lower rule never breaks a higher one.

1. Correct. The text says what the code does. Read the code first.
   Never add a fact the code does not show.
2. Complete. The reader can do the task with this text only. Never
   delete a necessary fact to make the text shorter.
3. Same claim. Never change the strength of a statement. A rewrite
   that turns a possibility into a fact is a different claim. Say the
   strength with STE words: "can" for a possibility, "must" for a
   requirement, "unknown" for an unknown cause, a stated condition
   for a case. Never should, would, may, might, could. A required
   "should" becomes "must", an optional "should" is deleted. "May
   have failed" becomes "possibly failed" or "failed in some cases",
   not "failed".
4. Short. Delete every word that carries no information.
5. Style. All rules below.

## 1a. Two kinds of rules, two modes

Structural rules describe the shape of a sentence, and `ste-lint.py`
tests most of them. Lexical rules depend on the dictionary, which
this standard does not contain: apply them as a direction, not as a
test.

Strict mode, every rule: comments, docstrings, error messages, CLI
help, man pages, CHANGELOG entries, warnings, procedures. Relaxed
mode, all structural rules (section 3) and the replace table as
advice: README narrative, docs/ explanations, design notes. When the
user does not name a mode, pick one from the text type.

## 2. Words

Use only four groups of words: approved STE words, technical nouns
(identifiers, flags, paths, config keys, product names, protocol
names, error strings, units, file formats), technical verbs (parse,
compile, deploy, mount, encrypt, serialize, and other verbs that name
a computer process), and protected idioms (section 8).

- One name for one thing. Do not alternate between "config file",
  "configuration", and "settings".
- Use the shortest common technical noun. No slang, no jargon.
- Do not use a technical noun as a verb ("to config") or a technical
  verb as a noun ("a deploy") unless the project already does.
- Multi-word nouns: three words maximum. Longer: write it in full or
  hyphenate ("read-only", "command-line").
- American spelling. No contractions. Spell out "for example", "that
  is", "and so on", never "e.g.", "i.e.", "etc.".
- Address the reader as "you", never "we". A third person is "they".
- Define a concept term at its first use, in fewer than ten words
  and in the same part of speech, so that the definition can replace
  the term: "idempotent (safe to run twice)". Do not define product
  names, standard names, or the tool the document is about. Do not
  capitalize the term.
- Abbreviations: the term in full at its first use, then the
  abbreviation in parentheses: "external dynamic list (EDL)", and
  the abbreviation after that. One abbreviation has one meaning. Do
  not coin one, do not introduce one in a heading, and do not
  capitalize the term because the abbreviation is uppercase. Write a
  unit only next to a number, and repeat it in a range: "10 ms to
  20 ms".
- Serial comma: "A, B, and C".
- No parenthetical plurals: "file(s)". Write "files" or "one or
  more files".
- Delete words that carry no fact: simply, just, easily, seamless,
  robust, powerful, comprehensive, leverage, crucial, "in order to",
  "it is worth noting", "is designed to", "aims to", "in conclusion",
  furthermore and moreover (write "also"), "out of the box" (write
  "by default"), "under the hood" (write "internally"). When a word
  claims a quality, give the measurement or delete the word.

Replace these words, verified against the Issue 9 dictionary, only
when the meaning stays the same. If the original carries a hedge or
a permission ("may have failed", "you may skip this step"), keep the
meaning in the new wording, or keep the word and report it.

| Do not write | Write |
|---|---|
| ensure, verify, assure | make sure |
| perform, execute, implement, carry out | do |
| prior to | before |
| utilize | use |
| allow, enable (verb) | let |
| may, might | can |
| should | must |
| via | through |
| obtain | get |
| modify | change |
| provide | give |
| indicate | show |
| commence | start |
| terminate | stop |
| whether | if (except Go "reports whether") |
| attempt (noun) | try (verb) |
| require | rewrite with "must" or "necessary" |

## 2a. Names and identities

Documentation names things, not people or organizations. Never write
the name of the organization that owns the code, a person, a handle,
or an email address, also when the code or the task text has it.
Write "the organization", "the user", a role, or a placeholder:
Cyberdyne Systems (`cyberdyne.example`), then ACME (`acme.example`).
People: Alice, Bob, Carol, Dave. Addresses: the RFC 5737 and RFC 3849
ranges. A vendor or product name stays. Never credit an author.
Read `reference/names.md` before you write an example with a name,
an address, or a domain.

## 2b. Word choice without the dictionary

When you must judge a word that section 2 does not name, read
`reference/word-choice.md`: 12 patterns from the Issue 9 dictionary,
such as keep the part of speech, choose the concrete word for one
meaning, and turn an action noun into its verb.

## 3. Verbs and sentences

- Verb forms: imperative, infinitive, simple present, simple past,
  simple future, past participle as adjective. Exception: keep a
  present perfect that carries a fact the simple form loses ("the
  job has completed" means the output exists now), and report it.
- No "-ing" verb forms. A technical noun can end in "-ing"
  (encoding, logging, routing, load balancing).
- Active voice. Passive only when the agent is unknown.
- Present tense for what the code does now.
- Instructions: 20 words per sentence maximum. Descriptions: 25. A
  number, a number with a unit, an identifier, a code span, a
  hyphenated word, and text in parentheses each count as one word.
- One instruction per sentence. One topic per sentence in a
  description.
- Condition first: "If the file does not exist, Open returns
  ErrNotExist."
- Keep the articles: "the buffer", "a reader". Do not drop words to
  save space.
- No semicolons. STE permits all other standard punctuation.
- No phrasal verbs: "shut down" is "stop", "set up" is "install" or
  "configure", "carry out" is "do".
- Paragraph: one topic, six sentences maximum. The general fact
  first, then detail.
- More than two items or conditions in a sentence: use a vertical
  list.
- List items are parallel: all commands, or all statements, or all
  noun phrases. Test: write the list as one sentence, then break it.
  Casing and punctuation follow from that sentence.
- A list has at least two items at a level. Order steps as the
  reader does them, explanations from general to specific, equal
  items alphabetically.
- To avoid "he or she", make the antecedent plural: "All operators
  must inform their supervisor", not "Each operator must inform
  their supervisor".

## 4. Document structure

Use the man-pages(7) section order, and only the sections that
apply. For README.md the headings are:

    <title> - <one line>       (NAME)
    BLUF                       (section 4c)
    Usage                      (SYNOPSIS)
    Description
    Options / Configuration    (OPTIONS)
    Exit status                (EXIT STATUS)
    Environment
    Files
    Examples
    Caveats, Bugs              (CAVEATS, BUGS)
    See also                   (SEE ALSO)

- NAME is one line: `name - what it does`, lowercase.
- SYNOPSIS notation: bold as-is text, italic replaceable text, `[]`
  optional, `|` choice, `...` repeat.
- DESCRIPTION gives the usual case, and no internals unless the
  reader needs them. Options go in OPTIONS. Mark the version that
  added a flag: `--apply (since v1.4)`.
- EXAMPLES are complete, run without change, show input distinct
  from output, and do error checking.
- CAVEATS lists typical misuse. BUGS lists known defects. SEE ALSO:
  related pages, ordered by section then name, no period.
- Semantic newlines in source: one sentence per line. Split a long
  sentence at a clause boundary.
- Numbers: zero to nine in words, 10 and above in numerals. Numerals
  always for versions, exit codes, ports, counts in tables, and a
  number next to a unit. Do not start a sentence with a numeral. Do
  not write a count twice: "four times", never "four (4) times".
  Hyphenate a count before a noun: "a 10-step procedure". Keep a
  number and its unit on one line.
- Casing: sentence case in text and headings. Capitalize a common
  noun only at the start of a sentence.
- No bold lead-ins, no bold for emphasis, no emoji, no heading over
  fewer than three sentences.
- A vertical list needs three or more parallel items or steps, a
  lead-in that ends with a colon, and one instruction or fact per
  item, which starts uppercase.
- Preferred terms and hyphens: `reference/word-choice.md` gives the
  man-pages terms (filename, hostname, timestamp) and the hyphen
  rules: a compound before a noun is hyphenated, no
  hyphen after multi, non, pre, re, and sub, and the hyphen stays
  before a capital.

## 4a. Text types

Strict mode: error message and CLI output, runbook, commit message
and MR description,
release notes and CHANGELOG, agent instructions, UI copy. Relaxed
mode: incident report. Out of scope: marketing text, announcements,
brand voice, because scraibe deletes persuasion. Say so and offer to
do the docs instead. `reference/text-types.md` gives the pattern of
each type. Read it before you write a text of that type.

## 4b. Interface instructions

For steps in a graphical interface or a web console, read
`reference/interface-instructions.md`: one action per step, the path
before the action, labels in bold as the interface shows them,
placeholders in italics, and the terms click, select, press, type.

## 4c. BLUF

BLUF is the bottom line up front (the conclusion before the detail).
Every prose document starts with it: README.md, each file under
docs/, a design note, a runbook, an incident report. The heading is
`BLUF`, at the level of the other top sections, directly after the
title. A CHANGELOG, a man page source, and a code comment have none.

A BLUF is one paragraph in strict mode, 80 words and six sentences
maximum. In this order: what this is and what it does for the reader
(one sentence), what state it is in (one to three sentences: what
works, what does not, what changed last), and what the reader must
do first (one imperative sentence, omitted when there is no action).

- No list, no code block, no heading, no link. A code span for an
  identifier is permitted.
- Every fact in the BLUF also stands in the body. The BLUF adds no
  fact.
- Write the state, not a version or a date: "The import works. The
  scheduler comes next.", not "Version 0.3 added the import."
- A BLUF goes out of date first, because it states the current
  state. At the start of every task in a repository, before you
  write, compare each sentence of the top-level README.md BLUF with
  the code, the commits since the last tag, and the `[Unreleased]`
  CHANGELOG entries. If a sentence is not true now, or there is no
  BLUF, correct it in the same task, also when the task did not name
  the README. Do the same for each document you change. Report it.

`ste-lint.py` tests the shape of the BLUF, not its truth. Keep the
Usage and Options sections complete: the rules `ghost-command` and
`ghost-flag` compare each command and flag in the prose with them.

## 5. Warnings

Level word first, then the command or condition, then the risk.
WARNING: a person can be injured. CAUTION: data or equipment can be
damaged. NOTE: information only, never an instruction.

    CAUTION: Do not run `apply` against production without
    `--dry-run` first. It writes all rules immediately.

Place a warning directly before the step or paragraph it applies to.
Several at one place: order WARNING, CAUTION, NOTE, and number two
of the same level. A warning is short but complete. In a CHANGELOG
the level word is `Breaking:`.

## 6. Code comments

- A comment says what or why. It does not repeat what the line says.
- Standalone comment: full sentences with a period. Tag comment on
  the code line: short phrase, no period.
- No history in comments. Git has the history.
- No TODO without a reason and a reference: `TODO(#142): reason`.
  Never a person's name or handle as the owner.
- Sections 2 and 3 apply.
- Three questions for each comment: does it say why, not what? Does
  the reader need it to understand the code? Does it give a fact the
  code does not show? Two answers no: delete the comment, or improve
  the code so that it needs none.
- A comment is evergreen: no date, no "now", "recently", "new". A
  constant has its source: a measurement, a limit of a system, a
  requirement, or an incident.

`reference/comments.md` gives the rules per language: Go doc
comments (go.dev/doc/comment, `gofmt` after every change, every
exported name has a comment that starts with the name), Python
docstrings, Bash headers, .NET XML doc. Read it before you write a
comment.

## 7. CHANGELOG

- Format: Keep a Changelog. `## [Unreleased]` at the top.
- Groups: Added, Changed, Deprecated, Removed, Fixed, Security. Omit
  empty groups.
- Entry: one sentence for the operator, not the author. Condition
  first when the change is conditional.
- One entry per user-visible change. An internal refactor with no
  visible effect gets no entry.
- `Breaking:` first, with the migration step.
- Reference an issue or MR number only when it is in the commit.

## 8. Protected idioms

These override the dictionary. Do not replace them.

- Go: "returns", "reports whether", "implements", "provides",
  "panics if", "Deprecated:", "Package x implements ...".
- Any language: "returns" for a function result, "implements" for an
  interface, "raises" for an exception.
- Never rewrite: quoted text, command output, error messages,
  identifiers, config keys, flag names, paths, URLs, product and
  protocol names, units, license text.

## 9. Lint

`reference/ste-lint.py` is a deterministic, stdlib-only linter for
the structural rules. Hard rules fail the run: semicolon,
long-sentence, phrasal-verb, nominalization, marketing-adjective,
filler-word (the delete list of section 2), latin-abbreviation,
contraction, parenthetical-plural, em-dash, numeral-start,
synonym-rotation, dangling-conjunction, bluf-missing, bluf-position,
bluf-shape, bluf-long, ghost-command, ghost-flag, and glossary.
Advisory rules:
passive-voice, present-perfect, progressive-tense, latin-verb, bold,
condition-last, ambiguous-this, heading-case. It never flags hedges
or "should". Confidence is content.

    python3 reference/ste-lint.py [--json] [--baseline N] [--disable RULE,...] [--fix] FILE...

Exit status 1 when hard findings exceed the baseline. Run it on every
Markdown file you touched, and on the comment text of a source file
when you changed comments. `--fix` applies the mechanical fixes. A
directive in an HTML comment disables a rule for a quotation. A
glossary `.scraibe.json` at the repository root gives the preferred
terms, the disabled rules, and the baseline. When a repository has
no `.scraibe.json`, suggest `reference/scraibe-init.sh` to the user
once, and do not run it yourself. `evals/run.sh` runs the evaluation
tasks. `reference/lint.md` gives the details. Read it before you use
them. The linter finds shape, not meaning. A
clean run is not a review.

## 10. AI use

The STEMG white paper on AI (June 2026) gives the direction. It is
not an approval of this standard.

- Every text the agent writes is a draft. A human reviews and commits
  it. The agent never commits.
- Say when AI wrote or changed a text, with a git trailer, not a note
  in the document. The agent gives the line in its report. The human
  adds it to the commit when they accept the draft:

      Doc-Draft: scraibe/0.4

- Keep the decision visible: the diff shows each word the standard
  changed. Do not batch a documentation rewrite into a code commit.
- Do not claim compliance. Write "follows the scraibe standard, which
  comes from ASD-STE100", never "ASD-STE100 compliant" or "STE
  certified". ASD and the STEMG do not endorse a tool, a checker, or
  an AI solution, and the standard gives no such status.
- When a text you write names ASD-STE100, include the statement of
  section 11. `reference/attribution.md` gives the limits of the
  automated checks and the note on confidentiality.

## 11. Attribution

ASD-STE100 Simplified Technical English is a Copyright and a Trade
Mark of ASD, Brussels, Belgium. scraibe names ASD-STE100 as its
source specification. The reference is a factual reference. It does
not imply an endorsement, an approval, a certification, a validation,
or an affiliation by ASD or by the ASD Simplified Technical English
Maintenance Group (STEMG), which did not review this standard.
scraibe does not contain the ASD-STE100 dictionary and does not
replace it. Get an official copy free of charge at
https://www.asd-ste100.org. `reference/attribution.md` gives the full
statement, the copyright rules of ASD, and the sources and licenses
of the rules, the linter, and the hook.

## 12. Review checklist

Check every item before you finish. Eight habits, each points at a
word:

1. Synonym rotation: one thing, several names. Pick one.
2. Hedge stacking: "it is important to note that this can possibly".
   State the claim or delete it. Do not delete a real hedge.
3. Nominalization: "perform an analysis of". Use the verb: "analyze".
4. Marketing adjectives: seamless, robust, powerful. Delete, or give
   the measurement.
5. Run-on sentences. One idea per sentence.
6. Soft phrasal verbs: spin up, reach out, kick off. Use start,
   contact, begin.
7. Sentences that fit any project: "This ensures the system remains
   secure." True of another project without a change, so it carries
   no information. Delete it, or write the mechanism: what the code
   does, when, and at what cost.
8. Significance instead of mechanism: "robust", "secure", "fast".
   Give the check, the limit, or the number behind the claim.

Then run `ste-lint.py`, then grep for what it does not test:
`grep -nE "should|would|may|might|could|, making|\([0-9]+\)" FILE`.
Each hit is a rule break or a named exemption. Then:

- No "-ing" verb outside a technical noun.
- No passive voice with a known agent.
- No word from the replace table in section 2.
- Identifiers, paths, flags, and error strings are verbatim.
- No organization, person, handle, or email in prose. No real
  hostname, domain, or address in an example.
- The reader can do the task with this text only.
- The top-level README.md has a BLUF, and every sentence in it is
  true now.
- Every command in the prose is in Usage, and every flag in Usage or
  Options.
- Every term of `.scraibe.json` is the one the text uses.
- Every directive comment covers a quotation or a name.
- Each comment passes the three questions of section 6.
- No claim got stronger or weaker. No fact got added.
- `ste-lint.py` shows no hard finding above the baseline.
- Go: `gofmt -l` prints nothing, `go vet` passes.
- The report gives the `Doc-Draft:` trailer.
