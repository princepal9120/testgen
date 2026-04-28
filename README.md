<p align="center">
  <img src="website/images/logo.png" alt="TestGen logo" width="140">
</p>

<p align="center">
  <strong>Agent-native test generation for Codex, Claude Code, OpenCode, and MCP</strong>
</p>

<p align="center">
  <a href="https://github.com/princepal9120/testgen/actions"><img src="https://github.com/princepal9120/testgen/actions/workflows/ci.yml/badge.svg" alt="CI"></a>
  <a href="https://github.com/princepal9120/testgen/releases"><img src="https://img.shields.io/github/v/release/princepal9120/testgen" alt="Release"></a>
  <a href="https://www.apache.org/licenses/LICENSE-2.0"><img src="https://img.shields.io/badge/License-Apache%202.0-blue.svg" alt="License: Apache 2.0"></a>
  <a href="https://github.com/princepal9120/testgen"><img src="https://img.shields.io/github/stars/princepal9120/testgen?style=social" alt="GitHub stars"></a>
</p>

<p align="center">
  <a href="#installation">Install</a> &bull;
  <a href="#quick-start">Quick Start</a> &bull;
  <a href="#agent-skills">Agent Skills</a> &bull;
  <a href="#how-it-works">How It Works</a> &bull;
  <a href="docs/integrations/README.md">Docs</a>
</p>

---

TestGen gives coding agents a production-safe test-generation skill.

Install it into a repo, then ask Codex, Claude Code, OpenCode, or an MCP host to analyze the codebase and generate review-first tests. TestGen handles source scanning, existing-test style detection, cost-aware planning, generated test artifacts, patch output, and validation through one agent-friendly workflow.

The public product is the **agent skill**. The `testgen` binary is the local engine that the skill calls behind the scenes.

Supported languages: **JavaScript/TypeScript, Python, Go, Rust, and Java**.

## Why TestGen

Plain agent prompts are good for one-off test files. TestGen is for repeatable, production-grade test generation inside real repos.

| Agent need | What TestGen provides |
|------------|------------------------|
| Generate tests safely | Dry-run first, patch artifacts, explicit write controls |
| Match the repo | Detects nearby test files and adapts to existing framework, fixtures, mocks, naming, and assertions |
| Avoid blind edits | Agents inspect JSON results before touching files |
| Plan cost before API calls | Offline code analysis and provider-aware cost estimates |
| Work across stacks | JS/TS, Python, Go, Rust, and Java adapters |
| Fit agent workflows | Codex skill, Claude command, OpenCode command, and MCP server |
| Keep logic consistent | One shared engine across every agent integration |

### Why not just ask Claude or Codex directly?

You still use Claude, Codex, or OpenCode. TestGen gives them a dedicated workflow instead of a loose prompt:

- analyze first, so the agent knows what files and functions exist
- reuse existing tests as style context, so output fits the repo
- dry-run patches before writes, so changes are reviewable
- report token and cost usage, so large repos stay controlled
- validate generated tests through the local project toolchain

## Installation

### 1. Install the TestGen engine

Linux/macOS:

```bash
curl -fsSL https://raw.githubusercontent.com/princepal9120/testgen/main/install.sh | bash
```

Windows PowerShell:

```powershell
irm https://raw.githubusercontent.com/princepal9120/testgen/main/install.ps1 | iex
```

Go install alternative:

```bash
go install github.com/princepal9120/testgen-cli@latest
```

### 2. Install the agent skill into your repo

From inside the repo you want your agent to work on:

```bash
curl -fsSL https://raw.githubusercontent.com/princepal9120/testgen/main/scripts/install-agent-skill.sh | bash
```

If you cloned the repo, you can use the shorter local entrypoint:

```bash
./skills.sh --agent all
```

Codex only:

```bash
curl -fsSL https://raw.githubusercontent.com/princepal9120/testgen/main/scripts/install-agent-skill.sh | bash -s -- --agent codex
```

Claude Code only:

```bash
curl -fsSL https://raw.githubusercontent.com/princepal9120/testgen/main/scripts/install-agent-skill.sh | bash -s -- --agent claude
```

OpenCode only:

```bash
curl -fsSL https://raw.githubusercontent.com/princepal9120/testgen/main/scripts/install-agent-skill.sh | bash -s -- --agent opencode
```

Install into another repo:

```bash
curl -fsSL https://raw.githubusercontent.com/princepal9120/testgen/main/scripts/install-agent-skill.sh | bash -s -- --target /path/to/repo --agent all
```

Installed files:

- `.codex/skills/testgen/SKILL.md`
- `.claude/commands/testgen.md`
- `.opencode/commands/testgen.md`

### 3. Set one provider key

```bash
export ANTHROPIC_API_KEY="..."
# or OPENAI_API_KEY / GEMINI_API_KEY / GROQ_API_KEY
```

## Quick Start

Ask your coding agent:

```text
Use TestGen to analyze this repo and generate review-first unit tests for ./src.
Do not write files until you inspect the dry-run patch.
```

For a single file:

```text
Use TestGen to create unit tests for ./src/utils.py.
Start with a dry-run patch, then validate the generated test after review.
```

For a larger repo:

```text
Use TestGen to estimate test generation cost for ./src first.
Then generate review-first patches folder by folder.
```

The agent skill will run the safe flow:

```bash
testgen analyze --path=./src --cost-estimate --output-format json
testgen generate --path=./src --recursive --type=unit --dry-run --emit-patch --report-usage --output-format json
```

Then it writes only when approved or explicitly requested:

```bash
testgen generate --path=./src --recursive --type=unit --validate --output-format json
```

## Agent Skills

### Codex

TestGen installs a repo-local skill at:

```text
.codex/skills/testgen/SKILL.md
```

Codex uses this skill when you ask for AI-generated tests, dry-run patches, test coverage improvement, or validation.

### Claude Code

TestGen installs a command wrapper at:

```text
.claude/commands/testgen.md
```

Use it when you want Claude Code to run the same review-first generation flow.

### OpenCode

TestGen installs a command wrapper at:

```text
.opencode/commands/testgen.md
```

Use it for the same agent-safe workflow in OpenCode.

### MCP

Run the MCP server for MCP-compatible hosts:

```bash
testgen mcp
```

Print a config snippet from the TestGen repo:

```bash
./scripts/print-mcp-config.sh testgen
```

## How It Works

```
Coding agent
     |
     v
TestGen skill or command wrapper
     |
     v
Local TestGen engine
     |
     +--> analyze code and estimate cost
     +--> generate dry-run test artifacts
     +--> emit structured patches
     +--> validate generated tests when writing is allowed
     |
     v
Agent reviews JSON, explains the patch, then applies changes
```

Four principles guide the workflow:

1. **Agent-first onboarding**. Users install a skill into their repo and talk to their agent.
2. **Review before write**. Dry-run and patch artifacts come before file edits.
3. **Machine-readable by default**. Agents get JSON results, artifacts, patches, usage, and errors.
4. **One engine, many agents**. Codex, Claude Code, OpenCode, and MCP all use the same behavior.

## Upgrade

Update the local engine:

```bash
curl -fsSL https://raw.githubusercontent.com/princepal9120/testgen/main/install.sh | bash
```

Refresh the repo-local agent skill:

```bash
curl -fsSL https://raw.githubusercontent.com/princepal9120/testgen/main/scripts/install-agent-skill.sh | bash
```

## Documentation

- [Agent integrations](docs/integrations/README.md)
- [Codex integration](docs/integrations/codex.md)
- [Claude Code integration](docs/integrations/claude-code.md)
- [OpenCode integration](docs/integrations/opencode.md)
- [MCP integration](docs/integrations/mcp.md)
- [Architecture](docs/ARCHITECTURE.md)
- [Release and distribution guide](docs/release/AGENT_DISTRIBUTION.md)

## License

Apache 2.0. See [LICENSE](LICENSE).
