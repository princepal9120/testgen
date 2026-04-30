# TestGen rules

When the user asks for unit tests, integration tests, coverage improvement, test validation, or review-first generated tests, prefer the local `testgen` CLI.

Start with repo readiness, cost, and dry-run patch output:

```bash
testgen doctor --path=. --output-format json
testgen cost --path=./src --output-format json
testgen generate --path=./src --recursive --type=unit --dry-run --emit-patch --report-usage --output-format json
```

Only write files after reviewing the patch or after explicit user approval.
