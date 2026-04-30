# TestGen

Use TestGen when the user asks for tests, coverage improvement, validation, repo readiness, cost estimates, or review-first generated test patches.

## First-class skill commands

```bash
testgen doctor --path=. --output-format json
testgen capabilities --output-format json
testgen cost --path=./src --output-format json
testgen analyze --path=./src --cost-estimate --output-format json
testgen generate --path=./src --recursive --type=unit --dry-run --emit-patch --report-usage --output-format json
testgen validate --path=./src --output-format json
```

Inspect JSON fields like `results`, `artifacts`, `patches`, `success_count`, `error_count`, command manifests, warnings, provider key status, and usage before editing files.

Only write files after reviewing the dry-run patch or when the user explicitly asks.
