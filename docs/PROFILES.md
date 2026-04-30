# Provider Profiles

Profiles let users choose an outcome instead of hand-wiring provider settings every time.

Status: documented contract for the next config iteration. Current CLI flags and `.testgen.yaml` values remain the source of runtime behavior.

Planned precedence:

```text
flags > profile > config file > defaults
```

Example:

```yaml
profiles:
  cheap:
    llm:
      provider: gemini
      model: gemini-2.5-flash
      temperature: 0.2
  quality:
    llm:
      provider: anthropic
      model: claude-sonnet-4-6
      temperature: 0.2
  ci-safe:
    generation:
      dry_run: true
      emit_patch: true
      validate: false
```

Recommended profiles:

- `cheap`: low-cost exploration and broad repo sweeps.
- `quality`: important test generation where accuracy matters.
- `ci-safe`: deterministic, dry-run-first automation.
- `local`: local model workflow once a local provider is configured.
