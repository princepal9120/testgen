# Codex integration

TestGen installs a repo-local Codex skill for agent-native test generation.

The user-facing flow is simple: install the skill into a repo, then ask Codex to generate review-first tests. Codex uses the local TestGen engine for analysis, dry-run patches, JSON output, and validation.

## Install the engine

Linux/macOS:

```bash
curl -fsSL https://raw.githubusercontent.com/princepal9120/testgen/main/install.sh | bash
```

Go install alternative:

```bash
go install github.com/princepal9120/testgen-cli@latest
```

## Install the Codex skill

From inside the repo where Codex should generate tests:

```bash
curl -fsSL https://raw.githubusercontent.com/princepal9120/testgen/main/scripts/install-agent-skill.sh | bash -s -- --agent codex
```

This creates:

```text
.codex/skills/testgen/SKILL.md
```

Install into another repo:

```bash
curl -fsSL https://raw.githubusercontent.com/princepal9120/testgen/main/scripts/install-agent-skill.sh | bash -s -- --target /path/to/repo --agent codex
```

From a cloned TestGen repo, local copy mode also works:

```bash
./scripts/install-agent-integrations.sh /path/to/target-repo copy
```

## Ask Codex

```text
Use TestGen to analyze this repo and generate review-first unit tests for ./src.
Do not write files until you inspect the dry-run patch.
```

Single file:

```text
Use TestGen to create unit tests for ./src/utils.py.
Start with a dry-run patch, then validate the generated test after review.
```

Cost-aware bulk run:

```text
Use TestGen to estimate generation cost for ./src first.
Then generate review-first patches folder by folder.
```

## Expected Codex behavior

Codex should first run:

```bash
testgen analyze --path=./src --cost-estimate --output-format json
```

Then generate reviewable artifacts:

```bash
testgen generate --path=./src \
  --recursive \
  --type=unit \
  --dry-run \
  --emit-patch \
  --report-usage \
  --output-format json
```

Codex should inspect `results`, `artifacts`, `patches`, and usage data before applying writes.

Write and validate only after review or explicit user instruction:

```bash
testgen generate --path=./src \
  --recursive \
  --type=unit \
  --validate \
  --output-format json
```

## Why this works well for Codex

- The skill is repo-local and discoverable.
- The CLI remains an engine, not the user-facing product.
- Dry-run patches make file writes explicit and reviewable.
- JSON output gives Codex stable fields for reasoning and follow-up edits.

## Troubleshooting

- `testgen: command not found`: install the engine or add `~/.local/bin` to PATH.
- Missing provider key: export one of `ANTHROPIC_API_KEY`, `OPENAI_API_KEY`, `GEMINI_API_KEY`, or `GROQ_API_KEY`.
- Large repo: analyze a narrow path first, then generate per folder.
- Validation failure: inspect generated tests, run the repo-native test command, patch the tests, then rerun validation.
