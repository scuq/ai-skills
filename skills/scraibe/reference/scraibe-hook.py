#!/usr/bin/env python3
"""scraibe-hook - advisory lint after Claude Code writes a Markdown file.

Register it as a PostToolUse hook on Write|Edit. When the written file
ends in .md and is not under a .claude directory, the hook runs
ste-lint.py on it. When there are hard findings, it prints a one-line
summary to stderr and exits 2. Exit 2 on PostToolUse is advisory: the
write already happened, and the model sees the message.

The hook never blocks and never loops. Every error path exits 0.

Usage in ~/.claude/settings.json:

    {
      "hooks": {
        "PostToolUse": [
          {
            "matcher": "Write|Edit",
            "hooks": [
              {
                "type": "command",
                "command": "python3 ~/.claude/skills/scraibe/reference/scraibe-hook.py",
                "timeout": 10
              }
            ]
          }
        ]
      }
    }

Set SCRAIBE_LINT_EXCLUDE to a colon-separated list of glob patterns
to skip more paths.

Pattern from src/hooks/lint_hook.py in github.com/AminBlg/SimpleEnglish,
MIT license.
"""
import fnmatch
import json
import os
import pathlib
import subprocess
import sys

HERE = pathlib.Path(__file__).resolve().parent
LINTER = HERE / "ste-lint.py"


def absolute(path, cwd=None):
    return pathlib.Path(cwd or ".", pathlib.Path(path).expanduser()).absolute()


def excluded(target):
    forms = {pathlib.Path(os.path.normpath(target)), target.resolve()}
    patterns = [
        os.path.expanduser(p)
        for p in os.environ.get("SCRAIBE_LINT_EXCLUDE", "").split(os.pathsep)
        if p
    ]
    for form in forms:
        if ".claude" in form.parts:
            return True
        if any(fnmatch.fnmatch(str(form), p) for p in patterns):
            return True
    return False


def main():
    try:
        event = json.load(sys.stdin)
    except Exception:
        return 0
    if event.get("hook_event_name") != "PostToolUse":
        return 0
    path = (event.get("tool_input") or {}).get("file_path") or ""
    if not path.endswith(".md"):
        return 0
    target = absolute(path, event.get("cwd"))
    if excluded(target) or not target.is_file() or not LINTER.is_file():
        return 0
    try:
        run = subprocess.run(
            [sys.executable, str(LINTER), "--json", str(target)],
            capture_output=True, text=True, timeout=8,
        )
        report = json.loads(run.stdout or "{}")
    except Exception:
        return 0
    hard = [f for f in report.get("violations", []) if f.get("level") != "advisory"]
    if not hard:
        return 0
    rules = {}
    for f in hard:
        rules[f.get("rule", "?")] = rules.get(f.get("rule", "?"), 0) + 1
    summary = ", ".join(f"{k} {v}" for k, v in sorted(rules.items()))
    sys.stderr.write(
        f"scraibe: {target.name} has {len(hard)} hard lint findings "
        f"({summary}). Correct them in the file you just wrote, then continue.\n"
    )
    return 2


if __name__ == "__main__":
    sys.exit(main())
