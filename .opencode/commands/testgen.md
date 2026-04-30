# TestGen

Use TestGen when the user asks for tests, coverage improvement, validation, or review-first generated test patches.

Safe flow:

```bash
testgen doctor --path=. --output-format json
testgen generate --path=./src --recursive --type=unit --dry-run --emit-patch --report-usage --output-format json
```

Inspect JSON fields like `results`, `artifacts`, `patches`, `success_count`, `error_count`, and usage before editing files.
