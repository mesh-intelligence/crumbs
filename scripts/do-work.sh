#!/usr/bin/env bash
#
# Pick the top task from beads and invoke Claude to do the work.
#
# The script handles task picking and reservation. Claude receives a clean
# prompt focused on the work itself, without beads-specific instructions.
#

set -e

cd "${1:-$(dirname "$0")/..}" || exit 1

# Get the top task issue as JSON
issue_json=$(bd ready -n 1 --json --type "task" 2>/dev/null)

if [ -z "$issue_json" ] || [ "$issue_json" = "[]" ]; then
  echo "No tasks available. Run 'bd ready' to see all issues."
  exit 0
fi

# Extract issue fields
issue_id=$(echo "$issue_json" | jq -r '.[0].id // empty')
issue_title=$(echo "$issue_json" | jq -r '.[0].title // empty')
issue_description=$(echo "$issue_json" | jq -r '.[0].description // empty')

if [ -z "$issue_id" ]; then
  echo "Failed to parse issue from beads output."
  exit 1
fi

echo "Picking up task: $issue_id - $issue_title"

# Claim the task
bd update "$issue_id" --status in_progress >/dev/null 2>&1
echo "Task claimed."
echo ""

# Build the prompt for Claude (beads-free)
prompt=$(cat <<EOF
/do-work

## Task: $issue_title

$issue_description

---

Complete this task. When done, commit your changes with a descriptive message that includes the task ID ($issue_id) in the commit message.
EOF
)

# Invoke Claude with the prompt
# --dangerously-skip-permissions: auto-approve all tool use
# --print: non-interactive mode, exit when done
# --verbose: show full turn-by-turn output
exec claude --dangerously-skip-permissions --print --verbose "$prompt"
