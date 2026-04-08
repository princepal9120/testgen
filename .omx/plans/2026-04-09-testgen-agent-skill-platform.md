# Plan: Turn TestGen into a portable agent skill platform

Generated: 2026-04-09
Repo: princepal9120/testgen
Branch: main
Mode: SELECTIVE_EXPANSION
Status: Drafted from architecture review

## Requirements Summary

Build TestGen so agentic coding tools like Codex, Claude Code, and OpenCode-style systems can invoke it as a reusable skill to generate tests on demand.

The platform should:
- keep the current standalone CLI value,
- expose a stable machine-readable contract for agent callers,
- support thin wrappers for multiple agent ecosystems,
- preserve the current language adapter and provider seams,
- make output more deterministic, observable, and verifiable than the current prompt-to-file flow.

## Current Architecture Review

### What is already strong
- `internal/adapters/adapter.go:13-50` gives you a real language seam already. That is the right foundation for multi-agent portability.
- `internal/llm/provider.go:23-42` gives you a provider seam. Good. The shape is reusable even if implementation quality needs work.
- `internal/generator/engine.go:77-175` already centralizes parse -> extract -> prompt -> LLM -> format -> write. That is the business core.
- `internal/adapters/registry.go:20-31` makes language onboarding straightforward.
- The repo is healthy enough to iterate on: `go test ./...` passed on 2026-04-09.

### Architectural mismatches and bottlenecks
- There is no stable machine API. The only first-class entrypoints are human-facing CLI/TUI flows in `cmd/generate.go:119-264` and `internal/ui/tui/running.go:145-236`.
- The orchestration is duplicated. CLI and TUI each build their own scan/generate flow instead of calling one application service. That will get worse once you add Codex/Claude/Open wrappers.
- Parallelism is mostly aspirational. `cmd/generate.go:280` says parallel processing will be added later, while `internal/generator/worker.go` exists but is not the primary path.
- Validation is a stub. `internal/validation/validator.go:46-69` always reports missing tests because `checkTestFileExists` always returns false.
- Analysis is heuristic, not source-aware. `cmd/analyze.go:135-186` estimates functions from `lines/20` and costs from static assumptions, which is too weak for agents that need trustworthy planning data.
- Prompt/response handling is fragile. `internal/generator/engine.go:185-215` uses raw string templates plus regex extraction, while `parseStructuredOutput` exists at `internal/generator/engine.go:306-320` but is unused.
- The engine writes files directly in `internal/generator/engine.go:159-165`, which makes it harder to support preview, patch proposals, agent approval flows, and remote/MCP use.
- The documentation overstates the architecture. `docs/ARCHITECTURE.md:5-130` says “Clean Architecture”, but the app still routes too much orchestration through the presentation layer.

## RALPLAN-DR Summary

### Principles
1. One test-generation core, many caller surfaces.
2. Thin agent wrappers, thick reusable application service.
3. Deterministic structured outputs before fancy UX.
4. Validation and observability are product scope, not cleanup.
5. Preserve the existing CLI while making agent integration first-class.

### Decision Drivers
1. Cross-agent portability.
2. Minimal rewrite with strong reuse of existing adapters/providers.
3. Trustworthy machine-readable outputs for autonomous callers.

### Viable Options

#### Approach A: Thin wrapper over current CLI
- Summary: Keep current architecture, add JSON output conventions and create skill wrappers that shell out to `testgen generate`.
- Effort: S
- Risk: High
- Pros:
  - Fastest to ship.
  - Lowest code churn.
  - Keeps current UX intact.
- Cons:
  - Bakes in duplicated orchestration.
  - Hard to support MCP cleanly.
  - Weak trust because validation/analyze remain approximate.
- Reuses:
  - `cmd/generate.go`, `cmd/analyze.go`, `cmd/validate.go`

#### Approach B: Extract app service + stable JSON contract + agent wrappers + MCP adapter
- Summary: Move generation/analyze/validate orchestration into a shared application layer, keep CLI/TUI as clients, and add portable agent wrappers plus an MCP server.
- Effort: M
- Risk: Medium
- Pros:
  - Best balance of speed and long-term shape.
  - Lets Codex, Claude Code, and OpenCode-style systems share the same contract.
  - Makes validation and observability fixable in one place.
- Cons:
  - Requires moving orchestration out of command handlers.
  - Needs a request/result schema and compatibility discipline.
  - MCP surface adds one more interface to maintain.
- Reuses:
  - `internal/generator`, `internal/adapters`, `internal/llm`, `internal/scanner`

#### Approach C: Long-running job service / daemon first
- Summary: Build TestGen as a persistent local service with sessions, jobs, and artifact history, then bolt CLI and skills on top.
- Effort: L
- Risk: Medium-High
- Pros:
  - Strongest long-term platform story.
  - Better for high-volume automation and retries.
  - Natural place for cost telemetry and resumability.
- Cons:
  - Too much platform before product proof.
  - Slower to ship.
  - Overbuilds for current repo maturity.
- Reuses:
  - Existing engine and adapters, but requires more new infrastructure.

## Recommendation

Choose **Approach B**.

It keeps the lake boilable. You already have the right domain seams. What you do not have is a reusable invocation contract. Extract that once, then let every agent ecosystem call the same thing.

## ADR

### Decision
Adopt a shared application-service architecture with a stable request/result contract, then layer CLI, TUI, skill wrappers, and an MCP server on top.

### Drivers
- Agent skills are converging on reusable skill/subagent packaging and machine-driven delegation.
- MCP gives you one integration point that multiple AI clients can use.
- Current CLI/TUI duplication will become a tax the moment you add more agent entrypoints.

### Alternatives considered
- Keep current CLI and add wrappers only.
- Build a daemon/job system first.

### Why chosen
It is the smallest change that actually fixes the architecture instead of papering over it.

### Consequences
- You will introduce a new application layer and request/result schema.
- CLI/TUI will become thinner.
- Agent integrations become mostly packaging work instead of logic forks.

### Follow-ups
- Formalize schema versioning.
- Add deterministic validation and artifact reporting.
- Add MCP tools only after the shared service exists.

## Accepted Scope
- Extract a shared `internal/app` or `internal/service` orchestration layer.
- Add a stable JSON request/result contract for generate/analyze/validate.
- Add portable wrappers for Codex/Claude/Open-style skill systems.
- Add an MCP server surface after the shared service exists.
- Replace heuristic analysis and stub validation on the main agent path.

## Not in Scope
- Hosted SaaS dashboard.
- IDE plugins.
- Multi-user remote job queue.
- PR automation / automatic commits.
- Full web UI rewrite.

## Dream State Delta

```text
CURRENT STATE                         THIS PLAN                             12-MONTH IDEAL
CLI/TUI-first tool with reusable      Shared application service +          Portable local test-generation platform
adapters but no portable machine  ->  stable JSON contract + wrappers  ->  callable by any agent, via CLI, MCP,
contract, weak validation, and         + MCP surface, with real               or native skills, with reliable
approximate analysis                   validation/analysis                    artifacts, telemetry, and extensibility
```

## What Already Exists
- Language abstraction: `internal/adapters/adapter.go:13-50`
- Provider abstraction: `internal/llm/provider.go:23-42`
- Core generation flow: `internal/generator/engine.go:77-175`
- Registry-based language dispatch: `internal/adapters/registry.go:20-31`
- TUI screens and app shell: `internal/ui/tui/*.go`
- Scanner and ignore behavior: `internal/scanner/scanner.go:16-123`

## Target Architecture

```text
                  +-------------------+
                  |  Agent Wrappers   |
                  | Codex / Claude /  |
                  | Open-style skill  |
                  +---------+---------+
                            |
                  +---------v---------+
                  |   MCP Server      |
                  | generate/analyze/ |
                  | validate tools    |
                  +---------+---------+
                            |
        +-------------------v-------------------+
        | Shared App Service / Request Router   |
        | Generate(), Analyze(), Validate()     |
        | JSON schema v1                         |
        +--------+---------------+--------------+
                 |               | 
        +--------v-----+  +------v-------+
        | Generator    |  | Validation   |
        | Engine       |  | + Analysis   |
        +--------+-----+  +------+-------+
                 |               |
        +--------v-----+  +------v-------+
        | Adapters     |  | Metrics /    |
        | per language |  | Artifacts    |
        +--------+-----+  +--------------+
                 |
        +--------v-----+
        | LLM Provider |
        +--------------+
```

## Machine Contract v1

### Generate request
```json
{
  "version": "v1",
  "action": "generate",
  "path": "./src",
  "file": null,
  "recursive": true,
  "test_types": ["unit", "edge-cases"],
  "framework": null,
  "provider": "anthropic",
  "dry_run": true,
  "validate": true,
  "artifact_mode": "inline"
}
```

### Generate result
```json
{
  "version": "v1",
  "status": "ok",
  "artifacts": [
    {
      "source_path": "src/foo.py",
      "test_path": "tests/test_foo.py",
      "functions_tested": ["bar"],
      "test_code": "...",
      "validation": {
        "status": "passed",
        "runner": "pytest",
        "output": "..."
      }
    }
  ],
  "usage": {
    "provider": "anthropic",
    "model": "claude-3-5-sonnet-20241022",
    "tokens_input": 0,
    "tokens_output": 0,
    "estimated_cost_usd": 0
  },
  "errors": []
}
```

## Acceptance Criteria
- A single shared service powers CLI, TUI, wrappers, and MCP without duplicating orchestration logic.
- `generate`, `analyze`, and `validate` all support JSON request/result contracts with schema versioning.
- Agent wrappers for Codex, Claude Code, and one Open-style skill repo work without re-implementing generation logic.
- MCP exposes at least `generate_tests`, `analyze_testability`, and `validate_generated_tests`.
- Validation no longer returns universal false negatives.
- Analysis uses parser/adapter-aware counts, not `lines/20` heuristics.
- Dry-run returns artifacts without writing files.
- File-writing mode produces explicit artifact metadata and validation summaries.
- Docs explain where skills live and how each agent ecosystem invokes TestGen.
- `go test ./...` stays green.

## Implementation Steps

### Phase 1: Extract the reusable application service
1. Add `internal/app` or `internal/service` with request/result types for `Generate`, `Analyze`, and `Validate`.
2. Move orchestration out of `cmd/generate.go:119-264` and `internal/ui/tui/running.go:145-236` into shared service methods.
3. Keep `cmd/*` and `internal/ui/tui/*` as adapters over the service instead of owning business flow.

### Phase 2: Make outputs machine-safe
4. Stop having `internal/generator/engine.go:159-165` own the final side effect directly. Return artifact objects first, then let the caller choose write/preview/patch behavior.
5. Introduce structured result envelopes with explicit per-file success/failure, validation status, and usage metadata.
6. Promote structured response parsing instead of regex-only extraction. Either remove `parseStructuredOutput` or make it the default path.

### Phase 3: Fix the trust gaps
7. Replace `cmd/analyze.go:135-186` heuristics with adapter-backed parsing and real definition counts.
8. Replace `internal/validation/validator.go:46-69` stub logic with language-aware test path detection plus actual runner validation where available.
9. Wire metrics and cache stats into user-visible results rather than keeping them buried in provider internals.

### Phase 4: Add agent surfaces
10. Ship a portable skill package layout:
    - `.codex/skills/testgen/SKILL.md`
    - `.claude/agents/testgen-generator.md`
    - `.agents/skills/testgen/` or equivalent wrapper docs
11. Keep wrappers thin. They should only translate agent-specific invocation format into the shared JSON contract.
12. Add an MCP server package exposing `generate_tests`, `analyze_testability`, and `validate_generated_tests` backed by the same shared service.

### Phase 5: Hardening and docs
13. Update `docs/ARCHITECTURE.md` to reflect the actual architecture instead of the aspirational one.
14. Add integration tests that call the shared service directly, then separate smoke tests for CLI and MCP surfaces.
15. Document agent install flows and example prompts for Codex, Claude Code, and Open-style skill systems.

## Verification Steps
- `go test ./...`
- Add tests for shared service request/result translation.
- Add contract tests that snapshot JSON envelopes for generate/analyze/validate.
- Add wrapper smoke tests:
  - Codex skill invokes shared service
  - Claude subagent wrapper invokes shared service
  - MCP tool invocation returns valid schema
- Run one dry-run generation against `examples/` and verify artifact payloads.
- Run one write+validate flow per supported language sample where practical.

## Risks and Mitigations
- Risk: Wrapper sprawl.
  - Mitigation: make wrappers declarative and keep all business logic in shared service.
- Risk: MCP shipped too early.
  - Mitigation: only add MCP after shared service and JSON contract stabilize.
- Risk: Existing UX regresses during extraction.
  - Mitigation: preserve CLI flags and TUI screens while swapping internals underneath.
- Risk: Validation becomes flaky across languages.
  - Mitigation: start with file-path detection + opt-in runner validation, then expand language by language.

## Pre-mortem
1. You ship wrappers first, each wrapper forks logic, and now every agent integration behaves differently.
2. You ship MCP first, but the underlying analyze/validate paths are still weak, so autonomous callers stop trusting results.
3. You keep engine-side file writes as the only path, so dry-run, patch, review, and approval workflows stay clumsy.

## Expanded Test Plan
- Unit: request/result schemas, adapter dispatch, validation path resolution, structured response parsing.
- Integration: shared service over `examples/`, provider mocks, artifact generation, validation results.
- E2E: CLI JSON mode, TUI invoke path, MCP tool call.
- Observability: usage metrics included in result envelope, error code mapping, traceable artifact IDs.

## Recommended Agent Staffing Guidance
- `architect` or `planner`: define shared service boundary and schema versioning.
- `executor`: perform extraction from CLI/TUI into shared service.
- `test-engineer`: add contract and integration tests.
- `writer`: produce skill install docs and updated architecture docs.
- `verifier`: prove CLI, wrapper, and MCP flows all use the same contract.

## Changelog
- Initial plan drafted from repo audit and CEO-style architecture review on 2026-04-09.
