# Test Spec: TestGen Portable Agent Skill

Date: 2026-04-09
Related PRD: `.omx/plans/prd-testgen-agent-skill.md`
Branch: main
Commit: 436a3e0

## Scope Under Test

Phase 1 covers:
- shared application service
- stable JSON request/response envelope
- CLI machine mode
- TUI reuse of the shared service
- validation improvements
- Codex / Claude / OpenCode wrappers

Phase 2 covers:
- MCP server mode

## Test Matrix

### 1. Contract tests

#### 1.1 Request schema round-trip
- Serialize and deserialize `GenerateRequest`.
- Verify defaults are explicit, not implicit magic.
- Verify unknown version fails loudly.

#### 1.2 Response schema round-trip
- Serialize and deserialize `GenerateResponse`.
- Verify success and partial-failure envelopes.
- Verify `failure_code` is stable.

#### 1.3 Backward-compatibility guard
- Golden snapshots for `v1` JSON responses.
- New fields may be additive, but existing keys cannot silently disappear.

### 2. Service-layer tests

#### 2.1 Happy path
- Input: supported source file, configured provider mock, dry run.
- Assert: one `FileResult`, generated test code, usage summary, no writes.

#### 2.2 Nil / empty / invalid shadow paths
- Nil-equivalent request fields.
- Empty target path.
- Empty test types.
- Unsupported language.
- Missing provider selection.

#### 2.3 Error-path tests
- Source file unreadable.
- Adapter parse failure.
- Provider timeout.
- Provider rate limit.
- Malformed provider output.
- Formatter failure with graceful fallback.
- Validation failure after generation.
- Output write failure.

### 3. CLI tests

#### 3.1 Machine-mode success
- `testgen generate --output-format json ...`
- Assert valid `v1` envelope.

#### 3.2 Machine-mode failure
- Missing API key returns non-zero exit code and structured failure.

#### 3.3 Human mode unchanged
- Existing human-readable behavior still works for non-JSON mode.

### 4. TUI tests

#### 4.1 TUI reuses service
- Unit test the TUI running model against a fake service.
- No direct orchestration logic should remain in the screen model.

#### 4.2 Cancellation behavior
- Cancel in-flight run.
- Assert structured cancelled response or known UI state.

### 5. Validation tests

#### 5.1 Language-aware discovery
- Go `_test.go`
- Python `test_*.py` and `*_test.py`
- JS/TS `.test.` and `.spec.`
- Java and Rust conventions

#### 5.2 Syntax/compile validation
- Mock or fixture-based tests per supported language.
- Failure returns `validation_failed`, not generic unknown error.

### 6. Wrapper tests

#### 6.1 Codex wrapper
- Fixture invokes wrapper against a canned request.
- Asserts wrapper calls binary and interprets JSON correctly.

#### 6.2 Claude wrapper
- Same as above.

#### 6.3 OpenCode wrapper
- Same as above.

### 7. Observability tests

#### 7.1 Metrics persisted
- Machine-mode run writes metrics artifact.
- Usage totals and cache stats are present.

#### 7.2 Error logs are contextual
- Failures include file path, provider, mode, and failure code.

## Required New Test Files

- `internal/app/service_test.go`
- `internal/app/types_test.go`
- `cmd/generate_machine_test.go`
- `internal/validation/validator_test.go`
- `tests/agent_wrapper_integration_test.go`

## Required Golden Fixtures

- `testdata/contracts/generate_success_v1.json`
- `testdata/contracts/generate_partial_failure_v1.json`
- `testdata/contracts/generate_missing_api_key_v1.json`

## Release Gate

Do not call this agent-ready until all of these pass:
1. `go test ./...`
2. JSON golden tests
3. wrapper integration tests
4. validation failure-path tests
5. one manual invocation per wrapper target

## Known Gaps to Close During Implementation

- `internal/validation/validator.go` is currently a stub and cannot be trusted.
- `cmd/analyze.go` uses rough heuristics and should not be treated as machine-truth.
- `internal/generator/worker.go` exists but is not exercised by the main CLI path.
