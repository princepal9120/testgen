# RTK Installer Overview for TestGen

**Scope:** High-level feature overview for making TestGen installable through RTK so AI-tool users can enable TestGen once and use it natively inside their preferred agent environment.

## Goal

Turn TestGen into an **RTK-installable agent integration**.

User flow:

```bash
# install once for the target tool
rtk init -g --codex
rtk init -g --claude
rtk init -g --opencode
rtk init -g --gemini

# restart the AI tool
# then use TestGen from that tool's native surface
```

The feature is not just “copy some wrapper files.”
It is a real install experience where RTK installs the correct TestGen integration payload for the chosen AI tool.

## Priority Order

1. **Codex**
2. **Claude Code**
3. **OpenCode**
4. **Gemini CLI**
5. Later: Cursor, Windsurf, Cline / Roo, Kilo Code, Antigravity

## What exists today

Current TestGen payloads:
- Codex skill: `.codex/skills/testgen/SKILL.md`
- Claude Code command: `.claude/commands/testgen.md`
- OpenCode command: `.opencode/commands/testgen.md`
- MCP server: `testgen mcp`

Current helper script:
- `scripts/install-agent-integrations.sh`

Current reality:
- good repo-local support for Codex / Claude Code / OpenCode
- good MCP fallback
- no Gemini CLI payload yet
- no real RTK-first install UX yet

## Recommended model

### Layer 1: canonical TestGen payloads
Each tool gets a thin payload owned by this repo.
Those payloads stay small and only bridge the tool UX to the same TestGen backend contract.

### Layer 2: RTK installer UX
RTK becomes the public install surface that:
- detects the target tool
- installs the right TestGen payload
- supports reinstall on upgrade
- keeps all tools aligned to one backend contract

## Why this is the right feature

- gives users the install experience they actually want
- avoids backend forks per tool
- lets Gemini CLI come next without reworking the core product
- keeps MCP as the fallback transport instead of the only portable option

## Release Order

### Phase 1
- Codex
- Claude Code
- OpenCode
- normalized payloads
- RTK installer contract

### Phase 2
- Gemini CLI
- Gemini docs
- Gemini install path through RTK

### Phase 3
- broader adapters for other tools

## Immediate Next Step

1. define the RTK/TestGen install matrix
2. normalize Codex / Claude Code / OpenCode payloads
3. design Gemini CLI payload
4. define RTK ownership and metadata contract
5. implement `rtk init -g --codex|--claude|--opencode|--gemini`

## Where to go next

- Shared integration docs: [`README.md`](./README.md)
- Codex setup: [`codex.md`](./codex.md)
- Claude Code setup: [`claude-code.md`](./claude-code.md)
- OpenCode setup: [`opencode.md`](./opencode.md)
- MCP usage: [`mcp.md`](./mcp.md)
- Install / upgrade / release posture: [`../release/AGENT_DISTRIBUTION.md`](../release/AGENT_DISTRIBUTION.md)
