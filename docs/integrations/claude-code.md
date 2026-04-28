# Claude Code integration

TestGen installs a repo-local Claude Code command for agent-native test generation.

## Install

From inside the repo where Claude Code should generate tests:

```bash
curl -fsSL https://raw.githubusercontent.com/princepal9120/testgen/main/scripts/install-agent-skill.sh | bash -s -- --agent claude
```

This creates:

```text
.claude/commands/testgen.md
```

From a cloned TestGen repo, local copy mode also works:

```bash
./scripts/install-agent-integrations.sh /path/to/target-repo copy
```

## Ask Claude Code

```text
/testgen ./src/utils.py
```

Or ask in plain language:

```text
Use TestGen to generate review-first unit tests for ./src.
Inspect the dry-run patch before writing files.
```

## Expected behavior

The command wrapper should keep TestGen as the source of truth and use the safe flow:

```bash
testgen generate --file "$ARGUMENTS" --type=unit --dry-run --emit-patch --output-format json
```

Write only after review or explicit instruction:

```bash
testgen generate --file "$ARGUMENTS" --type=unit --validate --output-format json
```

## Notes

- Prefer JSON output over terminal text parsing.
- Use dry-run first when the agent should inspect generated tests before writing.
- Keep the wrapper thin. The local TestGen engine owns scanning, generation, validation, patches, and usage reporting.
