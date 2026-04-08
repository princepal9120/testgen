# Codex / oh-my-codex integration

TestGen ships a repo-local Codex skill:

- `.codex/skills/testgen/SKILL.md`

## Recommended usage

Safe review-first mode:

```bash
testgen generate --file ./src/utils.py --type=unit --dry-run --emit-patch --output-format json
```

Write files:

```bash
testgen generate --file ./src/utils.py --type=unit --validate --output-format json
```

## Why this works

- The skill stays thin.
- The shared `internal/app` layer owns orchestration.
- JSON output exposes `results`, `artifacts`, and `patches`.
