#!/usr/bin/env python3
"""Deterministic linter for the structural STE rules in SKILL.md.

Checks only rules verifiable without ASD's dictionary. Deliberately never
flags hedges or modality (may/might/could): the skill treats confidence as
content, and a linter that pressures hedges out would rewrite claims.

The BLUF rules (bluf-missing, bluf-position, bluf-shape, bluf-long) test
the shape of the BLUF section of SKILL.md section 4c. They cannot test if
the BLUF is true.

The ghost-command rule compares each command in prose and in shell code
fences with the Usage or Synopsis section. A command that the section does
not have is a hard finding. The ghost-flag rule does the same for each
flag, against the Usage, Synopsis, Options, and Configuration sections.
Both rules test presence, not behavior.

A repository can have a glossary .scraibe.json at its root, found from the
directory of the file upward:
    {"terms": {"config file": ["configuration file"]}, "disable": ["bold"], "baseline": 0}
"terms" maps a preferred term to the words that the rule glossary flags.
"disable" names rules the repository does not use. "baseline" is the
default for --baseline.

--fix changes semicolons to periods, deletes the marketing adjectives and
the fillers "simply" and "easily", expands the contractions in -n't, and
replaces the phrasal verbs of PHRASAL_FIXES and the phrases of
PHRASE_FIXES. It changes nothing else, and it does not touch code fences
or inline code. Read the diff after a run.

A directive comment disables a rule for a line, a block, or a file:
    <!-- ste-lint: disable-next-line semicolon,filler-word -->
    <!-- ste-lint: disable-line em-dash -->
    <!-- ste-lint: disable marketing-adjective --> ... <!-- ste-lint: enable marketing-adjective -->
    <!-- ste-lint: disable-file bold -->

Usage:
    ste-lint.py FILE [FILE ...]
    echo "text" | ste-lint.py [--json]
    ste-lint.py --baseline 5 FILE      # pass unless hard violations exceed 5
    ste-lint.py --disable passive-voice,present-perfect FILE
    ste-lint.py --fix FILE             # apply the mechanical fixes, then lint
    ste-lint.py --no-glossary FILE     # ignore the .scraibe.json of the repository
    echo "text" | ste-lint.py --fix    # print the fixed text
    ste-lint.py --selftest

Exit 1 when hard ("advisory-free") violations exceed the baseline (default 0).
Advisory findings (passive voice, compound tenses, Latin verb suffixes, bold,
progressive tense, a condition after the instruction, "This" at the start of
a sentence, title case in a heading) never fail the run.

ASD-STE100 Simplified Technical English is a Copyright and a Trade Mark of ASD,
Brussels, Belgium. ASD and the STEMG do not endorse this linter.
"""
import json
import os
import re
import sys

# ponytail: regex heuristics, not a parser. No noun-cluster rule — needs POS
# tagging to avoid constant false positives; add spaCy-backed rule if ever needed.
# No ellipsis rule by owner's choice: technical writing sometimes earns one.
RULES = [
    ("semicolon", "advisory-free",
     re.compile(r";"),
     "STE bans the semicolon (Rule 8.1). Split into separate sentences."),
    ("phrasal-verb", "advisory-free",
     re.compile(r"\b(spin(?:ning|s)? up|spun up|reach(?:ing|es|ed)? out|div(?:e|es|ing|ed) into|dove into|kick(?:ing|s|ed)? off|circl(?:e|es|ing|ed) back|touch(?:ing|es|ed)? base)\b", re.I),
     "Soft phrasal verb. Use the single plain verb (start, contact, read, begin)."),
    ("marketing-adjective", "advisory-free",
     re.compile(r"\b(seamless(?:ly)?|robust(?:ly)?|cutting-edge|effortless(?:ly)?|blazing[- ]fast|world-class|state-of-the-art|game-chang(?:ing|er))\b", re.I),
     "Marketing adjective. Delete, or replace with the measurement that earns the claim."),
    ("nominalization", "advisory-free",
     re.compile(r"\b(perform|performs|performed|conduct|conducts|conducted|carry out|carries out|carried out)\s+(?:a|an|the)\s+\w+(?:tion|sion|ment|ance|ence|ysis)\b", re.I),
     "Action frozen into a noun. Use the verb (analyze, not perform an analysis of)."),
    ("passive-voice", "advisory",
     re.compile(r"\b(is|are|was|were|been|being)\s+(\w+ed|given|taken|made|done|found|seen|known|shown|written|built|sent|set|run|read|kept|held|left|put)\b(?!\s+(?:to|for|by)\s+\w+ing)", re.I),
     "Possible passive voice. Name the actor and use an active verb, unless the actor is unknown or irrelevant."),
    ("latin-verb", "advisory",
     # -ize/-ise/-ify verbs: about four of five such dictionary words are not
     # approved. Non-verb spellings that share the ending are excluded.
     re.compile(r"\b(?!(?:otherwise|likewise|clockwise|counterclockwise|anticlockwise|precise|concise|premise|premises|promise|promises|exercise|exercises|expertise|enterprise|enterprises|franchise|merchandise|surprise|surprises|sunrise|cruise|bruise|disguise|paradise|treatise|demise|reprise)\b)[a-z]{3,}(?:iz|is)(?:e|es|ed|ing)\b|\b[a-z]{3,}if(?:y|ies|ied|ying)\b", re.I),
     "Verb with a Latin suffix. Most of these are not approved in STE. Use the plain verb, unless the word is a technical verb."),
    ("latin-abbreviation", "advisory-free",
     re.compile(r"\b(?:e\.g\.|i\.e\.|etc\.?)(?=[\s,;:)]|$)", re.I),
     'Latin abbreviation. Write "for example", "that is", or "and so on".'),
    ("contraction", "advisory-free",
     re.compile(r"(?!Let['\u2019]s Encrypt)\b(?:\w+n['\u2019]t|it['\u2019]s|that['\u2019]s|there['\u2019]s|here['\u2019]s|what['\u2019]s|let['\u2019]s|who['\u2019]s|where['\u2019]s|how['\u2019]s|\w+['\u2019]re|\w+['\u2019]ve|\w+['\u2019]ll|(?:I|you|we|they)['\u2019]d|I['\u2019]m)\b", re.I),
     "Contraction. Write the two words."),
    ("parenthetical-plural", "advisory-free",
     re.compile(r"\w\(s\)"),
     'Parenthetical plural. Write the plural, or "one or more".'),
    ("em-dash", "advisory-free",
     re.compile(r"[\u2014\u2013]"),
     "Em or en dash. Use a comma, a colon, or a period."),
    ("filler-word", "advisory-free",
     re.compile(r"\b(?:simply|just|easily|powerful|comprehensive|leverages?|leveraging|crucial|furthermore|moreover)\b|\bin order to\b|\bit is worth noting\b|\bis designed to\b|\baims? to\b|\bout of the box\b|\bunder the hood\b|\bin conclusion\b", re.I),
     "Word without a fact (SKILL.md section 2 delete list). Delete it, or give the measurement."),
    ("bold", "advisory",
     re.compile(r"\*\*[^*\n]+\*\*"),
     "Bold. Use it only for an interface label (section 4b). No lead-in, no emphasis."),
    ("progressive-tense", "advisory",
     re.compile(r"\b(?:am|is|are|was|were|be|been|being)\s+(?!(?:nothing|something|anything|everything|morning|evening|during|according|missing|following|existing|remaining|pending|outstanding|interesting|willing|thing|string|king|ring|spring|ceiling|building|encoding|logging|routing)\b)[a-z]+ing\b", re.I),
     'Possible "-ing" verb form. Use the simple present or past, unless the word is a technical noun.'),
    ("present-perfect", "advisory",
     # modal + perfect infinitive ("may have failed") is a protected hedge, not present perfect
     re.compile(r"(?<!\bmay )(?<!\bmight )(?<!\bcould )(?<!\bshould )(?<!\bwould )(?<!\bmust )\b(has|have|had)\s+(?:been\s+)?\w+(?:ed|en)\b", re.I),
     "Compound tense. Use simple past/present unless current relevance is the point (then keep and flag)."),
]

# One word, one meaning: groups of verbs commonly rotated for the same action.
# Only pairs where the members are genuinely interchangeable — error/fault/failure
# are distinct concepts and stay out.
SYNONYM_GROUPS = [
    ("check", "verify", "confirm", "validate"),
    ("delete", "remove", "erase"),
    ("start", "launch", "begin", "initiate"),
    ("stop", "halt", "terminate"),
    ("show", "display"),
    ("use", "utilize", "employ"),
    ("fix", "repair", "correct"),
    ("send", "transmit"),
    ("get", "retrieve", "fetch", "obtain"),
    ("change", "modify", "alter"),
]

MAX_WORDS = 25  # descriptions cap; instructions cap is 20 but undetectable without context

CODE_FENCE = re.compile(r"^(```|~~~)")
INLINE_CODE = re.compile(r"`[^`]*`")
LIST_ITEM_START = re.compile(
    r"^(?P<indent> {0,3})(?P<marker>[-*+]|[0-9]+[.)])(?P<gap> +)(?P<body>.*)$"
)
CONJUNCTION_END = re.compile(r"\b(?:and|or)\s*$", re.I)
TABLE_SEPARATOR_CELL = re.compile(r"^:?-{3,}:?$")
HEADING = re.compile(r"^(?P<level>#{1,6})\s+(?P<text>.*?)\s*#*\s*$")
README_NAMES = {"readme.md"}
BLUF_MAX_WORDS = 80
BLUF_MAX_SENTENCES = 6
USAGE_HEADINGS = {"usage", "synopsis"}
COMMAND_WORD = re.compile(r"^[a-z][a-z0-9._-]*$")
SHELL_FENCE = re.compile(r"^(```|~~~)\s*(sh|bash|shell|zsh|console)\b", re.I)
SHELL_SPLIT = re.compile(r"\s*(?:&&|\|\||\||;)\s*")

# --fix: the mechanical rewrites. Each is (pattern, replacement). A
# replacement can be a function. The fixes run outside code fences and
# outside inline code. They change nothing else.
PHRASAL_FIXES = {
    "spin up": "start", "spins up": "starts", "spinning up": "starting",
    "spun up": "started",
    "reach out": "contact", "reaches out": "contacts",
    "reaching out": "contacting", "reached out": "contacted",
    "kick off": "start", "kicks off": "starts", "kicking off": "starting",
    "kicked off": "started",
    "dive into": "read", "dives into": "reads", "diving into": "reading",
    "dived into": "read", "dove into": "read",
    "circle back": "return", "circles back": "returns",
    "circling back": "returning", "circled back": "returned",
    "touch base": "contact", "touches base": "contacts",
    "touching base": "contacting", "touched base": "contacted",
}
PHRASE_FIXES = {
    "in order to": "to", "under the hood": "internally",
    "out of the box": "by default", "furthermore": "also", "moreover": "also",
    "e.g.": "for example", "i.e.": "that is",
}
HTML_ENTITY = re.compile(r"&[#A-Za-z0-9]+;")
HTML_COMMENT = re.compile(r"<!--.*?-->")
DIRECTIVE = re.compile(
    r"<!--\s*ste-lint:\s*(?P<kind>disable-next-line|disable-line|disable-file|disable|enable)\b(?P<rules>[\w\s,-]*)-->"
)
IMPERATIVE = re.compile(
    r"^(?:Run|Set|Use|Open|Click|Type|Install|Start|Stop|Add|Remove|Delete|Create|Write|Edit|Check|Make|Press|"
    r"Select|Enter|Change|Copy|Move|Restart|Enable|Disable|Configure|Apply|Commit|Push|Pull|Call|Pass|Put|Send|"
    r"Read|Save|Close|Build|Deploy|Mount|Update|Upgrade|Rename|Replace|Specify|Provide|Verify|Do|Give|Keep|"
    r"Leave|Omit|Export|Import|Load|Return|Wait|Repeat|Store|Print|Define|Name|List|Show|Test|Clear|Reset|Turn)\b"
)
CONDITION = re.compile(r"\b(?:if|when|unless)\b", re.I)
AMBIGUOUS_THIS = re.compile(
    r"^This (?:is|was|ensures|means|lets|makes|causes|gives|returns|helps|allows|enables|happens|results|requires|"
    r"prevents|keeps|works|runs|avoids|adds|removes|creates|uses|does|has|can|must|will|should|may|might|could)\b"
)
CAPITALIZED_WORD = re.compile(r"^[A-Z][a-z]+[,:]?$")
INDENTED_CODE = re.compile(r"^(?: {4}|\t)")
OPTIONS_HEADINGS = {"options", "configuration"}
FLAG = re.compile(r"(?<![\w-])(--?[A-Za-z][\w-]*)")
GLOSSARY_NAME = ".scraibe.json"


def _word_re(base):
    return re.compile(r"\b" + base + r"(?:s|es|ed|d|ing)?\b", re.I)


def _leading_spaces(line):
    return len(line) - len(line.lstrip(" "))


def _is_list_continuation(line, content_indent):
    if not line.strip():
        return True
    if LIST_ITEM_START.match(line):
        return False
    return _leading_spaces(line) >= content_indent


def _split_table_row(line):
    """Return trimmed table cells and their zero-based source columns.

    A pipe must separate at least two cells. Escaped pipes stay in their cell.
    This deliberately implements only the ordinary Markdown table shape; it is
    enough to distinguish a table from prose that happens to contain a pipe.
    """
    left = len(line) - len(line.lstrip())
    right = len(line.rstrip())
    content = line[left:right]
    if "|" not in content:
        return None
    if content.startswith("|"):
        content = content[1:]
        left += 1
    if content.endswith("|"):
        content = content[:-1]
    raw_cells = re.split(r"(?<!\\)\|", content)
    if len(raw_cells) < 2:
        return None

    cells = []
    column = left
    for raw_cell in raw_cells:
        leading = len(raw_cell) - len(raw_cell.lstrip())
        cells.append((raw_cell.strip(), column + leading))
        column += len(raw_cell) + 1
    return cells


def _markdown_table_cells(lines):
    """Map ordinary Markdown table rows to their prose cells.

    The separator row anchors detection, so pipe-containing prose is not
    treated as a table. Both leading-pipe and no-leading-pipe table styles are
    accepted when their header and body use the same number of cells.
    """
    table_cells = {}
    index = 1
    while index < len(lines):
        separator = _split_table_row(lines[index])
        header = _split_table_row(lines[index - 1])
        if (not separator or not header or len(separator) != len(header)
                or not all(TABLE_SEPARATOR_CELL.fullmatch(cell)
                           for cell, _ in separator)):
            index += 1
            continue

        table_cells[index - 1] = header
        table_cells[index] = []
        index += 1
        while index < len(lines):
            row = _split_table_row(lines[index])
            if not row or len(row) != len(separator):
                break
            table_cells[index] = row
            index += 1
    return table_cells


def _dangling_conjunction_findings(text, filename):
    lines = text.splitlines()
    findings = []
    in_fence = False
    index = 0
    while index < len(lines):
        line = lines[index]
        stripped = line.strip()
        if CODE_FENCE.match(stripped):
            in_fence = not in_fence
            index += 1
            continue
        if in_fence:
            index += 1
            continue
        start = LIST_ITEM_START.match(line)
        if not start:
            index += 1
            continue

        content_indent = (len(start.group("indent"))
                          + len(start.group("marker"))
                          + len(start.group("gap")))
        item_lines = [(index, start.group("body"))]
        next_index = index + 1
        item_fence = False
        while next_index < len(lines):
            candidate = lines[next_index]
            candidate_stripped = candidate.strip()
            if CODE_FENCE.match(candidate_stripped):
                # Fence delimiters are state markers, not meaningful item lines.
                item_fence = not item_fence
                next_index += 1
                continue
            if item_fence:
                next_index += 1
                continue
            if not _is_list_continuation(candidate, content_indent):
                break
            item_lines.append((next_index, candidate))
            next_index += 1

        meaningful = []
        for line_index, item_line in item_lines:
            # Preserve code spans as neutral operands while ignoring their contents.
            cleaned = INLINE_CODE.sub(" CODE ", item_line).strip()
            if cleaned:
                meaningful.append((line_index, cleaned))
        if meaningful:
            end_line_index, end_line = meaningful[-1]
            conjunction = CONJUNCTION_END.search(end_line)
        else:
            end_line_index, end_line, conjunction = None, None, None
        if conjunction:
            if end_line_index == index:
                finding_line = index + 1
                finding_col = start.start("marker") + 1
            else:
                raw_end_line = next(
                    raw for line_index, raw in item_lines
                    if line_index == end_line_index
                )
                masked_end_line = INLINE_CODE.sub(
                    lambda match: " " * len(match.group(0)), raw_end_line
                )
                raw_conjunction = CONJUNCTION_END.search(masked_end_line)
                finding_line = end_line_index + 1
                finding_col = raw_conjunction.start() + 1 if raw_conjunction else 1
            findings.append({
                "file": filename,
                "line": finding_line,
                "col": finding_col,
                "rule": "dangling-conjunction",
                "level": "advisory-free",
                "match": end_line,
                "message": "List item ends with a coordinating conjunction. Complete the item or join it with the next item.",
            })
        index = next_index
    return findings


def _code_lines(lines):
    """Return the 0-based indices of the lines that are code.

    Code is a fenced block, with its fence lines, or an indented block: a
    line indented by four spaces or a tab after a blank line, when the
    text before the blank line is not a list. A list continues in an
    indented paragraph, so that paragraph is prose.
    """
    code = set()
    in_fence = False
    in_indented = False
    previous_blank = True
    in_list = False
    for index, raw in enumerate(lines):
        stripped = raw.strip()
        if CODE_FENCE.match(stripped):
            in_fence = not in_fence
            in_indented = False
            previous_blank = False
            in_list = False
            code.add(index)
            continue
        if in_fence:
            code.add(index)
            continue
        if not stripped:
            previous_blank = True
            continue
        indented = bool(INDENTED_CODE.match(raw))
        if in_indented and indented:
            code.add(index)
            continue
        in_indented = False
        if indented and previous_blank and not in_list:
            in_indented = True
            previous_blank = False
            code.add(index)
            continue
        previous_blank = False
        if LIST_ITEM_START.match(raw):
            in_list = True
        elif not raw.startswith((" ", "\t")):
            in_list = False
    return code


def _headings(lines):
    """Return (index, level, text) for each ATX heading outside code."""
    found = []
    code = _code_lines(lines)
    for index, raw in enumerate(lines):
        if index in code:
            continue
        m = HEADING.match(raw)
        if m:
            found.append((index, len(m.group("level")), m.group("text")))
    return found


def load_glossary(start):
    """Return the glossary of the repository of START, and its path.

    The glossary is the file .scraibe.json in the directory of START or
    in a parent directory. Without one, the result is (None, None).
    """
    directory = os.path.abspath(start)
    if not os.path.isdir(directory):
        directory = os.path.dirname(directory)
    while True:
        candidate = os.path.join(directory, GLOSSARY_NAME)
        if os.path.isfile(candidate):
            with open(candidate, encoding="utf-8") as handle:
                return json.load(handle), candidate
        parent = os.path.dirname(directory)
        if parent == directory:
            return None, None
        directory = parent


def _glossary_patterns(glossary):
    """Return [(preferred, compiled pattern)] for the terms of a glossary."""
    patterns = []
    for preferred, synonyms in (glossary or {}).get("terms", {}).items():
        for synonym in synonyms:
            words = r"\s+".join(re.escape(w) for w in synonym.split())
            patterns.append((preferred, re.compile(r"\b" + words + r"(?:s|es|ed|ing)?\b", re.I)))
    return patterns


def _bluf_findings(text, filename):
    """Test the BLUF section (SKILL.md section 4c).

    A README.md must have a heading "BLUF". In every file, the BLUF is the
    first section after the title, holds at most BLUF_MAX_WORDS words and
    BLUF_MAX_SENTENCES sentences, and has no list, code block, or link.
    The linter cannot test if the BLUF is true.
    """
    lines = text.splitlines()
    headings = _headings(lines)
    bluf = [h for h in headings if h[2] == "BLUF"]
    basename = filename.replace("\\", "/").rsplit("/", 1)[-1].lower()
    if not bluf:
        if basename in README_NAMES:
            return [{"file": filename, "line": 1, "col": 1, "rule": "bluf-missing",
                     "level": "advisory-free", "match": "no BLUF heading",
                     "message": "README.md has no BLUF section. Add one after the title."}]
        return []
    index, level, _ = bluf[0]
    findings = []
    position = [h[0] for h in headings].index(index)
    if position > 1:
        findings.append({"file": filename, "line": index + 1, "col": 1,
                         "rule": "bluf-position", "level": "advisory-free",
                         "match": "#" * level + " BLUF",
                         "message": "The BLUF is not the first section after the title. Move it."})
    body = []
    for raw in lines[index + 1:]:
        if HEADING.match(raw):
            break
        body.append(raw)
    in_fence = False
    words = 0
    prose = []
    for offset, raw in enumerate(body):
        stripped = raw.strip()
        if CODE_FENCE.match(stripped):
            in_fence = not in_fence
            findings.append({"file": filename, "line": index + 2 + offset, "col": 1,
                             "rule": "bluf-shape", "level": "advisory-free",
                             "match": stripped[:3],
                             "message": "The BLUF has a code block. Write one paragraph of prose."})
            continue
        if in_fence:
            continue
        if LIST_ITEM_START.match(raw):
            findings.append({"file": filename, "line": index + 2 + offset, "col": 1,
                             "rule": "bluf-shape", "level": "advisory-free",
                             "match": stripped[:20],
                             "message": "The BLUF has a list. Write one paragraph of prose."})
        if "](" in raw:
            findings.append({"file": filename, "line": index + 2 + offset,
                             "col": raw.index("](") + 1,
                             "rule": "bluf-shape", "level": "advisory-free",
                             "match": "](",
                             "message": "The BLUF has a link. Put the link in the body of the document."})
        line = INLINE_CODE.sub("x", raw)
        words += len(line.split())
        prose.append(line.strip())
    sentences = len([t for t in re.split(r"(?<=[.!?])\s+", " ".join(prose)) if t.strip()])
    if words > BLUF_MAX_WORDS or sentences > BLUF_MAX_SENTENCES:
        findings.append({"file": filename, "line": index + 1, "col": 1,
                         "rule": "bluf-long", "level": "advisory-free",
                         "match": f"{words} words, {sentences} sentences",
                         "message": f"The BLUF has {words} words and {sentences} sentences "
                                    f"(cap {BLUF_MAX_WORDS} words, {BLUF_MAX_SENTENCES} sentences). Shorten it."})
    return findings


def _command_path(tokens):
    """Return the command paths in a usage or shell line.

    A path is the run of command words before the first placeholder,
    option, or value: ("vud", "team", "add"). A token "show|approve"
    branches into two paths. The result is a list of tuples. It is empty
    when the first token is not a command word.
    """
    paths = [()]
    for token in tokens:
        parts = token.split("|") if "|" in token else [token]
        if not all(COMMAND_WORD.match(part) for part in parts):
            break
        paths = [path + (part,) for path in paths for part in parts]
    return [path for path in paths if path]


def _usage_commands(lines):
    """Return the command paths of the Usage or Synopsis section.

    The section is the text under a heading "Usage" or "Synopsis" up to
    the next heading. Its commands are the lines in code fences and the
    lines indented by four spaces or a tab.
    """
    commands = set()
    headings = _headings(lines)
    for position, (index, _, text) in enumerate(headings):
        if text.strip().lower() not in USAGE_HEADINGS:
            continue
        end = headings[position + 1][0] if position + 1 < len(headings) else len(lines)
        in_fence = False
        for raw in lines[index + 1:end]:
            if CODE_FENCE.match(raw.strip()):
                in_fence = not in_fence
                continue
            if not in_fence and not raw.startswith(("    ", "\t")):
                continue
            tokens = raw.strip().split()
            if tokens and tokens[0] == "$":
                tokens = tokens[1:]
            for path in _command_path(tokens):
                commands.add(path)
    return commands


def _section_ranges(lines, names):
    """Return (start, end) line indices of each section whose heading is in NAMES."""
    headings = _headings(lines)
    ranges = []
    for position, (index, _, text) in enumerate(headings):
        if text.strip().lower() in names:
            end = headings[position + 1][0] if position + 1 < len(headings) else len(lines)
            ranges.append((index, end))
    return ranges


def _usage_flags(lines):
    """Return the flags of the Usage, Synopsis, Options, and Configuration sections."""
    flags = set()
    for start, end in _section_ranges(lines, USAGE_HEADINGS | OPTIONS_HEADINGS):
        for raw in lines[start + 1:end]:
            flags.update(FLAG.findall(raw))
    return flags


def _ghost_command_findings(text, filename):
    """Flag a command or a flag in prose that the Usage section does not have.

    The rule tests the inline code spans outside the Usage and Options
    sections, and the lines of shell code fences. A span counts when its
    first word is a program of the Usage section, or when it starts with
    a dash. Its command path must be a prefix of a usage command, or a
    usage command must be a prefix of it (`ghost-command`). Each of its
    flags must be in the Usage, Synopsis, Options, or Configuration
    section (`ghost-flag`). A file without a Usage section gets no
    finding, and a file whose sections have no flag gets no flag finding.
    """
    lines = text.splitlines()
    commands = _usage_commands(lines)
    if not commands:
        return []
    programs = {path[0] for path in commands}
    flags = _usage_flags(lines)

    def known(path):
        return any(path[:len(c)] == c or c[:len(path)] == path for c in commands)

    findings = []

    def test(tokens, lineno, col):
        if not tokens:
            return
        if tokens[0] in programs:
            for path in _command_path(tokens):
                if not known(path):
                    findings.append({"file": filename, "line": lineno, "col": col,
                                     "rule": "ghost-command", "level": "advisory-free",
                                     "match": " ".join(path),
                                     "message": f"The Usage section has no command `{' '.join(path)}`. "
                                                "Add it there, or correct the text."})
        elif not tokens[0].startswith("-"):
            return
        if not flags:
            return
        for token in tokens:
            for flag in FLAG.findall(token):
                if flag not in flags:
                    findings.append({"file": filename, "line": lineno, "col": col,
                                     "rule": "ghost-flag", "level": "advisory-free",
                                     "match": flag,
                                     "message": f"The Usage and Options sections have no flag `{flag}`. "
                                                "Add it there, or correct the text."})

    usage_ranges = _section_ranges(lines, USAGE_HEADINGS | OPTIONS_HEADINGS)
    in_fence = False
    shell_fence = False
    for lineno, raw in enumerate(lines, 1):
        stripped = raw.strip()
        if CODE_FENCE.match(stripped):
            if not in_fence:
                shell_fence = bool(SHELL_FENCE.match(stripped))
            in_fence = not in_fence
            continue
        if any(start <= lineno - 1 < end for start, end in usage_ranges):
            continue
        if in_fence:
            if shell_fence:
                for segment in SHELL_SPLIT.split(raw):
                    tokens = segment.split()
                    if tokens and tokens[0] == "$":
                        tokens = tokens[1:]
                    test(tokens, lineno, 1)
            continue
        for m in INLINE_CODE.finditer(raw):
            test(m.group(0).strip("`").split(), lineno, m.start() + 1)
    return findings


def _fix_segment(segment):
    """Apply the mechanical fixes to one piece of prose."""
    count = 0

    def semicolon(m):
        nonlocal count
        count += 1
        return ". " + m.group(1).upper()

    # A semicolon between two clauses becomes a period. An HTML entity
    # keeps its semicolon.
    pieces = []
    last = 0
    for entity in HTML_ENTITY.finditer(segment):
        pieces.append(re.sub(r";\s+([a-z])", semicolon, segment[last:entity.start()]))
        pieces.append(entity.group(0))
        last = entity.end()
    pieces.append(re.sub(r";\s+([a-z])", semicolon, segment[last:]))
    segment = "".join(pieces)

    def adjective(m):
        nonlocal count
        count += 1
        return ""

    segment = re.sub(
        r" ?\b(seamless(?:ly)?|robust(?:ly)?|cutting-edge|effortless(?:ly)?|blazing[- ]fast|world-class|state-of-the-art)\b",
        adjective, segment, flags=re.I)

    def phrasal(m):
        nonlocal count
        count += 1
        plain = PHRASAL_FIXES[m.group(0).lower()]
        return plain.capitalize() if m.group(0)[0].isupper() else plain

    pattern = r"\b(" + "|".join(re.escape(k) for k in sorted(PHRASAL_FIXES, key=len, reverse=True)) + r")\b"
    segment = re.sub(pattern, phrasal, segment, flags=re.I)

    def phrase(m):
        nonlocal count
        count += 1
        plain = PHRASE_FIXES[m.group(0).lower()]
        return plain.capitalize() if m.group(0)[0].isupper() else plain

    pattern = r"\b(" + "|".join(re.escape(k) for k in sorted(PHRASE_FIXES, key=len, reverse=True)) + r")(?=[\s,;:)]|$)"
    segment = re.sub(pattern, phrase, segment, flags=re.I)

    def filler(m):
        nonlocal count
        count += 1
        return ""

    segment = re.sub(r" ?\b(?:simply|easily)\b", filler, segment, flags=re.I)

    def contraction(m):
        nonlocal count
        count += 1
        word = m.group(1)
        if word.lower() == "ca":
            return word + "nnot"
        if word.lower() == "wo":
            return ("W" if word[0].isupper() else "w") + "ill not"
        return word + " not"

    segment = re.sub(r"\b(\w+)n['\u2019]t\b", contraction, segment)
    return segment, count


def fix(text):
    """Return the text with the mechanical fixes applied, and the count.

    The fixes run outside code fences and outside inline code. They
    change semicolons, the marketing adjectives, and the phrasal verbs of
    PHRASAL_FIXES. They change nothing else. Read the diff after a run.
    """
    out = []
    total = 0
    code_lines = _code_lines(text.splitlines())
    for index, raw in enumerate(text.splitlines(keepends=True)):
        if index in code_lines:
            out.append(raw)
            continue
        pieces = []
        last = 0
        for code in INLINE_CODE.finditer(raw):
            fixed, n = _fix_segment(raw[last:code.start()])
            total += n
            pieces.append(fixed)
            pieces.append(code.group(0))
            last = code.end()
        fixed, n = _fix_segment(raw[last:])
        total += n
        pieces.append(fixed)
        out.append("".join(pieces))
    return "".join(out), total


def _suppressions(lines):
    """Return the rules that a directive comment disables on each line.

    Directives are HTML comments:
        <!-- ste-lint: disable-next-line RULE[,RULE] -->
        <!-- ste-lint: disable-line RULE -->     on the line itself
        <!-- ste-lint: disable RULE -->          until the matching enable
        <!-- ste-lint: enable RULE -->
        <!-- ste-lint: disable-file RULE -->
    A directive without a rule name disables every rule. The result maps
    a 1-based line number to a set of rule ids, or {"*"} for every rule.
    """
    file_rules = set()
    block_rules = set()
    next_rules = set()
    result = {}
    for lineno, raw in enumerate(lines, 1):
        line_rules = set(block_rules) | set(next_rules)
        next_rules = set()
        for m in DIRECTIVE.finditer(raw):
            rules = {r for r in re.split(r"[\s,]+", m.group("rules").strip()) if r} or {"*"}
            kind = m.group("kind")
            if kind == "disable-file":
                file_rules |= rules
            elif kind == "disable-next-line":
                next_rules |= rules
            elif kind == "disable-line":
                line_rules |= rules
            elif kind == "disable":
                block_rules |= rules
                line_rules |= rules
            elif kind == "enable":
                block_rules -= rules
                if "*" in rules:
                    block_rules = set()
        if line_rules or file_rules:
            result[lineno] = line_rules | file_rules
    return result


def _heading_case_findings(lines, filename):
    """Flag a heading in title case: two or more capitalized words after the first."""
    findings = []
    for index, _, text in _headings(lines):
        words = INLINE_CODE.sub("", text).split()[1:]
        capitals = [w for w in words if CAPITALIZED_WORD.match(w)]
        if len(capitals) >= 2:
            findings.append({"file": filename, "line": index + 1, "col": 1,
                             "rule": "heading-case", "level": "advisory",
                             "match": " ".join(capitals),
                             "message": "Possible title case. Use sentence case in a heading, unless the words are names."})
    return findings


def lint(text, filename="<stdin>", glossary=None):
    findings = []
    words_total = 0
    lines = text.splitlines()
    code = _code_lines(lines)
    table_cells = _markdown_table_cells(lines)
    glossary_patterns = _glossary_patterns(glossary)
    # first occurrence of each synonym-group member: (group_idx, base) -> (line, col, match)
    seen_synonyms = {}
    # A hard-wrapped paragraph puts the middle of a sentence at the start
    # of a line. The sentence rules skip the first piece of such a line.
    previous_prose = ""
    for lineno, raw_line in enumerate(lines, 1):
        if (lineno - 1) in code:
            previous_prose = ""
            continue
        raw_line = HTML_COMMENT.sub("", raw_line)
        continuation = bool(previous_prose.strip()) and not re.search(r"[.!?:]\s*$", previous_prose)
        previous_prose = raw_line
        in_table = (lineno - 1) in table_cells
        segments = table_cells.get(lineno - 1, [(raw_line, 0)])
        for segment, source_column in segments:
            line = INLINE_CODE.sub("", segment)
            words_total += len(line.split())
            for rule_id, level, pattern, msg in RULES:
                for m in pattern.finditer(line):
                    findings.append({"file": filename, "line": lineno,
                                     "col": source_column + m.start() + 1,
                                     "rule": rule_id, "level": level,
                                     "match": m.group(0), "message": msg})
            for preferred, pattern in glossary_patterns:
                for m in pattern.finditer(line):
                    findings.append({"file": filename, "line": lineno,
                                     "col": source_column + m.start() + 1,
                                     "rule": "glossary", "level": "advisory-free",
                                     "match": m.group(0),
                                     "message": f'Write "{preferred}", not "{m.group(0)}" ({GLOSSARY_NAME}).'})
            for gi, group in enumerate(SYNONYM_GROUPS):
                if HEADING.match(raw_line):
                    break  # a heading word is a label, not a choice of verb
                for base in group:
                    if (gi, base) in seen_synonyms:
                        continue
                    m = _word_re(base).search(line)
                    if m:
                        seen_synonyms[(gi, base)] = (
                            lineno, source_column + m.start() + 1, m.group(0)
                        )
            prose = line
            item = LIST_ITEM_START.match(line)
            if item:
                prose = item.group("body")
            for position, sent in enumerate(re.split(r"(?<=[.!?])\s+", prose)):
                sent = sent.strip()
                # SKILL.md section 3: text in parentheses counts as one word.
                # A code span was removed above, so a comma it leaves is not a word.
                n = len([w for w in re.sub(r"\([^()]*\)", "x", sent).split() if re.search(r"\w", w)])
                starts_sentence = not (position == 0 and continuation and not item)
                if n > MAX_WORDS:
                    findings.append({"file": filename, "line": lineno,
                                     "col": source_column + 1,
                                     "rule": "long-sentence", "level": "advisory-free",
                                     "match": f"{n} words",
                                     "message": f"Sentence has {n} words (cap {MAX_WORDS}). Split it."})
                if not starts_sentence:
                    continue
                if n >= 2 and not in_table and sent[0].isdigit():
                    findings.append({"file": filename, "line": lineno,
                                     "col": source_column + 1,
                                     "rule": "numeral-start", "level": "advisory-free",
                                     "match": sent.split()[0],
                                     "message": "Sentence starts with a numeral. Rewrite the sentence, or write the number in words."})
                if IMPERATIVE.match(sent) and CONDITION.search(sent):
                    findings.append({"file": filename, "line": lineno,
                                     "col": source_column + 1,
                                     "rule": "condition-last", "level": "advisory",
                                     "match": sent[:40],
                                     "message": "Possible condition after the instruction. Write the condition first."})
                if AMBIGUOUS_THIS.match(sent):
                    findings.append({"file": filename, "line": lineno,
                                     "col": source_column + 1,
                                     "rule": "ambiguous-this", "level": "advisory",
                                     "match": sent[:40],
                                     "message": 'Sentence starts with "This". Name the thing it refers to.'})
    # synonym rotation: flag each member after the first, at its first occurrence
    for gi, group in enumerate(SYNONYM_GROUPS):
        present = [(seen_synonyms[(gi, b)], b) for b in group if (gi, b) in seen_synonyms]
        if len(present) > 1:
            present.sort()  # document order
            first_base = present[0][1]
            for (lineno, col, match), base in present[1:]:
                findings.append({"file": filename, "line": lineno, "col": col,
                                 "rule": "synonym-rotation", "level": "advisory-free",
                                 "match": match,
                                 "message": f"'{base}' and '{first_base}' name the same action. Pick one and use it every time."})
    findings.extend(_dangling_conjunction_findings(text, filename))
    findings.extend(_bluf_findings(text, filename))
    findings.extend(_ghost_command_findings(text, filename))
    findings.extend(_heading_case_findings(lines, filename))
    suppressed = _suppressions(lines)
    findings = [f for f in findings
                if not ({"*", f["rule"]} & suppressed.get(f["line"], set()))]
    findings.sort(key=lambda f: (f["line"], f["col"]))
    return findings, words_total


def report(findings, words_total, as_json, hard_count, baseline):
    rate = round(len(findings) * 100 / words_total, 1) if words_total else 0.0
    if as_json:
        print(json.dumps({"violations": findings, "count": len(findings),
                          "hard_count": hard_count, "baseline": baseline,
                          "words": words_total, "per_100_words": rate}, indent=2))
        return
    for f in findings:
        print(f"{f['file']}:{f['line']}:{f['col']} {f['rule']}: {f['message']} [{f['match']}]")
    print(f"\n{len(findings)} violations ({hard_count} hard, baseline {baseline}), "
          f"{words_total} words, {rate} per 100 words")
    print("Hedges/modality (may, might, could) are never flagged: confidence is content.")


def selftest():
    bad = ("The panel is removed; spin up the job. "
           "Perform an analysis of the seamless log. "
           "We have received the report. Initialize and verify the cache.")
    findings, _ = lint(bad)
    rules = {f["rule"] for f in findings}
    for expected in ("semicolon", "phrasal-verb", "nominalization",
                     "marketing-adjective", "passive-voice", "present-perfect",
                     "latin-verb"):
        assert expected in rules, expected
    # spellings that only look like a Latin verb are not flagged
    findings, _ = lint("Otherwise the noise, the size, and the premise rise.")
    assert not [f for f in findings if f["rule"] == "latin-verb"], findings
    # hedges must never be flagged, including modal + perfect infinitive
    findings, _ = lint("The request may have failed. It could be a timeout. "
                       "The disk might have filled.")
    assert findings == [], findings
    # code blocks skipped
    findings, _ = lint("```\nx = a; y = b\n```")
    assert findings == []
    # all supported list markers, case variants, and trailing whitespace
    findings, _ = lint(
        "- Confirm the target and\n"
        "* Record the result OR  \n"
        "+ Close the panel\n"
        "1. Start the task and\n"
        "2) Stop the task OR"
    )
    dangling = [f for f in findings if f["rule"] == "dangling-conjunction"]
    assert len(dangling) == 4, dangling
    assert [f["line"] for f in dangling] == [1, 2, 4, 5], dangling
    assert [f["col"] for f in dangling] == [1, 1, 1, 1], dangling
    assert all(f["level"] == "advisory-free" for f in dangling), dangling

    # valid continuation lines and standalone four-space code are ignored
    findings, _ = lint("  - Confirm the target and\n    record the result.")
    assert not any(f["rule"] == "dangling-conjunction" for f in findings)
    findings, _ = lint("- Confirm the target\n  and")
    dangling = [f for f in findings if f["rule"] == "dangling-conjunction"]
    assert len(dangling) == 1 and dangling[0]["line"] == 2, dangling
    assert dangling[0]["col"] == 3, dangling
    findings, _ = lint("    - code and")
    assert not any(f["rule"] == "dangling-conjunction" for f in findings)
    findings, _ = lint("> - Confirm the target and\n> - Record the result or")
    assert not any(f["rule"] == "dangling-conjunction" for f in findings)
    findings, _ = lint("- Do this and\n~~~\ncode and\n~~~")
    dangling = [f for f in findings if f["rule"] == "dangling-conjunction"]
    assert len(dangling) == 1 and dangling[0]["line"] == 1, dangling
    findings, _ = lint("```text\n- code and\n```")
    assert not any(f["rule"] == "dangling-conjunction" for f in findings)

    # one- and three-space markers and ordered continuation width
    findings, _ = lint(" - Start the task and\n   record the result.")
    assert not any(f["rule"] == "dangling-conjunction" for f in findings)
    findings, _ = lint("-  Start the task and\n   record the result.")
    assert not any(f["rule"] == "dangling-conjunction" for f in findings)
    findings, _ = lint("-\tStart the task and")
    assert not any(f["rule"] == "dangling-conjunction" for f in findings)
    findings, _ = lint("   - Start the task and", filename="fixture.md")
    dangling = [f for f in findings if f["rule"] == "dangling-conjunction"]
    assert len(dangling) == 1 and dangling[0]["col"] == 4, dangling
    assert dangling[0]["file"] == "fixture.md"
    assert dangling[0]["match"].endswith("and")
    assert "Complete the item" in dangling[0]["message"]
    findings, _ = lint("100. Start the task and\n  unrelated text")
    dangling = [f for f in findings if f["rule"] == "dangling-conjunction"]
    assert len(dangling) == 1, dangling
    findings, _ = lint("- Start the task and.\n- Stop the task or,")
    assert not any(f["rule"] == "dangling-conjunction" for f in findings)
    findings, _ = lint("- Start the task and\n\n  record the result.")
    assert not any(f["rule"] == "dangling-conjunction" for f in findings)
    findings, _ = lint("- Parent item and\n  - Nested item or")
    dangling = [f for f in findings if f["rule"] == "dangling-conjunction"]
    assert [f["line"] for f in dangling] == [1, 2], dangling

    # ordinary prose, inline code, and fenced code are ignored
    findings, _ = lint("The process may include steps and")
    assert not any(f["rule"] == "dangling-conjunction" for f in findings)
    findings, _ = lint("- Use `and` as a label")
    assert not any(f["rule"] == "dangling-conjunction" for f in findings)
    findings, _ = lint("- Combine `left` and `right`")
    assert not any(f["rule"] == "dangling-conjunction" for f in findings), findings
    findings, _ = lint("~~~\n- code and\n~~~")
    assert not any(f["rule"] == "dangling-conjunction" for f in findings)
    findings, _ = lint(("word " * 30).strip() + ".")
    assert any(f["rule"] == "long-sentence" for f in findings)
    # Markdown table syntax is layout, not prose. Each cell stays lintable.
    short_cell = " ".join(f"term{number}" for number in range(1, 25)) + "."
    for table in (
            "| Label | Detail |\n"
            "| --- | --- |\n"
            f"| Clear | {short_cell} |",
            "Label | Detail\n"
            "--- | ---\n"
            f"Clear | {short_cell}"):
        findings, words_total = lint(table)
        assert not any(f["rule"] == "long-sentence" for f in findings), findings
        assert words_total == 27, words_total
    long_cell = " ".join(f"term{number}" for number in range(1, 27)) + "."
    findings, _ = lint(
        "| Label | Detail |\n"
        "| --- | --- |\n"
        f"| Clear | {long_cell} |"
    )
    long_sentences = [f for f in findings if f["rule"] == "long-sentence"]
    assert len(long_sentences) == 1, long_sentences
    assert long_sentences[0]["match"] == "26 words", long_sentences
    # synonym rotation: second member flagged, first named as the keeper
    findings, _ = lint("Check the config file. Then verify the output. Verify twice.")
    rot = [f for f in findings if f["rule"] == "synonym-rotation"]
    assert len(rot) == 1 and "'verify' and 'check'" in rot[0]["message"], rot
    # a heading word is not a synonym choice
    findings, _ = lint("### Fixed\n\nThe tool corrects the path.\n")
    assert not any(f["rule"] == "synonym-rotation" for f in findings), findings
    # single consistent term: no flag
    findings, _ = lint("Check the config. Check the output.")
    assert not any(f["rule"] == "synonym-rotation" for f in findings)
    # per-file labels
    findings, _ = lint("a; b", filename="x.md")
    assert findings[0]["file"] == "x.md"
    # BLUF: a README.md without the section is flagged, other files are not
    findings, _ = lint("# tool - does a thing\n\n## Usage\n", filename="docs/README.md")
    assert [f["rule"] for f in findings] == ["bluf-missing"], findings
    findings, _ = lint("# tool - does a thing\n\n## Usage\n", filename="docs/guide.md")
    assert not any(f["rule"].startswith("bluf") for f in findings), findings
    good = "# tool - does a thing\n\n## BLUF\n\ntool copies files. The copy works.\n\n## Usage\n"
    findings, _ = lint(good, filename="README.md")
    assert not any(f["rule"].startswith("bluf") for f in findings), findings
    # position, shape, and length
    findings, _ = lint("# t\n\n## Usage\n\n## BLUF\n\nText.\n", filename="README.md")
    assert [f["rule"] for f in findings] == ["bluf-position"], findings
    findings, _ = lint("# t\n\n## BLUF\n\n- item\n\nSee [x](y).\n\n```\ncode\n```\n")
    assert [f["rule"] for f in findings] == ["bluf-shape"] * 4, findings
    findings, _ = lint("# t\n\n## BLUF\n\n" + "word " * 81 + "\n")
    assert [f["rule"] for f in findings if f["rule"].startswith("bluf")] == ["bluf-long"], findings
    findings, _ = lint("# t\n\n## BLUF\n\n" + "One. " * 7 + "\n")
    assert [f["rule"] for f in findings if f["rule"].startswith("bluf")] == ["bluf-long"], findings
    # indented code is code, a list continuation is prose
    findings, _ = lint("Text.\n\n    a; b etc.\n    c; d\n\nMore; text.\n")
    assert [f["line"] for f in findings if f["rule"] == "semicolon"] == [6], findings
    findings, _ = lint("- Item one.\n\n    Continued; here.\n")
    assert [f["line"] for f in findings if f["rule"] == "semicolon"] == [3], findings
    fixed, _ = fix("Text.\n\n    a; b\n")
    assert fixed == "Text.\n\n    a; b\n", fixed
    # glossary
    glossary = {"terms": {"config file": ["configuration file", "settings"]}}
    findings, _ = lint("Edit the configuration file and the Settings.", glossary=glossary)
    hits = [f for f in findings if f["rule"] == "glossary"]
    assert [f["match"] for f in hits] == ["configuration file", "Settings"], hits
    assert 'Write "config file"' in hits[0]["message"], hits
    assert not _glossary_patterns(None)
    # ghost-flag: flags against Usage and Options
    usage_flags = ("# t - tool\n\n## Usage\n\n    t run [--fast] [--config FILE]\n\n## Options\n\n"
                   "`--fast`\n: Run fast.\n\n`--verbose`\n: Say more.\n\n## Notes\n\n")
    findings, _ = lint(usage_flags + "Use `t run --fast --verbose`, `--config`, and `t run --slow`. Also `--dry-run`.\n")
    ghosts = [f["match"] for f in findings if f["rule"] == "ghost-flag"]
    assert ghosts == ["--slow", "--dry-run"], ghosts
    findings, _ = lint("# t\n\n## Usage\n\n    t run\n\n## Notes\n\nUse `t run --x` and `--y`.\n")
    assert not any(f["rule"] == "ghost-flag" for f in findings), findings
    # ghost-command: prose and shell fences against the Usage section
    usage = ("# vud - tracker\n\n## Usage\n\n    vud serve [--config FILE]\n"
             "    vud policy show|approve [--config FILE]\n"
             "    vud team add NAME [--names A,B]\n      [--runtime docker|podman]\n\n")
    findings, _ = lint(usage + "## Description\n\nRun `vud serve --config x`, `vud policy approve`, "
                       "`vud team add linux --names A`, and `vud team`.\n")
    assert not any(f["rule"] == "ghost-command" for f in findings), findings
    findings, _ = lint(usage + "## Description\n\nRun `vud team rename a b` or `vud stop`.\n"
                       "See `vud-postgres` and `other tool`.\n\n```sh\n$ vud serve && vud bogus\n```\n")
    ghosts = [f["match"] for f in findings if f["rule"] == "ghost-command"]
    assert ghosts == ["vud team rename a b", "vud stop", "vud bogus"], ghosts
    findings, _ = lint("# tool\n\nRun `tool whatever`.\n")
    assert not any(f["rule"] == "ghost-command" for f in findings), findings
    findings, _ = lint("# t\n\n## Synopsis\n\n```\nt run\n```\n\n## Notes\n\nUse `t walk`.\n")
    assert [f["match"] for f in findings if f["rule"] == "ghost-command"] == ["t walk"], findings
    # --fix: mechanical rewrites only, code untouched
    fixed, n = fix("The panel is removed; spin up the job. A robust tool.\n"
                   "Keep `a; b` and &amp; as is.\n```\nx; y\n```\n")
    assert fixed == ("The panel is removed. Start the job. A tool.\n"
                     "Keep `a; b` and &amp; as is.\n```\nx; y\n```\n"), fixed
    assert n == 3, n
    fixed, n = fix("We reach out and it runs seamlessly.")
    assert fixed == "We contact and it runs." and n == 2, (fixed, n)
    fixed, n = fix("The request may have failed; it could be a timeout.")
    assert fixed == "The request may have failed. It could be a timeout.", fixed
    # new hard rules
    findings, _ = lint("Use e.g. a file(s). Don't do it; it's — simply crucial, etc.")
    rules = {f["rule"] for f in findings}
    for expected in ("latin-abbreviation", "parenthetical-plural", "contraction",
                     "em-dash", "filler-word"):
        assert expected in rules, (expected, rules)
    assert sum(1 for f in findings if f["rule"] == "contraction") == 2, findings
    assert sum(1 for f in findings if f["rule"] == "latin-abbreviation") == 2, findings
    # a product name is not a contraction, and a parenthetical is one word
    findings, _ = lint("The certificate comes from Let's Encrypt.")
    assert not any(f["rule"] == "contraction" for f in findings), findings
    findings, _ = lint("The tool keeps " + "word " * 20 + "(and " + "more " * 10 + "here).")
    assert not any(f["rule"] == "long-sentence" for f in findings), findings
    findings, _ = lint("The keys are " + ", ".join("`k%d`" % i for i in range(30)) + " and the rest.")
    assert not any(f["rule"] == "long-sentence" for f in findings), findings
    # a possessive is not a contraction
    findings, _ = lint("The user's file and the host's name.")
    assert not any(f["rule"] == "contraction" for f in findings), findings
    # numeral at sentence start, not in a list marker or a table cell
    findings, _ = lint("0 on success. The tool exits with 1.\n1. Step one.\n| 0 | ok |\n|---|---|\n| 1 | fine |\n")
    numerals = [f for f in findings if f["rule"] == "numeral-start"]
    assert len(numerals) == 1 and numerals[0]["line"] == 1, numerals
    # a wrapped line that starts mid-sentence is not a sentence start
    findings, _ = lint('Hyphenate a count that modifies a noun: "a\n10-step procedure". Keep it.\n')
    assert not any(f["rule"] == "numeral-start" for f in findings), findings
    findings, _ = lint("Count twice.\n10 steps follow.\n")
    assert any(f["rule"] == "numeral-start" for f in findings), findings
    # advisory sentence rules
    findings, _ = lint("Run the job if the file exists. This ensures the cache is running.\n## Exit Status Codes\nUse **bold**.")
    rules = {f["rule"] for f in findings}
    for expected in ("condition-last", "ambiguous-this", "progressive-tense", "heading-case", "bold"):
        assert expected in rules, (expected, rules)
    assert all(f["level"] == "advisory" for f in findings
               if f["rule"] in ("condition-last", "ambiguous-this", "progressive-tense", "heading-case", "bold")), findings
    findings, _ = lint("If the file exists, run the job. This process is slow.\n## Teams and perspectives\n")
    assert not any(f["rule"] in ("condition-last", "ambiguous-this", "heading-case") for f in findings), findings
    # directives
    findings, _ = lint("<!-- ste-lint: disable-next-line semicolon -->\na; b\nc; d <!-- ste-lint: disable-line semicolon -->\ne; f\n")
    assert [f["line"] for f in findings if f["rule"] == "semicolon"] == [4], findings
    findings, _ = lint("<!-- ste-lint: disable semicolon,em-dash -->\na; b — c\n<!-- ste-lint: enable semicolon -->\nd; e — f\n")
    assert [(f["line"], f["rule"]) for f in findings if f["rule"] in ("semicolon", "em-dash")] == [(4, "semicolon")], findings
    findings, _ = lint("<!-- ste-lint: disable-file -->\na; b\n")
    assert findings == [], findings
    # --fix: phrases, fillers, contractions
    fixed, n = fix("In order to run it, don't wait; it can't fail, e.g. furthermore it simply works.")
    assert fixed == "To run it, do not wait. It cannot fail, for example also it works.", fixed
    assert n == 7, n
    # a fenced BLUF heading is not a heading
    findings, _ = lint("# t\n\n```\n## BLUF\n```\n", filename="README.md")
    assert [f["rule"] for f in findings] == ["bluf-missing"], findings
    print("selftest OK")


def main(argv):
    if "--selftest" in argv:
        selftest()
        return 0
    as_json = "--json" in argv
    apply_fix = "--fix" in argv
    use_glossary = "--no-glossary" not in argv
    baseline = None
    disabled = set()
    paths = []
    i = 0
    while i < len(argv):
        a = argv[i]
        if a == "--baseline":
            i += 1
            baseline = int(argv[i])
        elif a == "--disable":
            i += 1
            disabled = set(argv[i].split(","))
        elif not a.startswith("--"):
            paths.append(a)
        i += 1

    findings, words_total = [], 0
    glossary_baseline = 0
    if paths:
        for p in paths:
            glossary = load_glossary(p)[0] if use_glossary else None
            if glossary:
                disabled |= set(glossary.get("disable", []))
                glossary_baseline = max(glossary_baseline, int(glossary.get("baseline", 0)))
            text = open(p, encoding="utf-8").read()
            if apply_fix:
                fixed, n = fix(text)
                if n:
                    with open(p, "w", encoding="utf-8") as handle:
                        handle.write(fixed)
                    print(f"{p}: {n} fixes applied. Read the diff.", file=sys.stderr)
                text = fixed
            f, w = lint(text, filename=p, glossary=glossary)
            findings.extend(f)
            words_total += w
    else:
        text = sys.stdin.read()
        if apply_fix:
            text, _ = fix(text)
            sys.stdout.write(text)
            return 0
        glossary = load_glossary(os.getcwd())[0] if use_glossary else None
        if glossary:
            disabled |= set(glossary.get("disable", []))
            glossary_baseline = int(glossary.get("baseline", 0))
        findings, words_total = lint(text, glossary=glossary)

    if baseline is None:
        baseline = glossary_baseline
    findings = [f for f in findings if f["rule"] not in disabled]
    hard_count = sum(1 for f in findings if f["level"] == "advisory-free")
    report(findings, words_total, as_json, hard_count, baseline)
    return 1 if hard_count > baseline else 0


if __name__ == "__main__":
    sys.exit(main(sys.argv[1:]))
