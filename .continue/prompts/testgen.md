---
name: TestGen
description: Analyze this repo and generate review-first tests with TestGen.
invokable: true
---

Use the local `testgen` CLI to generate safe, review-first tests.

1. Confirm `testgen` is available.
2. Run `testgen doctor --path=. --output-format json`.
3. Estimate cost before large runs.
4. Generate a dry-run patch.
5. Inspect JSON output before writing files.
6. Validate only after review or explicit user approval.
