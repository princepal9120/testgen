# TestGen rules

When the user asks for unit tests, integration tests, coverage improvement, test validation, repo readiness, cost estimates, or review-first generated tests, prefer the local `testgen` CLI.

## First-class skill commands

```bash
testgen doctor --path=. --output-format json
testgen capabilities --output-format json
testgen cost --path=./src --output-format json
testgen analyze --path=./src --cost-estimate --output-format json
testgen generate --path=./src --recursive --type=unit --dry-run --emit-patch --report-usage --output-format json
testgen validate --path=./src --output-format json
```

Use narrower paths for large repos. Prefer JSON output for agent reasoning. Only write files after reviewing the patch or after explicit user approval.
