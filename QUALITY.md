# Quality Standards

This file defines the engineering bar for TestGen.

## Engineering Principles

- Single responsibility: keep command handlers thin and deterministic.
- Interface boundaries: prefer dependency inversion across adapters/providers.
- Explicit errors: return actionable errors with context.
- Determinism first: avoid hidden side effects and flaky behavior.
- Backward compatibility: document breaking CLI changes before release.

## Quality Gates

All pull requests should satisfy:

- Formatting: `gofmt -w -s .`
- Linting: `golangci-lint run`
- Static checks: `go vet ./...`
- Tests: `go test -race ./...`
- Security checks: `govulncheck ./...`
- Coverage threshold: minimum total coverage enforced in CI (currently 15%)

## Coverage Policy

- Current baseline is set to prevent regressions while test suite grows.
- Raise threshold gradually as critical packages gain tests.
- New logic in `internal/*` should include targeted unit tests.

## Release Quality Checklist

- [ ] CI green on Linux, macOS, and Windows
- [ ] Coverage threshold satisfied
- [ ] Security scan completed
- [ ] Changelog/release notes generated
- [ ] No known critical regressions in core commands (`generate`, `analyze`, `validate`, `tui`)
