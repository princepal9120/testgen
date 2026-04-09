# Claude Code integration

**Scope:** This page covers Claude Code-specific setup. For the shared integration model and safe defaults, start with the [integrations index](./README.md).

TestGen ships a repo-local Claude Code command:

- `.claude/commands/testgen.md`

## Install into another repo

### Automatic install

```bash
./scripts/install-agent-integrations.sh /path/to/target-repo copy
```

### Manual install

```bash
mkdir -p /path/to/target-repo/.claude/commands
cp .claude/commands/testgen.md /path/to/target-repo/.claude/commands/testgen.md
```

After that, Claude Code can use the repo-local `/testgen` command from inside the target repo.

## Recommended usage

Default safe mode:

```bash
testgen generate --file "$ARGUMENTS" --type=unit --dry-run --emit-patch --output-format json
```

Materialize tests:

```bash
testgen generate --file "$ARGUMENTS" --type=unit --validate --output-format json
```

## Notes

- Prefer JSON output over parsing terminal text.
- Use dry-run first when the agent should inspect generated tests before writing.
