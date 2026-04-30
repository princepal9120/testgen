---
name: TestGen
description: Analyze this repo and generate review-first tests with TestGen.
invokable: true
---

Use the local `testgen` CLI to generate safe, review-first tests.

## First-class skill commands

```bash
testgen doctor --path=. --output-format json
testgen capabilities --output-format json
testgen cost --path=./src --output-format json
testgen analyze --path=./src --cost-estimate --output-format json
testgen generate --path=./src --recursive --type=unit --dry-run --emit-patch --report-usage --output-format json
testgen validate --path=./src --output-format json
```

Workflow:

1. Confirm `testgen` is available.
2. Run `doctor` and `capabilities` first.
3. Estimate cost before large runs.
4. Generate a dry-run patch.
5. Inspect JSON output before writing files.
6. Validate only after review or explicit user approval.
