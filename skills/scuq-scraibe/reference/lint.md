# Lint details

## BLUF

This file gives the details of `ste-lint.py` that section 9 of the standard only names.
They are the fix mode, the directive comments, the glossary of a repository, the init script, and the evaluation tasks.
Read it before you use `--fix`, write a directive, change a glossary, or install the linter in a repository.

## Fix mode, directives, glossary, and init

<!-- ste-lint: disable filler-word -->
`--fix` changes a semicolon between clauses to a period, deletes a
marketing adjective and the fillers "simply" and "easily", writes a
contraction in -n't as two words, and replaces a soft phrasal verb,
"in order to", "under the hood", "out of the box", "furthermore",
"moreover", "e.g.", and "i.e." with the plain words. It changes
nothing else, and it does not touch code. It can leave "an" before a
word that now starts with a consonant. Read the diff after a run, and
run `--fix` only on text that you wrote or that the task asks you to
fix.
<!-- ste-lint: enable filler-word -->

A directive in an HTML comment disables a rule where the text quotes
a bad example or a protected name:

    <!-- ste-lint: disable-next-line semicolon,filler-word -->
    <!-- ste-lint: disable-line em-dash -->        (on the line itself)
    <!-- ste-lint: disable synonym-rotation -->    ... <!-- ste-lint: enable synonym-rotation -->
    <!-- ste-lint: disable-file bold -->

A directive without a rule name disables every rule. Use a directive
only for a quotation or a name. Never use it to pass a review.

A repository can have a glossary `.scraibe.json` at its root:

    {
      "terms": {"config file": ["configuration file", "settings file"]},
      "disable": ["bold"],
      "baseline": 0
    }

`terms` maps one preferred term to the words that the linter flags
with the rule `glossary`. `disable` names the rules that the
repository does not use. `baseline` is the default for `--baseline`.
The linter reads the glossary from the directory of the file or from
a parent directory. `--no-glossary` ignores it. Read the glossary
before you write, and use its preferred terms. When the task names a
term that the project uses for one thing, add it to the glossary.

When a repository has no `.scraibe.json`, suggest the script to the user once, in the report. Do not run it yourself, because it writes a hook and a workflow.
`reference/scraibe-init.sh DIR` installs the linter in a repository.
It copies `ste-lint.py` to `.scraibe/`, writes an empty glossary, a
pre-commit hook that lints the staged Markdown files, and a GitHub
Actions job that lints every Markdown file. An existing hook or
workflow stays, unless you use `--force`. Run the script again after
an update of the skill, so that the copy of the linter is current.

## Evaluations

`evals/evals.json` holds the evaluation tasks of this skill. Each task
has a query, a fixture directory, the expected behavior, and shell
checks that test the result. `evals/run.sh` runs each task in a copy
of its fixture with `claude -p`, then runs the checks. The run uses a
temporary configuration directory that holds the repository copy of
the skill and the agent, so the installed copies do not matter.
`--without-skill` runs the same tasks with no skill, which is the
baseline. `--model` selects the model, by default sonnet, the model
of the agent. Run the evaluations after a
change to the rules, with each model you use, and compare the two
runs. A check tests shape, as the linter does. A human reads the
output for meaning.
