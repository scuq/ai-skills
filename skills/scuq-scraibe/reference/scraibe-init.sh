#!/usr/bin/env bash
# scraibe-init.sh - install the scraibe linter in a repository.
#
# Usage: scraibe-init.sh [--no-hook] [--no-ci] [--force] [DIR]
#
# DIR is the root of a Git repository. The default is the current
# directory. The script does these steps, and prints one line for each
# step with a state word: created, kept, updated, or note.
#
#   1. Copies ste-lint.py and LICENSE-ste-lint to .scraibe/. A copy that
#      differs is updated.
#   2. Writes .scraibe.json with an empty glossary, if it does not exist.
#   3. Writes .git/hooks/pre-commit, if it does not exist. The hook lints
#      the staged Markdown files. --no-hook skips the step.
#   4. Writes .github/workflows/scraibe.yml, if it does not exist. The
#      job lints every Markdown file. --no-ci skips the step.
#
# An existing hook or workflow is never changed, unless you use --force.
# Run the script again after an update of the skill, so that the copy
# of the linter is current.
#
# Exit status:
#   0  All steps are done.
#   1  DIR is not a Git repository, or a write failed.
#   2  An option is not correct.

set -euo pipefail

here=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
hook=1
ci=1
force=0
dir=.
while [[ $# -gt 0 ]]; do
  case $1 in
    --no-hook) hook=0 ;;
    --no-ci) ci=0 ;;
    --force) force=1 ;;
    -*) echo "Usage: scraibe-init.sh [--no-hook] [--no-ci] [--force] [DIR]" >&2; exit 2 ;;
    *) dir=$1 ;;
  esac
  shift
done

if ! root=$(git -C "$dir" rev-parse --show-toplevel 2>/dev/null); then
  echo "$dir is not in a Git repository. Run git init first." >&2
  exit 1
fi
cd "$root"

# copy SRC to DST, and print the state.
copy() {
  local src=$1 dst=$2
  if [[ ! -f $dst ]]; then
    install -D -m 0644 "$src" "$dst"
    echo "created $dst"
  elif ! cmp -s "$src" "$dst"; then
    install -m 0644 "$src" "$dst"
    echo "updated $dst"
  else
    echo "kept    $dst"
  fi
}

# write the content of standard input to DST, unless it exists.
write() {
  local dst=$1 mode=$2
  if [[ -f $dst && $force -eq 0 ]]; then
    cat >/dev/null
    echo "kept    $dst"
    return
  fi
  mkdir -p "$(dirname "$dst")"
  cat >"$dst"
  chmod "$mode" "$dst"
  echo "created $dst"
}

copy "$here/ste-lint.py" .scraibe/ste-lint.py
copy "$here/LICENSE-ste-lint" .scraibe/LICENSE-ste-lint

if [[ -f .scraibe.json ]]; then
  echo "kept    .scraibe.json"
else
  cat >.scraibe.json <<'JSON'
{
  "terms": {},
  "disable": [],
  "baseline": 0
}
JSON
  echo "created .scraibe.json"
fi

if [[ $hook -eq 1 ]]; then
  hooks_dir=$(git rev-parse --git-path hooks)
  write "$hooks_dir/pre-commit" 0755 <<'HOOK'
#!/usr/bin/env bash
# pre-commit - lint the staged Markdown files with the scraibe linter.
#
# scraibe-init.sh installed this hook. Set SCRAIBE_SKIP=1 to skip it once.
#
# Exit status: 0 when the files pass, 1 when the linter finds a hard
# finding above the baseline of .scraibe.json or the linter is missing.
set -euo pipefail
if [[ ${SCRAIBE_SKIP:-} == 1 ]]; then
  exit 0
fi
root=$(git rev-parse --show-toplevel)
linter=$root/.scraibe/ste-lint.py
if [[ ! -f $linter ]]; then
  echo "pre-commit: $linter is missing. Run scraibe-init.sh, or set SCRAIBE_SKIP=1." >&2
  exit 1
fi
mapfile -d '' files < <(git diff --cached --name-only --diff-filter=ACM -z -- '*.md')
if [[ ${#files[@]} -eq 0 ]]; then
  exit 0
fi
python3 "$linter" "${files[@]}"
HOOK
fi

if [[ $ci -eq 1 ]]; then
  write .github/workflows/scraibe.yml 0644 <<'YAML'
# Lint the Markdown files with the scraibe linter on each push and pull request.
name: scraibe

on:
  push:
  pull_request:

jobs:
  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Lint the Markdown files
        run: |
          python3 .scraibe/ste-lint.py --selftest
          git ls-files -z '*.md' | xargs -0 python3 .scraibe/ste-lint.py
YAML
fi

echo "note    Run: python3 .scraibe/ste-lint.py \$(git ls-files '*.md')"
echo "note    If old files have hard findings, set \"baseline\" in .scraibe.json to the count, and lower it over time."
