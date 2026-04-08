#!/usr/bin/env bash
set -euo pipefail

ROOT=$(cd "$(dirname "$0")/.." && pwd)
TARGET="${1:-$(pwd)}"
MODE="${2:-copy}"

mkdir -p "$TARGET/.codex/skills/testgen" "$TARGET/.claude/commands" "$TARGET/.opencode/commands"

install_file() {
  local src="$1"
  local dst="$2"
  if [[ "$MODE" == "symlink" ]]; then
    rm -f "$dst"
    ln -s "$src" "$dst"
  else
    cp "$src" "$dst"
  fi
}

install_file "$ROOT/.codex/skills/testgen/SKILL.md" "$TARGET/.codex/skills/testgen/SKILL.md"
install_file "$ROOT/.claude/commands/testgen.md" "$TARGET/.claude/commands/testgen.md"
install_file "$ROOT/.opencode/commands/testgen.md" "$TARGET/.opencode/commands/testgen.md"

cat <<MSG
Installed TestGen agent integrations into: $TARGET
Mode: $MODE

Files:
- .codex/skills/testgen/SKILL.md
- .claude/commands/testgen.md
- .opencode/commands/testgen.md

Recommended safe command:
  testgen generate --file ./path/to/file --type=unit --dry-run --emit-patch --output-format json

Experimental MCP server:
  testgen mcp
MSG
