---
name: testgen
description: Generate tests with TestGen using the shared JSON contract
---

# TestGen skill

Use TestGen when you need repo-local test generation from source files.

## Purpose

Call the shared TestGen JSON interface instead of hand-writing tests from scratch when:
- the user asks to generate tests for a file or path
- you want a dry-run artifact before editing files
- you want structured patch operations for review

## Commands

### Dry run with structured output

```bash
testgen generate --file "$FILE" --type=unit --dry-run --emit-patch --output-format json
```

### Write files

```bash
testgen generate --file "$FILE" --type=unit --validate --output-format json
```

## Notes

- Prefer `--dry-run --emit-patch --output-format json` when you want to inspect generated artifacts before writing.
- The JSON response includes `artifacts` and `patches` for machine use.
- Keep this skill as a thin wrapper. Do not duplicate TestGen business logic here.
