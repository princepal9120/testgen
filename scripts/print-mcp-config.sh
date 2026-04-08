#!/usr/bin/env bash
set -euo pipefail

BIN_PATH="${1:-testgen}"
cat <<JSON
{
  "mcpServers": {
    "testgen": {
      "command": "$BIN_PATH",
      "args": ["mcp"]
    }
  }
}
JSON
