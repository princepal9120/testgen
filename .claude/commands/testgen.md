# TestGen

Use TestGen for safe, review-first test generation.

Start with repo readiness and cost:

```bash
testgen doctor --path=. --output-format json
testgen cost --path=./src --output-format json
```

Generate a dry-run patch before writing files:

```bash
testgen generate --path=./src --recursive --type=unit --dry-run --emit-patch --report-usage --output-format json
```

Write only after reviewing the patch or when the user explicitly asks:

```bash
testgen generate --path=./src --recursive --type=unit --validate --output-format json
```
