# TestGen

Use TestGen to generate safe, review-first tests.

## First-class skill commands

```bash
testgen doctor --path=. --output-format json
testgen capabilities --output-format json
testgen cost --path=./src --output-format json
testgen analyze --path=./src --cost-estimate --output-format json
testgen generate --path=./src --recursive --type=unit --dry-run --emit-patch --report-usage --output-format json
testgen validate --path=./src --output-format json
```

Inspect the JSON output, command manifest, repo readiness warnings, cost estimate, usage, and patch artifacts before editing files.

Only write files after review or explicit approval.
