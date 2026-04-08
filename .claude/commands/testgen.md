---
description: Generate tests with TestGen using the shared JSON contract
---

Use TestGen to generate tests for the requested file or path.

Default safe command:

```bash
testgen generate --file "$ARGUMENTS" --type=unit --dry-run --emit-patch --output-format json
```

If the user explicitly wants files written, use:

```bash
testgen generate --file "$ARGUMENTS" --type=unit --validate --output-format json
```

Prefer the JSON response over parsing text output.
