# Claude Code integration

TestGen ships a repo-local Claude Code command:

- `.claude/commands/testgen.md`

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
