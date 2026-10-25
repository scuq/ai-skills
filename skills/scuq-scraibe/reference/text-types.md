# Text types

## BLUF

This file gives the mode and the pattern of each text type.
The types are error message, runbook, incident report, commit message, release notes, agent instructions, and UI copy.
Read the entry for the type before you write it.

## Types

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
  limit in the body. Delete "this MR aims to". <!-- ste-lint: disable-line filler-word -->
- Release notes and CHANGELOG. Section 7. A `Breaking:` entry follows
  the warning pattern, command first: "Update your calls to
  `/v2/users`. The `name` field split into `first_name` and
  `last_name`."
- Agent instructions (CLAUDE.md, AGENTS.md, subagents, skills).
  Strict. One instruction per sentence. One word for one action. A
  condition first. Never "should": a model reads it as optional.
- UI copy. Strict, hard length limits. Buttons and labels are
  technical names and are exempt.
- Marketing text, launch posts, brand voice. Out of scope. scraibe <!-- ste-lint: disable-line synonym-rotation -->
  deletes persuasion. Say so and offer to do the docs instead.
