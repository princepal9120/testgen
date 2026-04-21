# TestGen AI Tool Install Implementation Guide

**Scope:** Concrete implementation plan for shipping TestGen as an installable integration for AI coding tools, with Codex, Claude Code, OpenCode, and Gemini CLI as the first supported targets.

## Product Goal

Users should be able to:

1. install TestGen for their AI tool with one command
2. restart the tool
3. immediately use TestGen from the tool-native surface

Target UX:

```bash
rtk init -g --codex
rtk init -g --claude
rtk init -g --opencode
rtk init -g --gemini
```

## Phase 1 Scope

Ship first-class support for:
- Codex
- Claude Code
- OpenCode

Prepare Gemini CLI integration design and payload contract so Phase 2 is straightforward.

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

## Concrete Deliverables

### Deliverable 1: install matrix
Create a single support matrix that defines for each target:
- support tier: GA / beta / future
- payload type: skill / command / plugin / MCP config
- install location
- invocation style after install
- upgrade/reinstall steps

**New file:**
- `docs/integrations/INSTALL_MATRIX.md`

### Deliverable 2: normalized payload contract
Normalize Codex / Claude Code / OpenCode payloads so they all:
- use the same safe review-first invocation
- document explicit write behavior
- describe the same JSON contract assumptions
- stay thin and avoid custom business logic

**Files:**
- `.codex/skills/testgen/SKILL.md`
- `.claude/commands/testgen.md`
- `.opencode/commands/testgen.md`

### Deliverable 3: installer abstraction
Refactor the current installer helper into a reusable primitive that can support both repo-local installs and future global installs.

**Files:**
- `scripts/install-agent-integrations.sh`
- maybe a new global bootstrap helper if `rtk` lives in this repo

### Deliverable 4: Gemini CLI integration design
Define Gemini CLI’s integration model before implementation.
Possible shapes:
- command file
- skill-like asset
- plugin manifest
- MCP-first configuration

**New file:**
- `docs/integrations/gemini-cli.md`

### Deliverable 5: public installer UX contract
Decide whether:
- `rtk` lives inside this repo, or
- `rtk` is external and consumes metadata from this repo

That decision changes where installer code lives, but not the shared payload contract.

## Implementation Sequence

### Step 1
Create:
- `docs/integrations/INSTALL_MATRIX.md`
- `docs/integrations/gemini-cli.md`

### Step 2
Normalize the existing top-3 payloads.

### Step 3
Refactor `scripts/install-agent-integrations.sh` into a payload-install primitive.

### Step 4
Define the `rtk init -g --<tool>` contract and ownership.

### Step 5
Implement Codex / Claude Code / OpenCode install flow through that contract.

### Step 6
Add Gemini CLI support.

### Step 7
Document upgrade / reinstall behavior clearly.

## Verification Checklist

### Codex
- install on clean profile
- restart Codex
- TestGen skill available
- safe dry-run works
- explicit write flow works

### Claude Code
- install on clean profile
- restart Claude Code
- `/testgen` available
- safe dry-run works
- explicit write flow works

### OpenCode
- install on clean profile
- restart OpenCode
- command available
- safe dry-run works
- explicit write flow works

### Gemini CLI
- install on clean profile
- restart Gemini CLI
- integration surface available
- safe dry-run works
- explicit write flow works

### Upgrade
- upgrade TestGen binary
- rerun installer
- verify payload remains aligned

## Non-Goals

- shipping every AI tool in the first release
- custom backend behavior per tool
- replacing MCP as fallback
- adding marketplace integrations before the top four are stable

## Recommendation

Do not start by building a broad plugin system.
Start by making the existing top-3 payloads installable and consistent, then add Gemini CLI as the next supported target.
