---
description: Generate tests with TestGen using the shared JSON contract
---

Use TestGen for repository-local test generation.

Safe review-first mode:

```bash
testgen generate --file "$ARGUMENTS" --type=unit --dry-run --emit-patch --output-format json
```

If the user explicitly wants files written:

```bash
testgen generate --file "$ARGUMENTS" --type=unit --validate --output-format json
```

Prefer the JSON payload over parsing terminal text.
