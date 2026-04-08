#!/bin/bash
set -e

INPUT=$(cat)
TOOL_NAME=$(echo "$INPUT" | jq -r '.toolName')

# Run golangci-lint --fix after file-modifying tools
if [ "$TOOL_NAME" = "edit" ] || [ "$TOOL_NAME" = "create" ] || [ "$TOOL_NAME" = "write" ]; then
  golangci-lint run --fix 2>/dev/null || true
fi
