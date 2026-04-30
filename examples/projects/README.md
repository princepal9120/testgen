# TestGen Example Projects

These tiny projects show how TestGen should behave across supported language families.

Each project is intentionally small and includes:

- source file
- existing test-style sample
- `.testgen/request.json`
- `golden/generated.patch`
- `golden/generate-response.json`
- `scripts/validate.sh`

Start with:

```bash
testgen doctor --path=examples/projects/go-testing --output-format json
testgen generate --path=examples/projects/go-testing/src --dry-run --emit-patch --output-format json
```
