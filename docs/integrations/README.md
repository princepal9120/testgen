# TestGen Agent Integrations

TestGen is agent-first. Users install a small repo-local skill or command wrapper, then ask their coding agent to generate review-first tests.

The `testgen` binary is the local engine behind the skill. Agent wrappers should stay thin and let the engine own scanning, generation, validation, JSON output, patches, and usage reporting.

## Install into a repo

From inside the target repo:

```bash
curl -fsSL https://raw.githubusercontent.com/princepal9120/testgen/main/scripts/install-agent-skill.sh | bash
```

Install one integration:

```bash
curl -fsSL https://raw.githubusercontent.com/princepal9120/testgen/main/scripts/install-agent-skill.sh | bash -s -- --agent codex
curl -fsSL https://raw.githubusercontent.com/princepal9120/testgen/main/scripts/install-agent-skill.sh | bash -s -- --agent claude
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

## Engine requirement

The wrappers call the local TestGen engine. Install it once per machine:

```bash
curl -fsSL https://raw.githubusercontent.com/princepal9120/testgen/main/install.sh | bash
```

or:

```bash
go install github.com/princepal9120/testgen-cli@latest
```

## Agent prompt examples

```text
Use TestGen to analyze this repo and generate review-first unit tests for ./src.
Do not write files until you inspect the dry-run patch.
```

```text
Use TestGen to estimate generation cost for ./src, then create dry-run patches for the highest-value missing tests.
```

```text
Use TestGen to generate tests for ./src/utils.py, validate them, and explain any failures before editing.
```

## Shared contract

All integrations use the same contract:

- review-first dry-run flows before file writes
- machine-readable JSON output
- optional patch-style artifacts for agent application
- provider-aware usage and cost transparency
- validation metadata after generated tests run

JSON payloads expose these core fields:

- `target_path`: resolved path being processed
- `results`: per-source-file generation results
- `artifacts`: generated test artifacts and validation metadata
- `patches`: structured write operations
- `success_count` and `error_count`: aggregate run status
- usage and cost fields when reporting is enabled

## Safe default flow

Agents should run this before writing files:

```bash
testgen analyze --path=./src --cost-estimate --output-format json
testgen generate --path=./src --recursive --type=unit --dry-run --emit-patch --report-usage --output-format json
```

Write only after review or explicit user instruction:

```bash
testgen generate --path=./src --recursive --type=unit --validate --output-format json
```

## Choose an integration

- [Codex](./codex.md)
- [Claude Code](./claude-code.md)
- [OpenCode](./opencode.md)
- [MCP](./mcp.md)

## Guidance for wrapper authors

- Keep wrappers thin.
- Do not duplicate TestGen business logic in prompt files.
- Prefer dry-run JSON output for agent reasoning.
- Prefer `testgen analyze --cost-estimate --output-format json` before large generation runs.
- Rerun the installer when the repo-local skill files change.
