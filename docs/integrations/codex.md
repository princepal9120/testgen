# Codex / oh-my-codex integration

**Scope:** This page covers Codex-specific setup. For the shared integration model and safe defaults, start with the [integrations index](./README.md).

TestGen ships one canonical shared skill source for discovery and one repo-local Codex compatibility path:

- Canonical source: `skills/testgen/SKILL.md`
- Codex compatibility path in this repo: `.codex/skills/testgen/SKILL.md`

## Install into another repo

### Automatic install

Preferred portable install:

```bash
./scripts/install-agent-integrations.sh /path/to/target-repo copy
```

Local-development-only symlink install:

```bash
./scripts/install-agent-integrations.sh /path/to/target-repo symlink
```

### Manual install

```bash
mkdir -p /path/to/target-repo/.codex/skills/testgen
cp skills/testgen/SKILL.md /path/to/target-repo/.codex/skills/testgen/SKILL.md
```

After that, invoke the repo-local `testgen` skill from Codex / oh-my-codex inside the target repo.

Inside this repo, `.codex/skills/testgen/SKILL.md` is a compatibility symlink to the canonical `skills/testgen/SKILL.md`. Use `copy` for portable target-repo installs; use `symlink` only when the target repo will stay on the same machine as this checkout. If you upgrade TestGen or the canonical skill asset, re-run the install step so copied target-repo assets stay aligned.

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


## skills.sh publishing

You do not open a manual listing request in `vercel-labs/skills` for this integration. The canonical `skills/testgen/SKILL.md` file should live in this repo, and users install the skill directly from your GitHub repository via the `skills` CLI. `skills.sh` visibility comes from install telemetry, while `.codex/skills/testgen/SKILL.md` remains only the repo-local Codex compatibility path.
