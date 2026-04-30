# TestGen

Use TestGen for safe, review-first test generation.

## First-class skill commands

- `testgen doctor --path=. --output-format json` checks repo readiness, frameworks, provider keys, test folders, warnings, and the safest next command.
- `testgen capabilities --output-format json` shows the agent-readable command/language/schema/provider manifest.
- `testgen cost --path=./src --output-format json` estimates generation cost before model calls.
- `testgen analyze --path=./src --cost-estimate --output-format json` analyzes scope and remains compatible with older workflows.
- `testgen generate ... --dry-run --emit-patch --output-format json` creates reviewable patches.
- `testgen validate --path=./src --output-format json` validates generated or existing tests.

## Safe flow

```bash
testgen doctor --path=. --output-format json
testgen capabilities --output-format json
testgen cost --path=./src --output-format json
testgen analyze --path=./src --cost-estimate --output-format json
testgen generate --path=./src --recursive --type=unit --dry-run --emit-patch --report-usage --output-format json
```

Write only after reviewing the patch or when the user explicitly asks:

```bash
testgen generate --path=./src --recursive --type=unit --validate --output-format json
```
