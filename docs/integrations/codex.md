# Codex integration

TestGen ships a repo-local Codex skill for review-first test generation.

Use it when Codex should inspect a codebase, estimate generation cost, create tests, emit reviewable patches, or validate generated tests without hand-writing the whole workflow from scratch.

## Install TestGen

Linux/macOS:

```bash
curl -fsSL https://raw.githubusercontent.com/princepal9120/testgen/main/install.sh | bash
```

Windows PowerShell:

```powershell
irm https://raw.githubusercontent.com/princepal9120/testgen/main/install.ps1 | iex
```

Go install:

```bash
go install github.com/princepal9120/testgen-cli@latest
```

Verify:

```bash
testgen --version
testgen --help
```

## Install the Codex skill into a repo

From the TestGen source repo:

```bash
./scripts/install-agent-integrations.sh /path/to/target-repo copy
```

This installs:

```text
/path/to/target-repo/.codex/skills/testgen/SKILL.md
```

Manual install:

```bash
mkdir -p /path/to/target-repo/.codex/skills/testgen
cp .codex/skills/testgen/SKILL.md /path/to/target-repo/.codex/skills/testgen/SKILL.md
```

For local wrapper development, use symlinks:

```bash
./scripts/install-agent-integrations.sh /path/to/target-repo symlink
```

If you upgrade TestGen or edit the skill asset, rerun the install command so copied repos stay aligned.

## Recommended Codex workflow

### 1. Analyze first

```bash
testgen analyze --path=./src --cost-estimate --output-format json
```

Use this before generation so Codex can reason about scope, language mix, and estimated provider cost.

### 2. Generate review-first artifacts

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

### 3. Write and validate when approved

```bash
testgen generate --path=./src \
  --recursive \
  --type=unit \
  --validate \
  --output-format json
```

For one file:

```bash
testgen generate --file ./src/utils.py --type=unit --validate --output-format json
```

## Machine request mode

When Codex has a structured request payload:

```bash
cat request.json | testgen generate --request-file=-
```

or:

```bash
testgen generate --request-file=./request.json
```

Machine mode writes the shared JSON envelope to stdout and suppresses human-oriented Cobra banners on stderr.

## Why this works well for Codex

- The skill is thin and procedural.
- TestGen owns scanning, generation, validation, and cost reporting.
- Dry-run patches make file writes explicit and reviewable.
- JSON output gives Codex stable fields for reasoning and follow-up edits.

## Troubleshooting

- `testgen: command not found`: install TestGen or add `~/.local/bin` to PATH.
- Missing provider key: export one of `ANTHROPIC_API_KEY`, `OPENAI_API_KEY`, `GEMINI_API_KEY`, or `GROQ_API_KEY`.
- Large repo: start with `testgen analyze --path=./src --cost-estimate --output-format json`, then generate per folder.
- Validation failure: inspect generated tests, run the repo-native test command, patch the tests, then rerun validation.
