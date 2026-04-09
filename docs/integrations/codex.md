# Codex / oh-my-codex integration

**Scope:** This page covers Codex-specific setup. For the shared integration model and safe defaults, start with the [integrations index](./README.md).

TestGen ships a repo-local Codex skill:

- `.codex/skills/testgen/SKILL.md`

## Install into another repo

### Automatic install

```bash
./scripts/install-agent-integrations.sh /path/to/target-repo copy
```

### Manual install

```bash
mkdir -p /path/to/target-repo/.codex/skills/testgen
cp .codex/skills/testgen/SKILL.md /path/to/target-repo/.codex/skills/testgen/SKILL.md
```

After that, invoke the repo-local `testgen` skill from Codex / oh-my-codex inside the target repo.

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
