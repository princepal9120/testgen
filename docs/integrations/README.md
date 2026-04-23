# TestGen Integrations

**Scope:** This page is the shared entry point for agent wrappers and MCP clients that want to call TestGen safely and consistently.

## Shared contract

All current integration surfaces rely on the same underlying model:
- a shared application/orchestration layer
- machine-readable JSON output
- review-first dry-run flows before file writes
- optional patch-style artifacts for agent application
- additive usage and cost transparency when reporting is enabled

Current JSON payloads expose the same core concepts across wrappers:

- `target_path`: the resolved path being processed
- `results`: per-source-file generation results
- `artifacts`: generated test artifacts and validation flags
- `patches`: structured write operations when dry-run or patch emission is requested
- `success_count` / `error_count`: aggregate run status
- additive usage/cost fields when the caller enables reporting

Recommended safe default:

```bash
testgen generate --file ./src/utils.py --type=unit --dry-run --emit-patch --report-usage --output-format json
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
- Prefer `testgen analyze --cost-estimate --output-format json` when a wrapper needs provider-aware budget guidance before any API call.
- Keep wrappers thin. TestGen should remain the source of truth for scanning, generation, and validation orchestration.
- When you upgrade the TestGen binary, re-run the wrapper install script in repos that copied the repo-local wrapper files.
- Use the per-tool docs only for the wrapper-specific installation and invocation details.

## Cost-aware preflight pattern

For large repositories or budget-sensitive workflows, prefer this sequence:

1. Run `testgen analyze --cost-estimate --output-format json` to get an offline provider-aware estimate.
2. If the estimate is acceptable, run `testgen generate --dry-run --emit-patch --report-usage --output-format json`.
3. Only enable file writes after the caller has reviewed artifacts and usage data.
