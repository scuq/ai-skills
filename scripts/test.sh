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
