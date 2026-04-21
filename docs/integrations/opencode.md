# OpenCode integration

**Scope:** This page covers OpenCode-specific setup. For the shared integration model and safe defaults, start with the [integrations index](./README.md).

TestGen can be wrapped by OpenCode-style agents using the shared JSON CLI contract.

## Install into another repo

### Automatic install

```bash
./scripts/install-agent-integrations.sh /path/to/target-repo copy
```

### Manual install

```bash
mkdir -p /path/to/target-repo/.opencode/commands
cp .opencode/commands/testgen.md /path/to/target-repo/.opencode/commands/testgen.md
```

After that, OpenCode can use the repo-local TestGen command from inside the target repo.

If you upgrade TestGen or the repo-local wrapper asset, re-run the install step so the copied command stays current.

## Recommended invocation

### Review first

```bash
testgen generate --path ./src --recursive --type unit --dry-run --emit-patch --output-format json
```

### Materialize tests

```bash
testgen generate --path ./src --recursive --type unit --validate --output-format json
```

## Contract highlights

- `results`: per-source-file generation outcome
- `artifacts`: generated test artifacts with path and code
- `patches`: structured write operations for agent patch application
- `success_count`, `error_count`: aggregate execution status

## Guidance

Keep wrappers thin. TestGen should remain the source of truth for scanning, generation, and validation orchestration.
