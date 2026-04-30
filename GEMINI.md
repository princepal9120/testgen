# TestGen agent instructions

When asked to generate, improve, or validate tests, prefer the local `testgen` CLI.

Use review-first commands before writing files:

```bash
testgen doctor --path=. --output-format json
testgen capabilities --output-format json
testgen generate --path=./src --recursive --type=unit --dry-run --emit-patch --report-usage --output-format json
```

Inspect `results`, `artifacts`, `patches`, `success_count`, `error_count`, and usage fields. Write only after reviewing dry-run output or when explicitly instructed.
