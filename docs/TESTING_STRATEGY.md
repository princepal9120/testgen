# Testing Strategy

**Scope:** This document is the contributor-facing testing strategy for the TestGen codebase. It explains how the project verifies behavior, not how end users run TestGen commands.

This document defines how TestGen verifies behavior with confidence.

## Test Pyramid

- Unit tests: parser, adapter, prompt/output processing, and validation logic.
- Integration tests: CLI command behavior, flags, exit codes, and filesystem interactions.
- End-to-end tests: minimal smoke coverage for complete command flows.

## What to Test

- Happy path behavior for each command.
- Edge cases: empty files, unsupported extensions, malformed config, API errors.
- Regression cases from resolved bugs.
- Output invariants: deterministic formatting and path generation.

## Test Design Guidelines

- Keep tests deterministic; avoid external network dependencies when possible.
- Use temporary directories for file-oriented tests.
- Prefer table-driven tests for parser and validator variants.
- Keep fixtures minimal and focused.
- For cost-efficiency work, prefer fake providers and fixture-backed usage totals over live API calls.

## Cost-Efficiency Regression Coverage

Goal 5 introduces cache, batching/chunking, and provider-aware reporting requirements that should stay locked down with deterministic tests:

- **Unit tests:** cache fingerprinting, cached-token accounting, pricing math, batch flush behavior, and chunk splitting/parsing.
- **Service tests:** provider-aware analyze output, additive generate usage blocks, offline/API-key-free cost estimates, and internal consistency between totals and per-file estimates.
- **Integration tests:** JSON/text snapshots for `testgen analyze --cost-estimate` and `testgen generate --report-usage`, plus non-breaking machine-mode envelopes.
- **Fixture-backed savings checks:** repeated-run fixtures proving cache reuse savings and bulk-run fixtures proving batching reduces request overhead.
- **Metrics persistence:** `.testgen/metrics` snapshots should reflect the same accounting totals surfaced in CLI/TUI/MCP responses.
- **Contract stability:** additive usage fields must not remove or rename the existing top-level machine-readable keys.

## Coverage Goals

- Maintain CI baseline coverage threshold.
- Increase coverage for `internal/generator`, `internal/llm`, and `internal/validation` first.
- Raise threshold incrementally with each milestone.
- Treat `internal/app`, `internal/generator`, `internal/llm`, and `tests/` as the minimum regression surface for Goal 5 changes.

## Local Commands

```bash
make fmt
make lint
go test ./...
go test ./internal/app ./internal/generator ./internal/llm ./tests
make test-coverage
make coverage-check
```

If you need a smaller local loop while iterating on Goal 5, start with the focused package set above, then finish with the full suite before merging.

```bash
go test -race ./...
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out
```
