---
name: scuq-scraibe
description: >
  scraibe v0.1 — the writing standard for all documentation and code
  comments. Load when you write or review a README, docs/, CHANGELOG,
  man page, Go doc comment, Python docstring, .NET XML doc, Bash
  header, or inline comment.
---

# scraibe v0.1

scraibe is one standard for all documentation and comments.
The goal: short text, simple words, no missing facts.

scraibe merges three sources. When they disagree, this file decides.

- ASD-STE100 Issue 9 (2025): words, verbs, sentences, warnings
- man-pages(7): document structure, preferred terms
- go.dev/doc/comment: Go doc comments

## 1. Priority

Apply these in order. A lower rule never breaks a higher one.

1. Correct. The text says what the code does. Read the code first.
   Never add a fact the code does not show.
2. Complete. The reader can do the task with this text only.
   Never remove a necessary fact to make the text shorter.
3. Same claim. Never change the strength of a statement. A rewrite
   that turns a possibility into a fact is a different claim, not a
   shorter one. Say the strength with STE words: "can" for a
   possibility, "must" for a requirement, "unknown" for an unknown
   cause, a stated condition for a case. Never should, would, may,
   might, could. A required "should" becomes "must". An optional
   "should" is deleted. A "may have failed" becomes "possibly failed"
   or "failed in some cases", not "failed".
4. Short. Remove every word that carries no information.
5. Style. All rules below.

## 1a. Two kinds of rules, two modes

Structural rules describe the shape of a sentence. You can apply
them without the STE dictionary, and `ste-lint.py` can test most of
them. Lexical rules depend on the dictionary, which this standard
does not contain. Apply them as a direction, not as a test, and do
not claim dictionary compliance.

Strict mode: comments, docstrings, error messages, CLI help, man
pages, CHANGELOG entries, warnings, procedures. Apply every rule.

Relaxed mode: README narrative, docs/ explanations, design notes.
Apply all structural rules (section 3). Treat the replace table in
section 2 as advice. Prose needs some range.

When the user does not name a mode, pick one from the text type.

## 2. Words

You can use only these four groups of words:

- Approved STE words (see section 9 to check a word)
- Technical nouns: identifiers, flags, paths, config keys, product
  names, protocol names, error strings, units, file formats
- Technical verbs: parse, compile, build, deploy, commit, push, mount,
  encrypt, decode, serialize, route, log, and other verbs that name a
  computer process
- Protected idioms (section 8)

Rules:

- One name for one thing. Do not alternate between "config file",
  "configuration", and "settings".
- Use the shortest common technical noun. No slang, no jargon.
- Do not use a technical noun as a verb ("to config") or a technical
  verb as a noun ("a deploy") unless the project already does.
- Multi-word nouns: three words maximum. Longer: write it in full or
  hyphenate ("read-only", "command-line").
- American spelling.
- No contractions.
- Spell out "for example", "that is", "and so on".
  Never "e.g.", "i.e.", "etc.".
- Address the reader as "you". Never "we". For a third person use
  "they".
- Define a concept term at its first use, in fewer than ten words:
  "idempotent (safe to run twice)". Do not define product names,
  standard names, or the tool the document is about. Write the
  definition in the same part of speech as the term, so it can
  replace the term in a sentence. Do not capitalize the term.
- Abbreviations: write the term in full at its first use, then the
  abbreviation in parentheses: "external dynamic list (EDL)". Use
  the abbreviation after that. One abbreviation has one meaning in a
  document. Do not coin new abbreviations. Do not introduce an
  abbreviation in a heading. Do not capitalize a term because its
  abbreviation is uppercase: "ground support equipment (GSE)".
  Write a unit abbreviation only next to a number, and repeat it in
  a range: "10 ms to 20 ms".
- Serial comma: "A, B, and C". In technical text the last comma
  removes an ambiguity.
- No parenthetical plurals: "file(s)". Write "files" or "one or
  more files".
- Delete words that carry no fact. A short list of the usual ones:
  simply, just, easily, seamless, robust, powerful, comprehensive,
  leverage, crucial, "in order to", "it is worth noting", "is
  designed to", "aims to", "out of the box" (write "by default"),
  "under the hood" (write "internally"), furthermore, moreover
  (write "also"), "in conclusion". When a word claims a quality,
  give the measurement or delete the word.

Replace these words. All verified against the Issue 9 dictionary.
Replace a word only when the meaning stays the same. If the original
word carries a hedge or a permission ("may have failed", "you may
skip this step"), keep the meaning in the new wording or keep the
original word and report it.

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

Documentation names things, not people or organizations.

- Never write the name of the organization that owns the code, its
  departments, or its internal teams. Write "the organization", "the
  operator", or the placeholder Cyberdyne Systems. This holds even when the name appears in
  the code, in git history, in hostnames, in a config file, or in
  the task text from the user or from another agent.
- Never write a real username, handle, email address, or personal
  name. Write "the user", or a role ("the administrator"), or a
  placeholder from the list below.
- Never credit an author. No AUTHORS section, no "written by", no
  name in a changelog entry or a comment. Git holds the history.
- The default organization in all examples is Cyberdyne Systems,
  domain `cyberdyne.example`. When a second organization is needed,
  it is ACME, domain `acme.example`. Use them in that order. Do not
  invent a third. The `.example` top-level domain is reserved (RFC
  2606) and never resolves. Do not use a live top-level domain for a
  placeholder, because someone can register it.
- People: Alice, Bob, Carol, Dave in that order. For security text:
  Eve (eavesdropper), Mallory (attacker), Trent (trusted third
  party). Addresses: `alice@cyberdyne.example`.
- Addresses: `192.0.2.0/24`, `198.51.100.0/24`, `203.0.113.0/24`
  (RFC 5737), `2001:db8::/32` (RFC 3849), MAC `00-53-00-xx-xx-xx`
  (RFC 7042). Hostnames: `fw01.cyberdyne.example`.
- Never a real hostname, domain, address, or organization name from
  the code, unless the user asks for the real value in that document.
- In certificate and ACME-protocol documents, do not use ACME as an
  organization. Use Cyberdyne Systems only.
- Vendor and product names are technical names and stay: Cisco,
  Palo Alto Networks, Debian, PostgreSQL. The line is: the name of a
  thing you can buy or download stays. The name of a person or of
  the organization that runs the code does not.
- When an identifier itself contains a person's name or the
  organization's name (a config key, a function name, a path), you
  can write the identifier verbatim in a code span, because the code
  is the truth. Do not repeat the name in prose around it.

## 3. Verbs and sentences

- Use only these verb forms: imperative, infinitive, simple present,
  simple past, simple future, past participle as adjective.
  Exception: keep a present perfect when it carries a fact the simple
  form loses ("the job has completed" means the output exists now).
  Report the exception.
- No "-ing" verb forms. A technical noun can end in "-ing"
  (encoding, logging, routing, load balancing).
- Active voice. Passive only when the agent is unknown.
- Present tense for what the code does now.
- Instructions: 20 words per sentence maximum.
  Descriptions: 25 words per sentence maximum.
  Count each of these as one word: a number, a number with a unit, an
  identifier, a code span, a hyphenated word, text in parentheses.
- One instruction per sentence. One topic per sentence in descriptions.
- Write the condition first:
  "If the file does not exist, Open returns ErrNotExist."
- Keep the articles: "the buffer", "a reader". Do not drop words to
  save space.
- No semicolons. STE permits all other standard punctuation.
- No phrasal verbs. "shut down" → "stop". "set up" → "install" or
  "configure". "carry out" → "do".
- Paragraph: one topic, six sentences maximum.
- Give information gradually: the general fact first, then detail.
- When a sentence has more than two items or conditions, use a
  vertical list.
- Items in a list or a series are parallel: all commands, or all
  statements, or all noun phrases. Test: write the list as one
  sentence first, then break it. The casing and punctuation follow
  from that sentence.
- A list has at least two items at a level. Order steps in the order
  the reader does them. Order explanations from general to specific.
  Order equal items alphabetically.
- To avoid "he or she", make the antecedent plural: "All operators
  must inform their supervisor", not "Each operator must inform
  their supervisor".

## 4. Document structure

Use the man-pages(7) section order. Use only the sections that apply.

    NAME, SYNOPSIS, DESCRIPTION, OPTIONS, EXIT STATUS, ENVIRONMENT,
    FILES, EXAMPLES, CAVEATS, BUGS, SEE ALSO

For README.md the same order with these headings:

    <title> - <one line>       (NAME)
    Usage                      (SYNOPSIS)
    Description
    Options / Configuration
    Exit status
    Environment
    Files
    Examples
    See also

Rules:

- NAME is one line: `name - what it does`, lowercase.
- SYNOPSIS notation: bold for as-is text, italic for replaceable
  text, `[]` optional, `|` choice, `...` repeat.
- DESCRIPTION gives the usual case. Options go in OPTIONS.
  Do not describe internals unless the reader needs them.
- Mark the version that added a flag: `--apply (since v1.4)`.
- EXAMPLES are complete and run without change. Show user input
  distinct from output. Examples do error checking.
- Semantic newlines in source: one sentence per line. Split a long
  sentence at a clause boundary.
- CAVEATS lists typical misuse. BUGS lists known defects.
- Numbers: write zero to nine in words and 10 and above in numerals
  in prose. Numerals always for versions, exit codes, ports, counts
  in tables, and any number next to a unit. Do not start a sentence
  with a numeral. Do not write a count twice: "four times", never
  "four (4) times". Hyphenate a count that modifies a noun: "a
  10-step procedure". Keep the number and its unit on one line.
- Casing: sentence case in running text and in headings. Capitalize
  a common noun only at the start of a sentence.
- No bold lead-ins and no bold for emphasis. No emoji. No heading
  over fewer than three sentences.
- A vertical list needs three or more parallel items or steps. The
  lead-in ends with a colon. Each item starts uppercase and holds
  one instruction or one fact.
- SEE ALSO: related pages, ordered by section then name, no period.

Preferred terms (man-pages): filename, filesystem, hostname, pathname,
timestamp, timezone, username, superuser, symbolic link, nonzero,
lowercase, uppercase, user space, run time (noun), run-time (adjective).

Avoid: "OS" → "operating system", "tty" → "terminal",
"non-root" → "unprivileged user", "manpage" → "man page".

Hyphenate a compound before a noun: "command-line argument",
"read-only file". No hyphen after multi, non, pre, re, sub:
nonblocking, subdirectory, reinitialize. Keep the hyphen before a
capital: non-ASCII.

## 4a. Text types

Each type names its mode and its pattern.

- Error message and CLI output. Strict. Say what happened (simple
  past). Say the cause if known. Give the command or condition that
  corrects it. "Connection to the database failed. The password for
  user `app` was not correct. Set `DB_PASSWORD` and connect again."
- Runbook. Strict, 20-word limit, one instruction per step,
  condition first. A warning stands before its step.
- Incident report. Relaxed, simple past only. No "has identified", no
  "may have impacted". State what is known. Write "unknown" for the
  rest. Give times in UTC.
- Commit message and MR description. Imperative subject line, 25-word
  limit in the body. Delete "this MR aims to".
- Release notes and CHANGELOG. Section 7. A `Breaking:` entry follows
  the warning pattern, command first: "Update your calls to
  `/v2/users`. The `name` field split into `first_name` and
  `last_name`."
- Agent instructions (CLAUDE.md, AGENTS.md, subagents, skills).
  Strict. One instruction per sentence. One word for one action. A
  condition first. Never "should": a model reads it as optional.
- UI copy. Strict, hard length limits. Buttons and labels are
  technical names and are exempt.
- Marketing text, launch posts, brand voice. Out of scope. scraibe
  deletes persuasion. Say so and offer to do the docs instead.

## 4b. Interface instructions

For steps in a graphical interface or a web console:

- Give steps in the order the reader does them. One action per step.
  Put multi-step operations in a numbered list.
- Give the path first, then the action: "Click the Objects tab, and
  then click Add", not "Click Add on the Objects tab".
- Write labels of windows, tabs, buttons, fields, and options exactly
  as the interface shows them, in bold. Keep box, list, check box,
  and tab as descriptors. Drop other descriptors when the step is
  clear.
- Placeholders for input in italics: "Type *hostname*."
- Keep explanations and notes out of the step list. Put them before
  or after the list.
- When there are two ways to do one thing, separate them so the
  reader cannot read them as two consecutive steps.
- Mouse instructions before keyboard instructions.
- Terms: click or select (not click on, choose, pick). Clear a check
  box (not deselect). Press a key. Type input (not enter). Point to
  a submenu. Available and unavailable (not active, grayed out).
  Dialog box. Double-click to open. Right-click for a context menu.
  "Under *Name*" for a group in a dialog box.

## 5. Warnings

Put the level word first. Then the command or condition. Then the risk.

- WARNING: a person can be injured.
- CAUTION: data or equipment can be damaged.
- NOTE: information only. A note never gives an instruction.

Example:

    CAUTION: Do not run `apply` against production without
    `--dry-run` first. It writes all rules immediately.

Place a warning directly before the step or paragraph it applies
to. When several apply to the same place, order them WARNING,
CAUTION, NOTE. When two of the same level apply, number them. A
warning is short but complete.

In a CHANGELOG the level word is `Breaking:`.

## 6. Code comments

Common to all languages:

- A comment says what or why. It does not repeat what the line says.
- Standalone comment: full sentences, ends with a period.
- Tag comment (same line as code): short phrase, no period.
- No history in comments. Git has the history.
- No TODO without a reason and a reference: `TODO(#142): reason`.
  Never a person's name or handle as the owner (section 2a).
- All rules from sections 2 and 3 apply.

### Go

Follow go.dev/doc/comment. Run `gofmt` after every change.

- Every exported name has a doc comment. The comment starts with the
  name: "Package netbox ...", "A Client ...", "Fetch returns ...".
- Command: the package comment starts with the program name and
  describes the behavior. Then a `Usage:` block, indented.
- Func: say what it returns, or what it does when called for the
  side effect. Boolean result: "reports whether". Name the results
  when the comment refers to them.
- Type: say what one instance represents. State concurrency safety
  if it is stronger than "one goroutine at a time". State the zero
  value meaning if it is not obvious. Document every exported field,
  in the type comment or per field.
- Special cases: list them explicitly, one per line, indented.
- Do not describe the algorithm in the doc comment. Put that inside
  the function body.
- Deprecation: a paragraph `Deprecated: <reason>. Use <X> instead.`
- Links: `[Name]`, `[pkg.Name]`, `[*pkg.Type]`. URL targets go at the
  end as `[Text]: URL`.
- Syntax: blank `//` line between paragraphs. `# Heading` unindented
  with blank lines around it. Lists indented with `-` or `1.`. Code
  blocks indented one tab. No nested lists. Do not indent a wrapped
  prose line.

### Python

- Docstring on every public module, class, and function.
- First line: one sentence, imperative, ends with a period.
  Blank line. Then details.
- Sections `Args:`, `Returns:`, `Raises:` when the function has them.

### Bash

- Header block: name, one-line purpose, usage line, exit codes.
- One comment line before each function: what it does.

### .NET

- `<summary>` is one sentence. Add `<param>`, `<returns>`,
  `<exception>` when they exist.

## 7. CHANGELOG

- Format: Keep a Changelog. `## [Unreleased]` at the top.
- Groups: Added, Changed, Deprecated, Removed, Fixed, Security.
  Omit empty groups.
- Entry: one sentence for the operator, not the author.
  Condition first when the change is conditional.
- One entry per user-visible change. Internal refactors with no
  visible effect do not get an entry.
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

`reference/ste-lint.py` is a deterministic, stdlib-only linter for the
structural rules. It tests: semicolons, sentence length, soft phrasal
verbs, nominalization, marketing adjectives, synonym rotation, list
items that end in "and" or "or", and, as advice only, passive voice
and compound tenses. It never flags hedges. Confidence is content.

    python3 reference/ste-lint.py FILE...
    python3 reference/ste-lint.py --json FILE
    python3 reference/ste-lint.py --baseline 20 FILE   # adopt on old docs
    python3 reference/ste-lint.py --disable synonym-rotation FILE

Exit status 1 when hard findings exceed the baseline. Run it on every
Markdown file you touched. Run it on the extracted comment text of a
source file when you changed comments.

The linter finds shape, not meaning. A clean run is not a review.

## 10. AI use

The STEMG white paper on AI (June 2026) sets the frame. scraibe obeys it.

- Every text the agent writes is a draft. A human reviews it and
  commits it. The agent never commits.
- Say when AI wrote or changed a text. Use a git trailer, not a note
  in the document:

      Doc-Draft: scraibe/0.1

  The agent gives this line in its report. The human adds it to the
  commit when they accept the draft.
- Keep the decision visible. When the agent changed a word because of
  the standard, the diff shows it. Do not batch a documentation
  rewrite into a code commit.
- Do not claim compliance. Write "follows the scraibe standard, which
  comes from ASD-STE100". Never write "ASD-STE100 compliant" or
  "STE certified". ASD does not endorse AI tools, and the standard
  gives no such status.
- Automated checks have limits. `ste-lint.py` tests sentence shape.
  It does not test meaning or dictionary compliance. The reviewer does.
- Confidentiality is a policy decision of the organization, not of
  this standard. Run the agent only where your policy lets the source
  and its documentation go.

## 11. Attribution

scraibe is a derived guideline. It does not contain the ASD-STE100
dictionary and does not replace it.

ASD holds the copyright of ASD-STE100 Simplified Technical English
(Aerospace, Security and Defence Industries Association of Europe,
Brussels). ASD-STE100 is a registered EU trademark of ASD. Get an
official copy free of charge at https://www.asd-ste100.org. Do not
redistribute the specification or the dictionary.

`reference/ste-lint.py` is from github.com/danyuchn/asd-ste100-skill,
© 2026 Dustin Yuchen Teng, MIT license. See `reference/LICENSE-ste-lint`.
The mode split, the claim-strength rule, and the six scan habits in
this standard also come from that project.

The text-type patterns (section 4a), the delete list (section 2), the
formatting rules for lists and headings, and the lint hook are from
github.com/AminBlg/SimpleEnglish, © 2026 AminBlg, MIT license.

The rules for abbreviations, numbers, casing, parallel lists, the
serial comma, warning placement, and the interface instructions
(section 4b) are from NASA KSC-DF-107 Revision F, Technical
Documentation Style Guide (2017), a United States Government work
approved for public release with unlimited distribution.

man-pages(7) is part of the Linux man-pages project.
Go doc comment conventions are from https://go.dev/doc/comment.

## 12. Review checklist

Check every item before you finish.

Six habits that make text hard to parse. Each one points at a word:

1. Synonym rotation: one thing has several names. Pick one.
2. Hedge stacking: "it is important to note that this can possibly".
   State the claim, or delete it. Do not delete a real hedge.
3. Nominalization: "perform an analysis of". Use the verb: "analyze".
4. Marketing adjectives: seamless, robust, powerful. Delete, or give
   the measurement.
5. Run-on sentences. One idea per sentence.
6. Soft phrasal verbs: spin up, reach out, kick off. Use the plain
   verb: start, contact, begin.

Then a grep. Each hit is a rule break or an exemption you can name:

    grep -nE "should|would|may|might|could|has been|have been|;|—|, making|\*\*|'|\(s\)|\([0-9]+\)" FILE

Then the rules:

- Every sentence is within the word limit.
- No semicolon, no contraction, no "e.g.", "i.e.", "etc.".
- No "-ing" verb outside a technical noun.
- No passive voice with a known agent.
- No word from the replace table in section 2.
- Identifiers, paths, flags, and error strings are verbatim.
- No organization name, person's name, handle, or email in prose.
  No real hostname, domain, or address in an example.
- The reader can do the task with this text only.
- No claim got stronger or weaker. No fact got added.
- `ste-lint.py` shows no hard finding above the baseline.
- Go: `gofmt -l` shows no file. `go vet` passes.
- The report gives the `Doc-Draft:` trailer.
