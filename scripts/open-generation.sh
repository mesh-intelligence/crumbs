#!/usr/bin/env bash
#
# Start a new generation (per eng02-generation-workflow).
#
# Tags main, creates a generation branch, deletes Go files, reinitializes
# the Go module. Must be run from main with no existing generation branch.
#
# Usage: open-generation.sh [repo-root]
#
# See docs/engineering/eng02-generation-workflow.md for the full workflow.
#

set -e

REPO_ROOT="${1:-$(dirname "$0")/..}"
cd "$REPO_ROOT" || exit 1
REPO_ROOT=$(pwd)

branch=$(git rev-parse --abbrev-ref HEAD)

if [ "$branch" != "main" ]; then
  echo "Error: must be on main (currently on $branch)."
  exit 1
fi

# Check no existing generation branch
if git branch --list 'generation-*' | grep -q .; then
  echo "Error: a generation branch already exists. Close it first or delete it."
  git branch --list 'generation-*'
  exit 1
fi

gen_name="generation-$(date +%Y-%m-%d-%H-%M)"

echo ""
echo "========================================"
echo "Opening generation: $gen_name"
echo "========================================"
echo ""

# Tag current main
echo "Tagging current state as $gen_name..."
git tag "$gen_name"

# Create and switch to generation branch
echo "Creating branch $gen_name..."
git checkout -b "$gen_name"

# Delete Go source files
echo "Deleting Go source files..."
find . -name '*.go' -not -path './.git/*' -delete 2>/dev/null || true

# Remove empty directories left behind in Go source dirs
for dir in cmd/ pkg/ internal/ tests/; do
  if [ -d "$dir" ]; then
    find "$dir" -type d -empty -delete 2>/dev/null || true
  fi
done

# Remove build artifacts and dependency lock
rm -rf bin/ go.sum

# Reinitialize Go module
echo "Reinitializing Go module..."
rm -f go.mod
go mod init github.com/mesh-intelligence/crumbs

# Commit the clean state
echo "Committing clean state..."
git add -A
git commit -m "Open generation: $gen_name

Delete Go files, reinitialize module.
Tagged previous state as $gen_name."

echo ""
echo "Generation opened on branch $gen_name."
echo "Run do-work.sh to start building."
echo ""
