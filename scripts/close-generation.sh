#!/usr/bin/env bash
#
# Close the current generation (per eng02-generation-workflow).
#
# Tags the generation branch as closed, merges to main, deletes the branch.
# Must be run from a generation-* branch.
#
# Usage: close-generation.sh [repo-root]
#
# See docs/engineering/eng02-generation-workflow.md for the full workflow.
#

set -e

REPO_ROOT="${1:-$(dirname "$0")/..}"
cd "$REPO_ROOT" || exit 1
REPO_ROOT=$(pwd)

branch=$(git rev-parse --abbrev-ref HEAD)

if [[ "$branch" != generation-* ]]; then
  echo "Error: must be on a generation branch (currently on $branch)."
  exit 1
fi

closed_tag="${branch}-closed"

echo ""
echo "========================================"
echo "Closing generation: $branch"
echo "========================================"
echo ""

# Tag the final state
echo "Tagging final state as $closed_tag..."
git tag "$closed_tag"

# Switch to main and merge
echo "Switching to main..."
git checkout main

echo "Merging $branch into main..."
git merge "$branch" --no-edit

# Delete the generation branch
echo "Deleting branch $branch..."
git branch -d "$branch"

echo ""
echo "Generation closed. Work is on main."
echo ""
