# OpenCode integration

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
