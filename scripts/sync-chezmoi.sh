#!/usr/bin/env bash
# sync-chezmoi.sh - copy the skills and agents into the chezmoi source directory.
#
# Usage: scripts/sync-chezmoi.sh [--dry-run]
#
# The script copies:
#   skills/<name>/  to  <chezmoi source>/dot_claude/skills/<name>/
#   agents/<name>.md  to  <chezmoi source>/dot_claude/agents/<name>.md
#   VERSION  to  <chezmoi source>/dot_claude/dot_ai-skills-version
#
# The script copies only files with different content. It removes files
# from a skill directory in the chezmoi source when the files are not in
# the same skill directory here. It does not touch skills or agents that
# are not in this repository.
#
# The script does not commit. It does not run chezmoi apply.
#
# Exit status:
#   0  The copy or the dry run completed.
#   1  A necessary tool, directory, or file is missing, or a file name is not safe for chezmoi.
#   2  A command-line argument is not correct.

set -euo pipefail
shopt -s nullglob

repo=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)

dry_run=0
case "${1:-}" in
  "") ;;
  --dry-run) dry_run=1 ;;
  *)
    echo "Usage: scripts/sync-chezmoi.sh [--dry-run]" >&2
    exit 2
    ;;
esac

for tool in chezmoi rsync; do
  if ! command -v "$tool" >/dev/null; then
    echo "The $tool command was not found. Install $tool and run the script again." >&2
    exit 1
  fi
done

if [[ ! -f $repo/VERSION ]]; then
  echo "The file $repo/VERSION does not exist. Create it and run the script again." >&2
  exit 1
fi

target="$(chezmoi source-path)/dot_claude"
if [[ ! -d $target ]]; then
  echo "The directory $target does not exist. Add ~/.claude to chezmoi and run the script again." >&2
  exit 1
fi

# chezmoi reads a leading dot and the executable bit as attributes of the
# source file name. A plain copy loses them, so stop before the copy.
unsafe=$(find "$repo/skills" "$repo/agents" \( -name '.*' -o \( -type f -perm -u+x \) \) -print)
if [[ -n $unsafe ]]; then
  echo "These files have a leading dot or the executable bit, and a plain copy to chezmoi loses them:" >&2
  echo "$unsafe" >&2
  echo "Rename the files or remove the executable bit, and run the script again." >&2
  exit 1
fi

version=$(<"$repo/VERSION")
echo "ai-skills $version -> $target"

# -c compares content, not timestamps. -i prints one line for each changed file.
rsync_opts=(-c -i)
if ((dry_run)); then
  rsync_opts+=(--dry-run)
else
  mkdir -p "$target/skills" "$target/agents"
fi

for skill in "$repo"/skills/*/; do
  rsync "${rsync_opts[@]}" -r --delete "$skill" "$target/skills/$(basename "$skill")/"
done

for agent in "$repo"/agents/*.md; do
  rsync "${rsync_opts[@]}" "$agent" "$target/agents/"
done

rsync "${rsync_opts[@]}" "$repo/VERSION" "$target/dot_ai-skills-version"

if ((dry_run)); then
  echo "Dry run. No file changed."
  exit 0
fi

cat <<EOF
Next steps:
  chezmoi diff
  chezmoi apply
  chezmoi git -- add dot_claude
  chezmoi git -- commit -m "Update ai-skills to $version"
  chezmoi git -- push
EOF
