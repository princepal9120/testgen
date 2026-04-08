# Test Specification: portable agent skill for TestGen

Date: 2026-04-09
Related PRD: `.omx/plans/prd-testgen-agent-skill-2026-04-09.md`

## Test Strategy

The new architecture only counts as done if the same generation core is exercised by:
- direct Go unit tests
- CLI integration tests
- wrapper-level contract tests

## New test surfaces

### 1. Application layer contract tests

Add tests for:
- `GenerateRequest` validation
- `GenerateResponse` serialization
- dry-run patch generation
- write mode file materialization
- structured warnings and errors

Required cases:
- single file, supported language
- directory input, recursive
- unsupported language
- zero definitions found
- provider failure
- malformed model output

### 2. Generator pipeline tests

Add or expand tests around:
- parse -> definition extraction -> prompt -> output flow
- duplicate test types per function
- post-processing imports
- patch mode vs write mode

### 3. Validation tests

Replace placeholder behavior with real test coverage for:
- test file exists
- compile/parse succeeds
- runner invocation succeeds/fails
- coverage parsing for Go, Python, JS/TS, Rust, Java where supported

### 4. Wrapper contract tests

For each wrapper:
- input template maps correctly into the JSON request
- app-layer response maps back into wrapper output
- errors remain structured and machine-readable

## Acceptance test matrix

| Area | Happy path | Failure path | Edge case |
|---|---|---|---|
| App request | valid request returns result | invalid request returns structured error | empty path / both file and path missing |
| Generation | supported file returns artifact | provider failure returns explicit error | file with no functions |
| Patch mode | dry-run returns patch object | patch creation fails cleanly | existing target file already present |
| Validation | valid tests pass | invalid tests fail with report | no runner installed |
| CLI bridge | CLI JSON matches app result | non-zero exit on fatal error | quiet/json modes together |
| TUI bridge | TUI uses app layer | canceled run propagates cancellation | no API key configured |
| Wrapper bridge | wrapper invokes core correctly | wrapper surfaces structured error | wrapper-specific optional args omitted |

## Required failure-mode tests

1. Missing API key
2. Provider returns 429
3. Provider returns non-JSON or malformed code block
4. Adapter parse failure
5. Unsupported language
6. Validation runner missing from PATH
7. File write permission failure
8. Context timeout / cancellation

## Observability checks

Add tests or assertions for:
- request ID / run ID present in result metadata
- warning list preserved
- provider/model recorded in usage output
- per-file status is explicit, no silent skips

## Regression guardrails

1. Existing `go test ./...` must stay green.
2. Existing CLI help behavior must stay green.
3. Output-format JSON snapshots should be versioned.
4. Architecture docs must only claim behavior covered by tests.

## Exit criteria

The work is only complete when:
1. one application-layer package exists and is used by CLI/TUI/wrapper code,
2. validation is no longer placeholder behavior,
3. structured contract tests pass,
4. one wrapper example for a real agent ecosystem is verified end to end.
