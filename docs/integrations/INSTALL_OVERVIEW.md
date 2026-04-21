# TestGen AI Tool Install Overview

**Scope:** High-level guide for making TestGen available inside AI coding tools through native skills, commands, or MCP-backed installer flows.

## Goal

Install TestGen once, restart your AI tool, and use TestGen through the tool's native surface.

Target experience:

```bash
# examples of the installer UX we want to support
rtk init -g --codex
rtk init -g --claude
rtk init -g --opencode
rtk init -g --gemini
```

Then, after restart, the user should be able to invoke TestGen from that tool without manually copying files around.

## Priority Order

1. **Codex**
2. **Claude Code**
3. **OpenCode**
4. **Gemini CLI**
5. Later: Cursor, Windsurf, Cline / Roo, Kilo Code, Antigravity

## What exists today

Current first-class repo payloads:

- Codex skill: `.codex/skills/testgen/SKILL.md`
- Claude Code command: `.claude/commands/testgen.md`
- OpenCode command: `.opencode/commands/testgen.md`
- MCP server: `testgen mcp`

Current installer primitive:

- `scripts/install-agent-integrations.sh`

Current install posture:

- good support for repo-local installs
- good support for MCP fallback
- no Gemini CLI integration yet
- no single global installer UX yet

## Recommended packaging model

Use two layers:

### 1. Canonical per-tool payloads

Keep thin repo-owned integration payloads for each tool.
They should only adapt the tool UX to the same TestGen backend contract.

### 2. Installer/bootstrap UX

Expose a simple installer entrypoint that:
- detects the target tool
- installs the right payload to the right location
- supports reinstall on upgrade
- keeps all tools aligned to one backend contract

## Why this approach wins

- avoids backend forks per tool
- keeps Codex / Claude Code / OpenCode stable
- lets Gemini CLI be added cleanly next
- gives users the install story they actually want

## Recommended release order

### Phase 1
- Codex
- Claude Code
- OpenCode
- normalized payloads
- installer UX contract

### Phase 2
- Gemini CLI
- Gemini docs
- Gemini installer path

### Phase 3
- broader adapters for other tools

## First implementation steps

1. Add a support matrix for all target tools
2. Normalize current Codex / Claude Code / OpenCode payloads
3. Design Gemini CLI integration
4. Upgrade the installer from repo-copy helper to a real install surface
5. Define upgrade / reinstall behavior

## Where to go next

- Shared integration docs: [`README.md`](./README.md)
- Codex setup: [`codex.md`](./codex.md)
- Claude Code setup: [`claude-code.md`](./claude-code.md)
- OpenCode setup: [`opencode.md`](./opencode.md)
- MCP usage: [`mcp.md`](./mcp.md)
- Install / upgrade / release posture: [`../release/AGENT_DISTRIBUTION.md`](../release/AGENT_DISTRIBUTION.md)
