#!/usr/bin/env bash
# test.sh - run gofmt, go vet, and go test for each Go module in the repository.
#
# Usage: scripts/test.sh
#
# Exit status:
#   0  All checks passed.
#   1  A check failed.

set -euo pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/.."

status=0

# check_name reports an item whose name does not start with scuq-, or whose
# frontmatter name is not the directory or file name.
check_name() {
  local name=$1 file=$2 declared
  if [[ $name != scuq-* ]]; then
    echo "The name $name does not start with scuq-. Rename it, and change the frontmatter name." >&2
    status=1
  fi
  if [[ ! -f $file ]]; then
    echo "The file $file does not exist." >&2
    status=1
    return
  fi
  declared=$(awk 'NR == 1 && $0 != "---" { exit } NR > 1 && $0 == "---" { exit } /^name:/ { sub(/^name:[ \t]*/, ""); print; exit }' "$file")
  if [[ $declared != "$name" ]]; then
    echo "The frontmatter name in $file is \"$declared\". Change it to $name." >&2
    status=1
  fi
}

echo "== item names"
shopt -s nullglob
for dir in skills/*/; do
  check_name "$(basename "$dir")" "${dir}SKILL.md"
done
for file in agents/*.md; do
  check_name "$(basename "$file" .md)" "$file"
done
shopt -u nullglob

while IFS= read -r mod; do
  dir=$(dirname "$mod")
  echo "== $dir"
  unformatted=$(cd "$dir" && gofmt -l .)
  if [[ -n $unformatted ]]; then
    echo "These files are not formatted. Run gofmt -w on them:" >&2
    echo "$unformatted" >&2
    status=1
  fi
  (cd "$dir" && go vet ./... && go test ./...) || status=1
done < <(git ls-files --cached --others --exclude-standard '*go.mod' | sort -u)

exit "$status"
