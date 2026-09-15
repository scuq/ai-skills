#!/usr/bin/env bash
# release-notes.sh - print the CHANGELOG.md section of one version.
#
# Usage: scripts/release-notes.sh VERSION
#
# VERSION has the format MAJOR.MINOR.PATCH, without "v".
#
# Exit status:
#   0  The script printed the section.
#   1  CHANGELOG.md has no section for VERSION, or the section is empty.
#   2  VERSION is missing.

set -euo pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/.."

if [[ $# -ne 1 ]]; then
  echo "Usage: scripts/release-notes.sh VERSION" >&2
  exit 2
fi

notes=$(awk -v v="$1" '
  index($0, "## [" v "]") == 1 { found = 1; next }
  found && /^## \[/ { exit }
  found { print }
' CHANGELOG.md)

if [[ -z ${notes//[[:space:]]/} ]]; then
  echo "CHANGELOG.md has no section for $1. Add the section ## [$1] - DATE, commit, and tag again." >&2
  exit 1
fi
printf '%s\n' "$notes"
