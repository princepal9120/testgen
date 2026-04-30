# Contributing to TestGen

**Scope:** This document is the contributor workflow guide for TestGen. Use [`README.md`](README.md) for product onboarding and [`docs/CLI_REFERENCE.md`](docs/CLI_REFERENCE.md) for end-user command details.

Thank you for contributing to TestGen. This project welcomes code changes, bug reports, documentation improvements, and design feedback.

## Ground Rules

- Be respectful and constructive in all project spaces.
- Keep pull requests focused and small enough to review quickly.
- Prefer explicit, testable behavior over implicit or magic behavior.
- Follow the engineering principles in `QUALITY.md` and `docs/TESTING_STRATEGY.md`.

## Development Setup

1. Install Go 1.25.9 or newer for local development. CI currently uses Go 1.25.9, while `go.mod` declares the module minimum.
2. Fork and clone the repository.
3. Create a branch with the `codex/` prefix for feature work.
4. Install tools:

```bash
go mod download
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
go install golang.org/x/vuln/cmd/govulncheck@latest
```

## Build, Test, and Lint

Run these before opening a pull request:

```bash
make fmt
make lint
go test -race ./...
make test-coverage
make coverage-check
```

You can also run the CI-equivalent local quality check:

```bash
make ci
```

## Coding Standards

- Follow Go idioms and keep functions small and focused.
- Prefer composition and interface-based boundaries.
- Propagate errors with useful context.
- Avoid global mutable state where practical.
- Keep command layer thin (`cmd/`), move business logic to `internal/`.

## Test Expectations

- Add or update tests for behavior changes.
- Add regression tests for every bug fix.
- Prefer unit tests for parser, adapter, and generation logic.
- Use integration tests for CLI behavior and workflows.

## Documentation Expectations

- If onboarding or positioning changes, update `README.md`.
- If commands, flags, output, or config behavior change, update `docs/CLI_REFERENCE.md`.
- If architecture boundaries change, update `docs/ARCHITECTURE.md`.
- If LLM-agent adoption, integration targets, capability manifests, or agent install flows change, update `docs/LLM_AGENT_ADOPTION.md`.
- If integration surfaces change, update the relevant file under `docs/integrations/`.
- Do not present `PRD-TestGen.md` or `TechSpec-TestGen.md` as current implementation truth unless they are fully synchronized.

## Commit and PR Guidance

- Use clear commit messages in imperative mood.
- Keep one logical change per PR.
- Update docs if command behavior, flags, output, or supported languages change.
- Include a short test plan in your PR description.

## PR Checklist

- [ ] Code is formatted (`gofmt`)
- [ ] Lint passes (`golangci-lint`, `go vet`)
- [ ] Tests pass locally (`go test -race ./...`)
- [ ] Coverage does not regress below project baseline
- [ ] Relevant docs were updated

## Reporting Security Issues

Please do not open public GitHub issues for security vulnerabilities.
Follow `SECURITY.md` for private reporting instructions.
