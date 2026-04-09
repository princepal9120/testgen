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

## Coverage Goals

- Maintain CI baseline coverage threshold.
- Increase coverage for `internal/generator`, `internal/llm`, and `internal/validation` first.
- Raise threshold incrementally with each milestone.

## Local Commands

```bash
go test -race ./...
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out
```
