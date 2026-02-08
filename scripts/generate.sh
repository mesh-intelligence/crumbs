#!/usr/bin/env bash
#
# Generation loop: call make-work to create tasks, then do-work to execute them.
#
# This is the "generate" phase of the generation lifecycle (see eng02).
# Run this after open-generation.sh and before close-generation.sh.
#
# Usage: generate.sh [options] [repo-root]
#
# Options:
#   --silence-claude       Suppress Claude's output
#   --make-work-limit N    Number of issues to create per cycle (default: 5)
#   --cycles N             Number of make-work/do-work cycles (default: 1)
#
# See docs/engineering/eng02-generation-workflow.md for the full workflow.
#

set -e

# Parse arguments
SILENCE_CLAUDE=false
MAKE_WORK_LIMIT=5
CYCLES=1
REPO_ARG=""

while [[ $# -gt 0 ]]; do
  case $1 in
    --silence-claude)
      SILENCE_CLAUDE=true
      shift
      ;;
    --make-work-limit)
      MAKE_WORK_LIMIT="$2"
      shift 2
      ;;
    --cycles)
      CYCLES="$2"
      shift 2
      ;;
    *)
      REPO_ARG="$1"
      shift
      ;;
  esac
done

REPO_ROOT="${REPO_ARG:-$(dirname "$0")/..}"
cd "$REPO_ROOT" || exit 1
REPO_ROOT=$(pwd)

SCRIPT_DIR="$REPO_ROOT/scripts"

echo ""
echo "========================================"
echo "Generate: $CYCLES cycle(s), $MAKE_WORK_LIMIT issues per cycle"
echo "========================================"
echo ""

for cycle in $(seq 1 "$CYCLES"); do
  echo ""
  echo "========================================"
  echo "Cycle $cycle of $CYCLES"
  echo "========================================"
  echo ""

  # Make work
  echo "--- make-work ---"
  make_work_args="--limit $MAKE_WORK_LIMIT"
  if [ "$SILENCE_CLAUDE" = true ]; then
    make_work_args="$make_work_args --silence-claude"
  fi
  "$SCRIPT_DIR/make-work.sh" $make_work_args

  # Do work
  echo ""
  echo "--- do-work ---"
  do_work_args=""
  if [ "$SILENCE_CLAUDE" = true ]; then
    do_work_args="--silence-claude"
  fi
  "$SCRIPT_DIR/do-work.sh" $do_work_args "$REPO_ROOT"
done

echo ""
echo "========================================"
echo "Generate complete. Ran $CYCLES cycle(s)."
echo "========================================"
