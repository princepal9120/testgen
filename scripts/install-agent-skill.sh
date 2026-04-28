#!/usr/bin/env bash
set -euo pipefail

REPO="princepal9120/testgen"
REF="main"
TARGET="$(pwd)"
AGENT="all"
MODE="copy"

usage() {
  cat <<'EOF'
Install TestGen agent integrations into a target repo.

Usage:
  install-agent-skill.sh [options]

Options:
  --target <path>       Target repo. Defaults to current directory.
  --agent <name>        all, codex, claude, or opencode. Defaults to all.
  --ref <git-ref>       Git ref to install from. Defaults to main.
  --mode <copy>         Reserved for compatibility. Only copy is supported for remote installs.
  -h, --help            Show help.

Examples:
  curl -fsSL https://raw.githubusercontent.com/princepal9120/testgen/main/scripts/install-agent-skill.sh | bash
  curl -fsSL https://raw.githubusercontent.com/princepal9120/testgen/main/scripts/install-agent-skill.sh | bash -s -- --agent codex
  curl -fsSL https://raw.githubusercontent.com/princepal9120/testgen/main/scripts/install-agent-skill.sh | bash -s -- --target /path/to/repo --agent all
EOF
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --target)
      TARGET="${2:?missing value for --target}"
      shift 2
      ;;
    --agent)
      AGENT="${2:?missing value for --agent}"
      shift 2
      ;;
    --ref)
      REF="${2:?missing value for --ref}"
      shift 2
      ;;
    --mode)
      MODE="${2:?missing value for --mode}"
      shift 2
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      echo "Unknown argument: $1" >&2
      usage >&2
      exit 1
      ;;
  esac
done

if [[ "$MODE" != "copy" ]]; then
  echo "Remote installer only supports --mode copy." >&2
  exit 1
fi

case "$AGENT" in
  all|codex|claude|opencode) ;;
  *)
    echo "Unsupported agent: $AGENT. Use all, codex, claude, or opencode." >&2
    exit 1
    ;;
esac

if ! command -v curl >/dev/null 2>&1; then
  echo "curl is required." >&2
  exit 1
fi

RAW="https://raw.githubusercontent.com/${REPO}/${REF}"
mkdir -p "$TARGET"

install_file() {
  local src="$1"
  local dst="$2"
  mkdir -p "$(dirname "$dst")"
  curl -fsSL "${RAW}/${src}" -o "$dst"
}

if [[ "$AGENT" == "all" || "$AGENT" == "codex" ]]; then
  install_file ".codex/skills/testgen/SKILL.md" "$TARGET/.codex/skills/testgen/SKILL.md"
fi

if [[ "$AGENT" == "all" || "$AGENT" == "claude" ]]; then
  install_file ".claude/commands/testgen.md" "$TARGET/.claude/commands/testgen.md"
fi

if [[ "$AGENT" == "all" || "$AGENT" == "opencode" ]]; then
  install_file ".opencode/commands/testgen.md" "$TARGET/.opencode/commands/testgen.md"
fi

cat <<MSG
Installed TestGen agent integration into: $TARGET
Agent: $AGENT
Ref: $REF

Next:
1. Make sure the TestGen engine is available:
   curl -fsSL https://raw.githubusercontent.com/princepal9120/testgen/main/install.sh | bash

2. Ask your coding agent:
   Use TestGen to analyze this repo and generate review-first unit tests.
MSG
