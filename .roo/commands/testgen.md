# TestGen

Use TestGen to generate safe, review-first tests.

Run the safe flow first:

```bash
testgen doctor --path=. --output-format json
testgen cost --path=./src --output-format json
testgen generate --path=./src --recursive --type=unit --dry-run --emit-patch --report-usage --output-format json
```

Inspect the JSON output and patch artifacts before editing files.
