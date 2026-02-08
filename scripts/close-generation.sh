#!/usr/bin/env bash
#
# Close the current generation (per eng02-generation-workflow).
#
# Before merging, deletes Go code from main so the generation's code
# replaces it cleanly. Documentation is preserved on main so that any
# doc changes from the generation merge normally.
#
# Steps:
# 1. Tag the generation branch as closed
# 2. Switch to main
# 3. Delete Go code from main and commit
# 4. Merge the generation branch (brings in new code + doc changes)
# 5. Tag main
# 6. Delete the generation branch
#
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

# Tag the final state of the generation branch
echo "Tagging generation as $closed_tag..."
git tag "$closed_tag"

# Switch to main
echo "Switching to main..."
git checkout main

# Delete Go code from main to prepare for a clean merge.
# Documentation stays so doc changes from the generation merge normally.
echo "Deleting Go code from main..."
find . -name '*.go' -not -path './.git/*' -delete 2>/dev/null || true

for dir in cmd/ pkg/ internal/ tests/; do
  if [ -d "$dir" ]; then
    find "$dir" -type d -empty -delete 2>/dev/null || true
  fi
done

rm -rf bin/ go.sum

# Reinitialize Go module
rm -f go.mod
go mod init github.com/mesh-intelligence/crumbs

git add -A
git commit -m "Prepare main for generation merge: delete Go code

Documentation preserved for merge. Code will be replaced by $branch."

# Merge the generation branch
echo "Merging $branch into main..."
git merge "$branch" --no-edit

# Tag main after merge
main_tag="${branch}-merged"
echo "Tagging main as $main_tag..."
git tag "$main_tag"

# Delete the generation branch
echo "Deleting branch $branch..."
git branch -d "$branch"

echo ""
echo "Generation closed. Work is on main."
echo ""
