# Future Project Workflow

Use TestGen as the default test-generation skill whenever starting or onboarding a new codebase.

## Install into any repo

From the target project root:

```bash
curl -fsSL https://raw.githubusercontent.com/princepal9120/testgen/main/scripts/install-agent-skill.sh | bash -s -- --agent all
```

Or from a local TestGen checkout:

```bash
/path/to/testgen/skills.sh --target /path/to/project --agent all
```

This installs:

- `.codex/skills/testgen/SKILL.md`
- `.claude/commands/testgen.md`
- `.opencode/commands/testgen.md`

## Agent prompt to use

```text
Use TestGen to analyze this repo first, estimate cost, then generate review-first unit test patches for the highest-impact untested files. Do not write files until the dry-run patch is inspected. Validate after writing.
```

## Default workflow

```bash
testgen analyze --path=. --cost-estimate --output-format json
testgen generate --path=. --recursive --type=unit --dry-run --emit-patch --report-usage --output-format json
```

After review:

```bash
testgen generate --path=. --recursive --type=unit --validate --output-format json
```

## When to use it

- New projects with little or no test coverage.
- Existing repos before refactors.
- PRs that need regression tests.
- Agent coding sessions where tests should be generated safely.
- Coverage improvement passes by folder or feature.

## What to avoid

- Do not run bulk writes before analysis.
- Do not force unsupported languages.
- Do not skip validation when the repo has a runnable test command.
- Do not generate tests for the whole repo at once if the codebase is large. Start folder by folder.
