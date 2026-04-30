#!/usr/bin/env bash
set -euo pipefail

ROOT=$(cd "$(dirname "$0")/.." && pwd)
TARGET="${1:-$(pwd)}"
MODE="${2:-copy}"

mkdir -p \
  "$TARGET/.codex/skills/testgen" \
  "$TARGET/.claude/commands" \
  "$TARGET/.opencode/commands" \
  "$TARGET/.cursor/rules" \
  "$TARGET/.clinerules" \
  "$TARGET/.continue/prompts" \
  "$TARGET/.roo/commands"

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

install_file "$ROOT/skills/testgen/SKILL.md" "$TARGET/.codex/skills/testgen/SKILL.md"
install_file "$ROOT/.claude/commands/testgen.md" "$TARGET/.claude/commands/testgen.md"
install_file "$ROOT/.opencode/commands/testgen.md" "$TARGET/.opencode/commands/testgen.md"
install_file "$ROOT/.cursor/rules/testgen.mdc" "$TARGET/.cursor/rules/testgen.mdc"
install_file "$ROOT/.clinerules/testgen.md" "$TARGET/.clinerules/testgen.md"
install_file "$ROOT/.continue/prompts/testgen.md" "$TARGET/.continue/prompts/testgen.md"
install_file "$ROOT/.roo/commands/testgen.md" "$TARGET/.roo/commands/testgen.md"
if [[ ! -e "$TARGET/GEMINI.md" ]]; then
  install_file "$ROOT/GEMINI.md" "$TARGET/GEMINI.md"
else
  echo "Skipping GEMINI.md because it already exists in target."
fi

cat <<MSG
Installed TestGen agent integrations into: $TARGET
Mode: $MODE

Files:
- .codex/skills/testgen/SKILL.md
- .claude/commands/testgen.md
- .opencode/commands/testgen.md
- .cursor/rules/testgen.mdc
- .clinerules/testgen.md
- .continue/prompts/testgen.md
- .roo/commands/testgen.md
- GEMINI.md (if target did not already have one)

Recommended agent flow:
  testgen analyze --path ./src --cost-estimate --output-format json
  testgen generate --path ./src --recursive --type=unit --dry-run --emit-patch --report-usage --output-format json

Write after review:
  testgen generate --path ./src --recursive --type=unit --validate --output-format json

MCP server:
  testgen mcp
MSG
