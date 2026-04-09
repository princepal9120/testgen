# TestGen Integrations

**Scope:** This page is the shared entry point for agent wrappers and MCP clients that want to call TestGen safely and consistently.

## Shared contract

All current integration surfaces rely on the same underlying model:
- a shared application/orchestration layer
- machine-readable JSON output
- review-first dry-run flows before file writes
- optional patch-style artifacts for agent application

Recommended safe default:

```bash
testgen generate --file ./src/utils.py --type=unit --dry-run --emit-patch --output-format json
```

When you want TestGen to materialize files and validate them:

```bash
testgen generate --file ./src/utils.py --type=unit --validate --output-format json
```

## Choose an integration

- [`codex.md`](./codex.md) — Codex / oh-my-codex setup
- [`claude-code.md`](./claude-code.md) — Claude Code setup
- [`opencode.md`](./opencode.md) — OpenCode setup
- [`mcp.md`](./mcp.md) — direct MCP stdio usage

## Guidance

- Prefer **dry-run + JSON output** when an agent should inspect artifacts before writing.
- Keep wrappers thin. TestGen should remain the source of truth for scanning, generation, and validation orchestration.
- Use the per-tool docs only for the wrapper-specific installation and invocation details.
