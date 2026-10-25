#!/usr/bin/env bash
# setup.sh - build the git history of the changelog fixture.
#
# Usage: ./setup.sh
#
# The script creates the tag v1.0.0 with the old fetch.py, then one
# commit "fix timeout" that changes the default timeout, the exit code
# for a timeout, and the log line. The commit subject hides three
# user-visible changes. The final files are the files of this directory.
#
# Exit status: 0 on success, 1 when git fails.
set -euo pipefail
cd "$(dirname "${BASH_SOURCE[0]}")"
git_c() { git -c user.name=eval -c user.email=eval@cyberdyne.example "$@"; }
rm -rf .git
git init -q
cp fetch.py fetch.py.new
python3 - <<'PY'
import re, pathlib
p = pathlib.Path("fetch.py")
s = p.read_text()
s = s.replace("# 10 seconds is the longest wait that the operators accepted in the review.\nDEFAULT_TIMEOUT = 10", "DEFAULT_TIMEOUT = 30")
s = s.replace("        return 3\n", "        return 1\n")
s = s.replace('    print(f"{args.out} {time.time() - start:.1f}s")', '    print(args.out)')
p.write_text(s)
PY
git add CHANGELOG.md fetch.py
git_c commit -q -m "Release 1.0.0"
git_c tag -a v1.0.0 -m v1.0.0
mv fetch.py.new fetch.py
git add fetch.py
git_c commit -q -m "fix timeout"
rm -f setup.sh
