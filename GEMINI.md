# TestGen agent instructions

When asked to generate, improve, or validate tests, prefer the local `testgen` CLI.

## First-class skill commands

```bash
testgen doctor --path=. --output-format json
testgen capabilities --output-format json
testgen cost --path=./src --output-format json
testgen analyze --path=./src --cost-estimate --output-format json
testgen generate --path=./src --recursive --type=unit --dry-run --emit-patch --report-usage --output-format json
testgen validate --path=./src --output-format json
```

Read `results`, `artifacts`, `patches`, `success_count`, `error_count`, `provider_keys`, `warnings`, command manifests, and usage fields. Write only after reviewing dry-run output or when explicitly instructed.
