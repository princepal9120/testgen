# PRD: TestGen as a portable agent skill

Date: 2026-04-09
Repo: `princepal9120/testgen`
Mode: SELECTIVE_EXPANSION
Status: Drafted from architecture review

## Requirements Summary

Turn TestGen from a human-first CLI/TUI into a portable test-generation capability that agentic coding tools can call safely and predictably.

Primary targets:
- Codex-style instruction-file / agent workflows
- Claude Code custom slash-command workflows
- OpenCode-style tool wrappers

Core outcome:
- An agent can request test generation with a stable machine contract, receive structured results, and choose whether to write files, return patches, or only return a plan.

## System Audit

### Current strengths

1. The repo already has a clean-enough separation between scanning, adapters, LLM providers, and orchestration:
   - scanner: `internal/scanner/scanner.go`
   - adapters: `internal/adapters/adapter.go`, `internal/adapters/registry.go`
   - engine: `internal/generator/engine.go`
   - providers: `internal/llm/provider.go`
2. The adapter registry gives TestGen a real extension seam for languages, not a pile of `switch` statements, `internal/adapters/registry.go:20-31`.
3. The project passes `go test ./...` on 2026-04-09.

### Current architecture gaps

1. The core engine is still file-system-first, not agent-first. `Engine.Generate` reads files from disk, writes output files, and optionally validates them in one flow, `internal/generator/engine.go:77-175`.
2. There is no stable structured request/response contract for external callers. The CLI is the contract today.
3. Provider handling is thin. It has almost no retry/backoff, malformed-output recovery, or refusal handling, `internal/llm/provider.go:23-42`, `internal/llm/openai.go:97-193`, `internal/llm/anthropic.go:90-182`.
4. The TUI bypasses the worker pool and reimplements orchestration in-screen, `internal/ui/tui/running.go:145-210`.
5. `validate` is mostly a placeholder. `checkTestFileExists` always returns false, so the command is not trustworthy as an agent-facing verification primitive, `internal/validation/validator.go:54-58`.
6. `analyze` is heuristic, not semantic. It estimates function count by line count, `cmd/analyze.go:134-186`.
7. The architecture docs oversell the implementation. Docs mention a cleaner layered system and stronger parsing than the current code actually provides, `docs/ARCHITECTURE.md:5-130`, `TechSpec-TestGen.md:23-131`, versus regex-heavy parsing in `internal/adapters/*.go`.

## What already exists

| Sub-problem | Existing code | Reuse verdict |
|---|---|---|
| Source discovery | `internal/scanner/scanner.go` | Reuse |
| Language routing | `internal/adapters/registry.go` | Reuse |
| Per-language parsing/prompt logic | `internal/adapters/*.go` | Reuse, but harden |
| LLM abstraction | `internal/llm/provider.go` and provider impls | Reuse, but add reliability layer |
| Multi-file orchestration | `internal/generator/engine.go`, `worker.go` | Reuse after extraction |
| Human UX | `cmd/*`, `internal/ui/tui/*` | Keep as outer shells, not the platform core |

## Dream state delta

```text
CURRENT STATE
Human-first CLI/TUI. Works from disk. Loose contracts. Best-effort validation.

        ->

THIS PLAN
Extract an agent-safe application layer with a stable JSON contract.
Keep CLI/TUI as wrappers. Add tool-specific skill shims for Codex/Claude/OpenCode.

        ->

12-MONTH IDEAL
TestGen is both:
1) a great standalone CLI for humans
2) a transport-agnostic test-generation service callable by agents via JSON/MCP/wrappers

One core engine. Many fronts. No duplicated business logic.
```

## Implementation alternatives

### Approach A: Wrap the existing CLI as-is

- Effort: S
- Risk: High
- Pros:
  - Fastest path to “something works”
  - Minimal code movement
- Cons:
  - CLI text output becomes the API
  - Hard to support patch mode, dry-run diff mode, and deterministic agent retries
  - Keeps current validation and orchestration weaknesses
- Reuses:
  - Nearly all existing code unchanged

### Approach B: Extract an application layer plus JSON contract, then add wrappers

- Effort: M
- Risk: Medium
- Pros:
  - Best balance of speed and long-term sanity
  - Lets CLI, TUI, and agent wrappers all share the same core
  - Makes MCP or daemon support possible later without rewriting everything
- Cons:
  - Requires one real refactor before adding wrappers
  - Forces contract design work up front
- Reuses:
  - scanner, adapters, provider abstraction, worker pool, most engine logic

### Approach C: Build MCP server first and treat CLI as a client

- Effort: L
- Risk: Medium
- Pros:
  - Strongest long-term portability story
  - Natural fit for agent ecosystems adopting MCP-like tool invocation
- Cons:
  - More platform work before solving current product gaps
  - Risks building infra before fixing trustworthiness of generation and validation
- Reuses:
  - Same core packages, but needs deeper transport refactor

## Recommendation

Choose **Approach B**.

Reason: the current repo already has useful seams, but the business logic is still glued to CLI/file-system execution. Extracting one application-layer contract gets you portability without prematurely turning the whole product into infra.

## ADR

### Decision

Adopt a **core-engine plus thin-wrapper** architecture.

### Drivers

1. Agents need a stable machine contract, not human-formatted CLI text.
2. The current engine is reusable, but only after file I/O and write-side effects are separated from generation logic.
3. A portable skill should support multiple fronts without cloning logic for each agent ecosystem.

### Alternatives considered

- Keep CLI as the API surface
- Jump straight to MCP-first architecture

### Why chosen

This is the minimum architecture that is still correct six months from now.

### Consequences

- Short-term refactor before shipping wrappers
- Clearer boundaries
- Easier testing
- Future MCP support without redoing business logic

### Follow-ups

- Consider MCP transport after the JSON application layer lands
- Add patch/diff output mode after structured generation is stable

## Product scope

### In scope

1. A reusable application-layer API for generation requests and results.
2. Structured non-interactive execution modes:
   - `plan`
   - `generate`
   - `validate`
   - `patch`
3. Agent wrapper packages/docs for:
   - Codex
   - Claude Code
   - OpenCode-like tool runners
4. Safer validation and error reporting.
5. A stable contract for file writes, dry runs, and patch output.

### Not in scope

1. Full IDE plugin platform
2. Cloud dashboard
3. Multi-repo orchestration service
4. Full MCP server in the first milestone

## Acceptance Criteria

1. There is a single application entry point callable from Go tests without shelling out.
2. The application layer accepts a typed request and returns a typed result with:
   - discovered files
   - generated artifacts
   - validation results
   - warnings/errors
   - usage metadata
3. CLI `generate`, TUI generate flow, and at least one agent wrapper all use the same application layer.
4. Dry-run mode can return structured patches without writing files.
5. Validation no longer reports false negatives for every file.
6. Error cases are explicit:
   - missing API key
   - provider rate limit
   - malformed model output
   - unsupported language
   - parse failure
   - validation failure
7. Regression tests cover request/response behavior for the application layer.

## Target architecture

```text
                    +----------------------+
                    |  Codex wrapper       |
                    +----------+-----------+
                               |
                    +----------v-----------+
                    |  Claude wrapper      |
                    +----------+-----------+
                               |
                    +----------v-----------+
                    |  OpenCode wrapper    |
                    +----------+-----------+
                               |
        +----------------------+----------------------+
        |   cmd/* and TUI also call the same layer    |
        +----------------------+----------------------+
                               |
                    +----------v-----------+
                    | application/service  |
                    | request -> result    |
                    +----------+-----------+
                               |
               +---------------+----------------+
               |                                |
      +--------v--------+              +--------v--------+
      | scan/adapt      |              | provider facade |
      +--------+--------+              +--------+--------+
               |                                |
               +---------------+----------------+
                               |
                    +----------v-----------+
                    | generation pipeline  |
                    +----------+-----------+
                               |
                    +----------v-----------+
                    | patch/write/validate |
                    +----------------------+
```

## Implementation Steps

### Step 1. Extract an application layer

Create a package like `internal/app` or `internal/service` with typed request/response models:
- `GenerateRequest`
- `GenerateResponse`
- `Artifact`
- `Patch`
- `ValidationReport`

Move orchestration out of `cmd/generate.go` and `internal/ui/tui/running.go`.

### Step 2. Split generation from side effects

Refactor `internal/generator/engine.go:77-175` so generation can return artifacts without writing files.

Introduce explicit stages:
1. discover
2. parse
3. generate
4. post-process
5. validate
6. materialize as patch or file write

### Step 3. Add a provider reliability facade

Wrap `internal/llm.Provider` with:
- retries for transient failures
- rate-limit backoff
- malformed-output normalization
- refusal/empty-response handling
- usage recording

### Step 4. Make validation trustworthy

Replace placeholder validation in `internal/validation/validator.go:54-58`.

Validation must distinguish:
- test file existence
- compile/parse success
- runner execution success
- coverage extraction

### Step 5. Add a stable JSON contract

Expose structured CLI modes:
- `testgen generate --output-format json`
- `testgen generate --dry-run --emit-patch`
- `testgen validate --output-format json`

This is the compatibility layer for wrapper skills.

### Step 6. Add wrapper surfaces

Ship example integrations:
- `.codex/skills/testgen/`
- `.claude/commands/testgen.md`
- `docs/integrations/opencode.md`

Each wrapper should do input shaping only. No business logic duplication.

### Step 7. Harden docs and contract truthfulness

Update:
- `README.md`
- `docs/ARCHITECTURE.md`
- `docs/CLI_REFERENCE.md`

Docs must match code. No more “Tree-sitter” claims unless it is actually in the codebase.

## Risks and Mitigations

| Risk | Why it matters | Mitigation |
|---|---|---|
| CLI refactor breaks existing UX | Current users are humans first | Keep CLI flags stable, refactor behind command handlers |
| Regex parsing causes silent bad generations | Agent trust dies fast | Make parse failures explicit and add parser regression tests |
| Validation still weak | Agents need proof, not hope | Treat validation as first-class output, not optional decoration |
| Wrapper sprawl | Three ecosystems can become three products | Keep wrappers dumb and the core typed |
| MCP temptation too early | Infra rabbit hole | Land JSON app layer first, then evaluate MCP |

## Verification Steps

1. `go test ./...`
2. Golden tests for application-layer JSON request/response.
3. Snapshot tests for patch output.
4. Integration tests for:
   - CLI -> app layer
   - TUI -> app layer
   - one wrapper -> app layer
5. Failure-path tests for:
   - missing API key
   - 429 / rate limiting
   - empty model response
   - malformed code block
   - unsupported file type

## Recommended staffing if executed later

- `executor` high: application-layer extraction
- `executor` medium: wrapper shims and docs
- `test-engineer` medium: contract and regression coverage
- `verifier` high: final evidence pass

## Evidence used

- `internal/generator/engine.go:45-215`
- `internal/generator/worker.go:11-121`
- `internal/adapters/adapter.go:13-50`
- `internal/adapters/registry.go:20-87`
- `internal/llm/provider.go:23-42`
- `internal/llm/openai.go:97-193`
- `internal/llm/anthropic.go:90-182`
- `internal/validation/validator.go:30-58`
- `cmd/analyze.go:120-187`
- `internal/ui/tui/running.go:145-236`
- `docs/ARCHITECTURE.md:5-130`
