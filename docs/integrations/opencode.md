# OpenCode integration

TestGen installs a repo-local OpenCode command for agent-native test generation.

## Install

From inside the repo where OpenCode should generate tests:

```bash
curl -fsSL https://raw.githubusercontent.com/princepal9120/testgen/main/scripts/install-agent-skill.sh | bash -s -- --agent opencode
```

This creates:

```text
.opencode/commands/testgen.md
```

From a cloned TestGen repo, local copy mode also works:

```bash
./scripts/install-agent-integrations.sh /path/to/target-repo copy
```

## Ask OpenCode

```text
Use TestGen to generate review-first unit tests for ./src.
Inspect the dry-run patch before writing files.
```

Single file:

```text
Use TestGen to create tests for ./src/utils.py, then validate after review.
```

## Expected behavior

Review first:

```bash
testgen generate --path ./src --recursive --type unit --dry-run --emit-patch --output-format json
```

Write after review or explicit instruction:

```bash
testgen generate --path ./src --recursive --type unit --validate --output-format json
```

## Contract highlights

- `results`: per-source-file generation outcome
- `artifacts`: generated test artifacts with path and code
- `patches`: structured write operations for agent patch application
- `success_count` and `error_count`: aggregate execution status

## Guidance

Keep wrappers thin. The local TestGen engine owns scanning, generation, validation, patches, and usage reporting.
