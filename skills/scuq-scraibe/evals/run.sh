#!/usr/bin/env bash
# run.sh - run the evaluation tasks of the scuq-scraibe skill.
#
# Usage: evals/run.sh [--model MODEL] [--only ID] [--without-skill]
#
# MODEL is a model alias or name for claude. The default is sonnet, the
# model of the agent.
#
# For each task in evals.json, the script copies the fixture to a
# temporary directory, runs its setup.sh when it exists, runs the query
# with `claude -p`, and then runs the checks of the task in that
# directory. The script prints one line per check and one line per task.
# The output of claude goes to OUT/ID.log, where OUT is a temporary
# directory that the script names at the end.
#
# claude runs with a temporary configuration directory. It links the
# credentials, the settings, and the plugins of the real one, and it
# holds this repository's copy of the skill and the agent in place of
# the installed ones. With --without-skill it holds
# no skill and no agent, which gives the baseline that the skill is
# measured against.
#
# The checks test shape. Read OUT/ID.log for meaning.
#
# Exit status:
#   0  Every check of every task passed.
#   1  A check failed, or claude failed.
#   2  A tool is missing, or an option is not correct.

set -uo pipefail
cd "$(dirname "${BASH_SOURCE[0]}")"

model=sonnet
only=""
with_skill=1
while [[ $# -gt 0 ]]; do
  case $1 in
    --model) model=$2; shift 2 ;;
    --only) only=$2; shift 2 ;;
    --without-skill) with_skill=0; shift ;;
    *) echo "Usage: evals/run.sh [--model MODEL] [--only ID] [--without-skill]" >&2; exit 2 ;;
  esac
done
for tool in claude python3 git jq; do
  if ! command -v "$tool" >/dev/null; then
    echo "$tool is not installed." >&2
    exit 2
  fi
done

export LINTER
LINTER=$(cd ../reference && pwd)/ste-lint.py
skill_dir=$(cd .. && pwd)
repo_dir=$(cd ../../.. && pwd)
out=$(mktemp -d)
failed=0

# The temporary configuration directory for claude. It links the
# credentials, the settings, and the plugins of the real one, and it
# copies .claude.json, so that the session state of the runs stays in
# the temporary directory.
real_config=${CLAUDE_CONFIG_DIR:-$HOME/.claude}
config=$out/config
mkdir -p "$config/skills" "$config/agents"
for name in .credentials.json settings.json settings.local.json plugins ide; do
  if [[ -e $real_config/$name ]]; then
    ln -s "$real_config/$name" "$config/$name"
  fi
done
for candidate in "$real_config/.claude.json" "$HOME/.claude.json"; do
  if [[ -f $candidate ]]; then
    cp "$candidate" "$config/.claude.json"
    break
  fi
done
if [[ $with_skill -eq 1 ]]; then
  ln -s "$skill_dir" "$config/skills/scuq-scraibe"
  ln -s "$repo_dir/agents/scuq-scraibe.md" "$config/agents/scuq-scraibe.md"
  echo "skill:  $skill_dir"
else
  echo "skill:  none (baseline)"
fi
echo "model:  $model"
export CLAUDE_CONFIG_DIR=$config

while IFS=$'\t' read -r id fixture query; do
  if [[ -n $only && $only != "$id" ]]; then
    continue
  fi
  work=$out/$id
  cp -r "$fixture" "$work"
  (
    cd "$work" || exit 2
    if [[ -x setup.sh ]]; then
      ./setup.sh
    elif [[ ! -d .git ]]; then
      git init -q && git add -A && git -c user.name=eval -c user.email=eval@cyberdyne.example commit -q -m "fixture"
    fi
    # The query stands before the options, because --allowedTools takes
    # every following word. claude reads standard input, which is the
    # task list of the loop without the redirection.
    claude -p "$query" --permission-mode acceptEdits --model "$model" \
      --allowedTools "Read Edit Write Glob Grep Bash(git *) Bash(gofmt *) Bash(go *) Bash(python3 *) Bash(python *) Bash(ls *) Bash(cat *)" \
      >"$out/$id.log" 2>&1 </dev/null
    rc=$?
    if [[ $rc -ne 0 ]]; then
      echo "FAIL $id: claude exited with $rc, see $out/$id.log" >&2
      exit 1
    fi
  ) || { failed=1; continue; }
  task_failed=0
  while IFS= read -r check; do
    if (cd "$work" && bash -c "$check" >/dev/null 2>&1); then
      echo "  ok    $check"
    else
      echo "  FAIL  $check"
      task_failed=1
    fi
  done < <(jq -r --arg id "$id" '.evals[] | select(.id == $id) | .checks[]' evals.json)
  if [[ $task_failed -eq 0 ]]; then
    echo "PASS $id"
  else
    echo "FAIL $id"
    failed=1
  fi
done < <(jq -r '.evals[] | [.id, .fixture, .query] | @tsv' evals.json)

echo "Output: $out"
exit $failed
