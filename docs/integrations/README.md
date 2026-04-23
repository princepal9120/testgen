# TestGen Integrations

**Scope:** This page is the shared entry point for agent wrappers and MCP clients that want to call TestGen safely and consistently.

## Shared contract

All current integration surfaces rely on the same underlying model:
- a shared application/orchestration layer
- machine-readable JSON output
- review-first dry-run flows before file writes
- optional patch-style artifacts for agent application

Current JSON payloads expose the same core concepts across wrappers:

- `target_path`: the resolved path being processed
- `results`: per-source-file generation results
- `artifacts`: generated test artifacts and validation flags
- `patches`: structured write operations when dry-run or patch emission is requested
- `success_count` / `error_count`: aggregate run status

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

- The canonical shared TestGen skill for skills.sh-style discovery lives at `skills/testgen/SKILL.md`.
- The repo-local Codex path `.codex/skills/testgen/SKILL.md` exists only as a compatibility symlink in this repo.
- Prefer **dry-run + JSON output** when an agent should inspect artifacts before writing.
- Keep wrappers thin. TestGen should remain the source of truth for scanning, generation, and validation orchestration.
- When you upgrade the TestGen binary, re-run the wrapper install script in repos that copied the repo-local wrapper files.
- Use the per-tool docs only for the wrapper-specific installation and invocation details.


## skills.sh publishing note

You do not manually submit this skill to `vercel-labs/skills`. That repository hosts the CLI/tooling. To publish the TestGen skill, keep `skills/testgen/SKILL.md` in this repo, push the repo to GitHub, and let users install it directly with the `skills` CLI, for example:

```bash
npx skills add https://github.com/princepal9120/testgen --skill testgen
```

Listing visibility on `skills.sh` comes from anonymous install telemetry in the `skills` CLI, not from a manual registry request.
