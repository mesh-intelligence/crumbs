#!/bin/bash
set -e

# Install Claude Code OAuth credentials from a token file.
#
# Mount your .claude-tokens directory and set CLAUDE_TOKEN_FILE
# to the filename you want to use:
#
#   docker run \
#     -v ./.claude-tokens:/claude-tokens:ro \
#     -e CLAUDE_TOKEN_FILE=claude-max.json \
#     -v $(pwd):/workspace \
#     crumbs
#
# The entrypoint copies the token file to ~/.claude/.credentials.json
# where Claude Code reads it on Linux.

TOKENS_DIR="/claude-tokens"

if [ -n "$CLAUDE_TOKEN_FILE" ] && [ -f "$TOKENS_DIR/$CLAUDE_TOKEN_FILE" ]; then
    cp "$TOKENS_DIR/$CLAUDE_TOKEN_FILE" /root/.claude/.credentials.json
    echo "Loaded credentials from $CLAUDE_TOKEN_FILE"
elif [ -f "$TOKENS_DIR/claude-max.json" ]; then
    cp "$TOKENS_DIR/claude-max.json" /root/.claude/.credentials.json
    echo "Loaded credentials from claude-max.json (default)"
fi

exec "$@"
