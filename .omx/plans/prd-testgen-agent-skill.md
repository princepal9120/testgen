# PRD: TestGen as a Portable Agent Skill

Date: 2026-04-09
Branch: main
Commit: 436a3e0
Planning mode: SELECTIVE EXPANSION
Status: Draft approved for implementation planning

## Requirements Summary

Turn TestGen from a human-first CLI into a portable test-generation capability that agentic coding tools can call reliably.

The target is not "make three separate products." The target is one stable execution core with thin wrappers for:
- Codex-style skill workflows
- Claude Code / subagent workflows
- OpenCode agent workflows
- later, an MCP surface for broader tool interoperability

The product job to be done:
- An agent receives a repo or file path and a testing request.
- The agent calls TestGen in a machine-safe way.
- TestGen returns structured results, generated file paths, failures, and usage data.
- The wrapper agent decides whether to write, validate, retry, or ask the user.

## CEO Review Summary

### Current strengths
- The repo already has a good domain split between scanning, language adapters, LLM providers, and generation engine. See `internal/scanner/scanner.go`, `internal/adapters/adapter.go`, `internal/llm/provider.go`, `internal/generator/engine.go`.
- The adapter registry is the right shape for future language growth. `internal/adapters/registry.go:17-65`.
- The provider abstraction already supports multiple vendors, which matters for portable agent wrappers. `internal/llm/provider.go:18-98`.
- The repo already passes `go test ./...`, so this is not a broken foundation.

### Current architectural gaps
1. **No stable machine contract**
   - `cmd/generate.go:326-355` emits ad-hoc JSON output, but there is no versioned request/response schema, no stdin contract, and no explicit error taxonomy.
2. **Human UI and execution logic are coupled to orchestration**
   - `cmd/generate.go:119-324` performs scan, engine setup, processing, output formatting, and UI reporting in one command handler.
   - `internal/ui/tui/running.go:145-236` duplicates execution flow instead of reusing a shared application service.
3. **The architecture docs overstate what the system does**
   - `docs/ARCHITECTURE.md:62-70` claims caching, rate limiting, and batching in `internal/llm`, but the engine only calls `Provider.Complete` per definition and never uses batch execution.
   - `cmd/generate.go:280-281` literally says parallel processing will be added later, while `internal/generator/worker.go:11-121` already has a worker pool that is not used.
4. **Validation and analysis are placeholders, not trustworthy contracts**
   - `internal/validation/validator.go:33-60` always reports test-file existence via a stub that returns false.
   - `cmd/analyze.go:134-186` estimates functions and cost heuristically, not from actual adapter parsing or provider token counting.
5. **Config and metrics are not wired into the main execution path**
   - `internal/config/config.go` defines a real config model, but the app mostly uses direct Viper access from command handlers.
   - `internal/metrics/metrics.go` exists, but the core flow does not persist run metrics.

### Dream state delta

CURRENT STATE
`CLI/TUI app with useful internals, but no portable tool contract and duplicated execution paths`

THIS PLAN
`Extract a shared application service, define a versioned JSON contract, add agent wrappers, then add MCP server mode`

12-MONTH IDEAL
`TestGen is the reusable test-generation backend for humans, CI, Codex, Claude Code, OpenCode, and any MCP-capable client`

## RALPLAN-DR Summary

### Principles
1. One execution core, many shells.
2. Machine contract before wrapper proliferation.
3. Thin adapters over stable JSON, not per-agent reimplementation.
4. Validation must be real before agents can trust automation.
5. Observability is part of the product, not cleanup.

### Decision Drivers
1. Portability across agent ecosystems.
2. Deterministic, machine-safe invocation.
3. Low-maintenance architecture that preserves the current CLI value.

### Viable Options

#### Option A: Wrapper-only around current CLI
- Pros: fastest path, smallest diff.
- Cons: bakes in duplicated orchestration, weak JSON contract, fragile parsing.
- Verdict: useful as a spike, not strong enough as the product architecture.

#### Option B: Shared application service + stable JSON contract + wrappers
- Pros: best near-term leverage, preserves current CLI, unlocks wrappers fast.
- Cons: requires refactor of command/TUI execution path before shipping wrappers.
- Verdict: **recommended**.

#### Option C: MCP-first rewrite
- Pros: strongest long-term interoperability story.
- Cons: too early, forces protocol work before the core contract is trustworthy.
- Verdict: phase 2, not phase 1.

## ADR

### Decision
Adopt **Option B** now: create a shared application service and versioned JSON contract, then ship agent wrappers on top. Add MCP after the JSON contract is proven.

### Drivers
- Current code already has strong reusable internals.
- The biggest missing piece is a trustworthy programmatic interface.
- Agent ecosystems differ at the wrapper layer more than at the execution layer.

### Alternatives considered
- Keep the current CLI and just add skills.
- Jump straight to an MCP server.

### Why chosen
This is the smallest move that fixes the real product problem. It keeps the CLI alive, avoids wrapper drift, and does not force an early protocol bet.

### Consequences
- Some refactor work is mandatory before user-visible wrapper work.
- Validation and result schemas become first-class API surface.
- TUI and CLI must consume the same execution core.

### Follow-ups
- Add MCP once the JSON contract and wrapper ergonomics are stable.
- Add provider retries, rate-limit handling, and richer validation before claiming "agent-ready".

## What already exists

| Sub-problem | Existing code | Reuse decision |
|---|---|---|
| File discovery | `internal/scanner/scanner.go` | Reuse with minimal API cleanup |
| Language-specific parsing and formatting | `internal/adapters/*.go` + `internal/adapters/registry.go` | Reuse directly |
| Multi-provider LLM abstraction | `internal/llm/provider.go`, provider implementations | Reuse, but add error taxonomy + retry policy |
| Core generation loop | `internal/generator/engine.go` | Reuse after extracting request/response service boundary |
| Parallelism primitive | `internal/generator/worker.go` | Reuse by wiring into service layer |
| Human CLI output | `cmd/generate.go`, `cmd/analyze.go`, `cmd/validate.go` | Keep as wrapper |
| TUI flow | `internal/ui/tui/*.go` | Keep as wrapper, but remove duplicated execution logic |

## NOT in scope
- Full SaaS dashboard.
- IDE plugin ecosystem.
- Automatic production-code refactoring.
- Multi-turn autonomous bug fixing beyond test generation.
- Broad MCP tool catalog beyond the initial testgen surface.

## Acceptance Criteria

1. A versioned request/response contract exists in Go types and docs.
   - Example: `APIVersion: "v1"`, `Mode`, `TargetPath`, `TestTypes`, `Provider`, `WriteMode`, `Validate`, `Output`.
2. `testgen` can run in a machine mode without scraping human text.
   - The command accepts structured input or flags and emits a stable JSON envelope.
3. CLI and TUI both call the same application service.
   - No second orchestration path like `internal/ui/tui/running.go:145-236`.
4. Validation is real.
   - Generated-output validation checks at least syntax/compile/discovery for supported languages, not the current placeholder in `internal/validation/validator.go:52-60`.
5. Wrappers exist for three agent ecosystems.
   - Codex skill wrapper.
   - Claude Code wrapper or prompt/command file.
   - OpenCode agent definition.
6. Wrapper contracts are documented with copy-paste install instructions.
7. The core flow emits structured failure reasons.
   - Unsupported language, missing API key, provider timeout, rate limit, malformed model output, validation failure, write failure.
8. Metrics are persisted for machine-mode runs.
9. The implementation passes `go test ./...` and includes new contract/service tests.

## Implementation Steps

### Step 1: Define the reusable application contract
**Files:**
- add `internal/app/service.go`
- add `internal/app/types.go`
- update `pkg/models/models.go` only if shared DTO reuse is actually cleaner

**Work:**
- Define `GenerateRequest`, `GenerateResponse`, `FileResult`, `UsageSummary`, `FailureCode`.
- Make `GenerateResponse` the single truth for CLI JSON, TUI handoff, and future MCP responses.
- Include explicit versioning.

**Why first:**
Without this, every wrapper becomes brittle.

### Step 2: Move execution orchestration out of command handlers
**Files:**
- refactor `cmd/generate.go`
- refactor `internal/ui/tui/running.go`
- reuse `internal/generator/engine.go`
- wire `internal/generator/worker.go`

**Work:**
- Extract scan -> adapter selection -> generation -> validation -> metrics into `internal/app`.
- Make CLI and TUI thin callers.
- Actually use configured parallelism through `WorkerPool`.

### Step 3: Make validation and analysis trustworthy
**Files:**
- refactor `internal/validation/validator.go`
- refactor `cmd/analyze.go`
- potentially extend adapters for language-specific validation hooks

**Work:**
- Replace `checkTestFileExists` stub with language-aware detection.
- Move analysis to use adapter parsing and provider token estimation where possible.
- Return confidence markers when estimates are heuristic.

### Step 4: Add machine mode
**Files:**
- refactor `cmd/generate.go`
- optionally add `cmd/serve.go` later, but not in phase 1
- update README/docs

**Work:**
- Support a stable JSON envelope on stdout.
- Add an input mode suitable for wrappers, either JSON via stdin or a dedicated `--request-file` / `--stdin-json` path.
- Add non-zero exit codes mapped to failure classes.

### Step 5: Ship wrapper packs
**Files:**
- add `.codex/skills/testgen/SKILL.md`
- add `.claude/commands/testgen.md` or repo-appropriate Claude wrapper docs
- add `.opencode/agents/testgen.md`
- add `docs/agent-integration.md`

**Work:**
- Each wrapper calls the same machine mode.
- Keep wrapper logic tiny: validate inputs, invoke binary, summarize response, handle failure codes.
- Document required env vars and install path.

### Step 6: Add MCP in phase 2
**Files:**
- add `cmd/testgen-mcp/` or equivalent
- add `internal/mcp/`

**Work:**
- Expose tools like `analyze_codebase`, `generate_tests`, and `validate_tests`.
- Reuse the exact same application service and response types.

## Risks and Mitigations

| Risk | Why it matters | Mitigation |
|---|---|---|
| Wrapper drift across ecosystems | You end up maintaining 3 products | Keep wrappers as shells over one JSON contract |
| JSON contract churn | Agents break silently | Version the schema and keep old fields stable for one release |
| Validation remains fake | Agents will trust broken outputs | Make validation part of phase 1 gate |
| TUI/CLI divergence | Two bug surfaces forever | Force both through shared service |
| Provider failures stay generic | Bad agent UX and poor retries | Add explicit failure codes and retry guidance |
| MCP too early | Big protocol surface before core is ready | Delay until the contract is proven by wrappers |

## Verification Steps

1. `go test ./...`
2. Contract tests for JSON request/response serialization.
3. Golden tests for CLI machine-mode output.
4. Integration tests where wrapper fixtures invoke the binary and assert response envelopes.
5. Failure-path tests for: missing API key, unsupported language, malformed provider output, validation failure.
6. Manual proof:
   - Codex wrapper calls local binary and gets JSON.
   - Claude wrapper does the same.
   - OpenCode wrapper does the same.

## Follow-up Staffing Guidance

Recommended lanes if you execute this next:
- `architect`, high: define request/response schema and migration boundaries.
- `executor`, high: extract shared service and refactor CLI/TUI.
- `test-engineer`, medium: add contract + wrapper integration tests.
- `writer`, medium: wrapper installation docs and examples.

## Changelog for this plan
- Chose shared-service + JSON-contract architecture over wrapper-only and MCP-first paths.
- Elevated validation from nice-to-have to phase-1 gate.
- Sequenced MCP after wrapper proof instead of before it.
