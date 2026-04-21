# RTK Installer Implementation Plan for TestGen

**Scope:** Concrete implementation guide for shipping TestGen as an RTK-installable integration for AI coding tools, with Codex, Claude Code, OpenCode, and Gemini CLI as the first supported targets.

## Product Goal

Users should be able to:

1. run one RTK install command
2. restart their AI tool
3. immediately use TestGen from the tool-native surface

Target UX:

```bash
rtk init -g --codex
rtk init -g --claude
rtk init -g --opencode
rtk init -g --gemini
```

## Feature Definition

This feature means:
- TestGen ships the canonical payloads and metadata
- RTK installs those payloads into the correct tool locations
- all tools still use the same TestGen backend contract

## Phase 1 Scope

Support through RTK:
- Codex
- Claude Code
- OpenCode

Prepare Gemini CLI payload and RTK contract so Phase 2 is straightforward.

## Existing Surfaces To Reuse

### Payloads
- Codex skill: `.codex/skills/testgen/SKILL.md`
- Claude Code command: `.claude/commands/testgen.md`
- OpenCode command: `.opencode/commands/testgen.md`

### Installer / distribution helpers
- `scripts/install-agent-integrations.sh`
- `scripts/print-mcp-config.sh`

### Shared backend
- `cmd/generate.go`
- `cmd/mcp.go`
- `internal/app/`
- `internal/mcp/server.go`

### Docs
- `docs/integrations/README.md`
- `docs/integrations/codex.md`
- `docs/integrations/claude-code.md`
- `docs/integrations/opencode.md`
- `docs/integrations/mcp.md`
- `docs/release/AGENT_DISTRIBUTION.md`

## Deliverables

### Deliverable 1: RTK install matrix
Create a single matrix that defines for each target:
- support tier: GA / beta / future
- payload type
- install location
- invocation style after install
- upgrade / reinstall steps
- RTK flag mapping

**New file:**
- `docs/integrations/INSTALL_MATRIX.md`

### Deliverable 2: normalized top-3 payloads
Make sure Codex / Claude Code / OpenCode payloads all:
- use the same review-first invocation
- document explicit write behavior
- rely on the same shared backend contract
- avoid tool-specific business logic

### Deliverable 3: RTK metadata contract
Define what RTK needs from this repo:
- payload source paths
- install destinations
- support level
- reinstall behavior
- version / compatibility metadata

### Deliverable 4: Gemini CLI design
Define Gemini CLI’s native integration shape and how RTK should install it.

**New file:**
- `docs/integrations/gemini-cli.md`

### Deliverable 5: user-facing docs update
Document RTK as the preferred install path for supported tools.

## Implementation Sequence

### Step 1
Create:
- `docs/integrations/INSTALL_MATRIX.md`
- `docs/integrations/gemini-cli.md`

### Step 2
Normalize existing payloads for Codex / Claude Code / OpenCode.

### Step 3
Refactor `scripts/install-agent-integrations.sh` into a reusable payload-install primitive that RTK can either call directly or mirror.

### Step 4
Define RTK metadata ownership:
- does this repo export metadata for RTK?
- or does RTK hardcode TestGen integration behavior?

### Step 5
Implement RTK-first install support for:
- `--codex`
- `--claude`
- `--opencode`

### Step 6
Add Gemini CLI support behind RTK.

### Step 7
Document upgrade / reinstall behavior.

## Verification Checklist

### Codex
- `rtk init -g --codex`
- restart Codex
- TestGen skill available
- safe dry-run works
- explicit write flow works

### Claude Code
- `rtk init -g --claude`
- restart Claude Code
- `/testgen` available
- safe dry-run works
- explicit write flow works

### OpenCode
- `rtk init -g --opencode`
- restart OpenCode
- command available
- safe dry-run works
- explicit write flow works

### Gemini CLI
- `rtk init -g --gemini`
- restart Gemini CLI
- integration available
- safe dry-run works
- explicit write flow works

### Upgrade
- upgrade TestGen binary
- rerun RTK installer
- verify payload still works

## Non-Goals

- supporting every AI tool in the first release
- per-tool backend forks
- replacing MCP as fallback
- adding marketplace integrations before the top four are stable

## Recommendation

Make this a clearly named feature:
**“RTK installer support for TestGen”**

Then execute it in this order:
1. Codex / Claude Code / OpenCode through RTK
2. Gemini CLI through RTK
3. future adapters later
