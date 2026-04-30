# LLM Agent Adoption Roadmap

TestGen should be useful to people who already use Codex, Claude Code, OpenCode, Cursor, Cline, Continue, Roo Code, Gemini CLI, or any MCP-compatible LLM host.

The goal is simple: make test generation feel safer, more repeatable, and easier to trust than asking an LLM with a one-off prompt.

## Product promise

TestGen is the local test-generation engine for coding agents.

A good LLM integration should let the agent:

1. inspect the repo before generating tests
2. estimate cost and scope before calling a model
3. generate reviewable patches before writing files
4. match the repo's existing test framework and style
5. validate generated tests with the project's own toolchain
6. explain what changed in machine-readable output

## Highest-impact enhancements

### 1. First-class agent manifests

Add copy-paste installation and config files for popular LLM tools:

- Codex skill
- Claude Code command
- OpenCode command
- MCP server config
- Cursor rules
- Cline custom instructions
- Continue slash command
- Roo Code mode/instructions
- Gemini CLI prompt pack

Why this matters: people should not have to translate TestGen into their agent's format manually.

Install targets are exposed through the repo installer:

```bash
curl -fsSL https://raw.githubusercontent.com/princepal9120/testgen/main/scripts/install-agent-skill.sh | bash -s -- --agent cursor
curl -fsSL https://raw.githubusercontent.com/princepal9120/testgen/main/scripts/install-agent-skill.sh | bash -s -- --agent cline
curl -fsSL https://raw.githubusercontent.com/princepal9120/testgen/main/scripts/install-agent-skill.sh | bash -s -- --agent continue
curl -fsSL https://raw.githubusercontent.com/princepal9120/testgen/main/scripts/install-agent-skill.sh | bash -s -- --agent roo
curl -fsSL https://raw.githubusercontent.com/princepal9120/testgen/main/scripts/install-agent-skill.sh | bash -s -- --agent all
```

### 2. Capability manifest for agents

Expose one stable JSON command that tells an agent exactly what TestGen can do.

Already started:

```bash
testgen languages --output-format=json
```

Now available:

```bash
testgen capabilities --output-format=json
```

The response should include:

- supported languages
- supported frameworks
- available commands and flags
- output schema version
- dry-run and write behavior
- validation support per language
- known limitations

Why this matters: agents can adapt instead of guessing.

### 3. Repo onboarding report

TestGen includes a command that tells users whether their repo is ready for agent-native test generation.

```bash
testgen doctor --path=. --output-format=json
```

Checks:

- detected languages and frameworks
- existing test directories
- native test command candidates
- missing provider API keys
- ignored/generated directories
- unsupported files
- suggested safe first command

Why this matters: new users need confidence before running generation.

### 4. Framework detection per language

Move beyond language detection into framework detection.

Examples:

- JavaScript/TypeScript: Jest, Vitest, Mocha, Playwright
- Python: pytest, unittest
- Go: testing, testify
- Rust: cargo test
- Java: JUnit, TestNG, Maven, Gradle
- C#: xUnit, NUnit, MSTest
- PHP: PHPUnit, Pest
- Ruby: RSpec, Minitest
- C++: GoogleTest, Catch2, doctest
- Kotlin: JUnit, Kotest, MockK

Why this matters: tests are only useful if they fit the repo's actual framework.

### 5. Safer patch-first workflow

Make the review-first workflow impossible to miss:

```bash
testgen plan --path=./src --output-format=json
testgen generate --plan=.testgen/plan.json --dry-run --emit-patch
testgen apply --patch=.testgen/patch.json
testgen validate --changed
```

Why this matters: open-source users trust tools that avoid surprise writes.

### 6. Agent-readable JSON schemas

Publish JSON schemas for command outputs:

- analyze/cost response
- generation response
- patch response
- validation response
- capabilities response
- error envelope

Suggested path:

```text
docs/schemas/
```

Why this matters: agents and wrappers can parse TestGen reliably.

### 7. Example repos and golden fixtures

Add small example projects for each supported language:

```text
examples/projects/javascript-jest/
examples/projects/python-pytest/
examples/projects/go-testing/
examples/projects/rust-cargo/
examples/projects/java-junit/
examples/projects/csharp-xunit/
examples/projects/php-phpunit/
examples/projects/ruby-rspec/
examples/projects/cpp-googletest/
examples/projects/kotlin-junit/
```

Each example should include:

- source file
- existing test style
- expected generated test patch
- validation command

Why this matters: examples convert better than claims.

### 8. Provider and model profiles

Add named profiles for common LLM setups:

```yaml
profiles:
  cheap:
    provider: gemini
    model: gemini-2.5-flash
  quality:
    provider: anthropic
    model: claude-sonnet-4-6
  local:
    provider: ollama
    model: qwen2.5-coder
```

Why this matters: users think in outcomes, not provider plumbing.

### 9. Contribution-friendly adapter SDK

Make new language support easy to contribute.

Needed docs:

- adapter interface guide
- parser strategy guide
- fixture requirements
- framework detection checklist
- prompt template expectations
- validation command expectations

Why this matters: open-source growth depends on low-friction contributions.

### 10. Trust and governance polish

Add or improve:

- Code of Conduct
- clear issue labels
- `good first issue` list
- security policy with supported versions
- release checklist
- changelog
- architecture decision records
- compatibility policy for JSON output

Why this matters: people adopt projects that look maintained.

## Best next PRs

### PR 1: `testgen doctor`

User-facing value: high.
Agent value: high.
Risk: low.

Deliverables:

- repo readiness checks
- JSON and text output
- suggested next command
- docs and tests

### PR 2: `testgen capabilities`

User-facing value: medium.
Agent value: very high.
Risk: low.

Deliverables:

- command manifest
- supported language/framework metadata
- output schema version
- limitation notes

### PR 3: Cursor/Cline/Continue/Roo install targets

User-facing value: very high.
Agent value: high.
Risk: medium.

Deliverables:

- install script flags
- generated config files
- integration docs

### PR 4: JSON schemas

User-facing value: medium.
Agent value: very high.
Risk: low.

Deliverables:

- schemas under `docs/schemas/`
- schema links in CLI docs
- tests that sample output validates against schemas

### PR 5: framework detection

User-facing value: very high.
Agent value: high.
Risk: medium.

Deliverables:

- package-file scanners
- framework metadata in `analyze`, `cost`, and `languages`
- per-language tests

## Positioning for users

Use this message in docs, posts, and release notes:

> TestGen turns coding agents into safer test-generation agents. It analyzes your repo, estimates cost, generates dry-run patches, follows your existing test style, and validates before you merge.

Short version:

> Agent-native test generation. Cost-aware, patch-first, repo-style aware.

## What not to build yet

Avoid these until the core workflow is trusted:

- hosted SaaS dashboard
- broad benchmark claims without reproducible fixtures
- auto-writing tests by default
- too many providers before the output contract is stable
- complex plugin systems before the adapter SDK is documented
