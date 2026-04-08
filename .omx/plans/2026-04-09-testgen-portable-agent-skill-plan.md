# TestGen Portable Agent Skill Plan

Date: 2026-04-09
Branch: main
Mode: SELECTIVE_EXPANSION
Status: Drafted from architecture review

## Executive Recommendation

Do **not** turn TestGen into a pure prompt-only skill.

Keep TestGen as the execution engine, then make it portable in this order:

1. **Stabilize a versioned JSON contract** around generation requests and results.
2. **Refactor the orchestration into a reusable application service** that CLI, TUI, wrappers, and servers all call.
3. **Ship thin wrappers for Claude/Codex-style skill systems** that call the stable contract.
4. **Add an MCP server interface** so OpenCode and other MCP-native agents can invoke TestGen as a tool.

This preserves the working CLI, avoids a rewrite, and gives you the one thing every agent ecosystem can share: a deterministic machine-callable interface.

## Why this matters

Right now TestGen is a useful developer CLI. The user goal is bigger: make it a reusable test-generation capability that other agent systems can compose into their own workflows.

That only works if TestGen becomes:
- deterministic enough for automation
- machine-readable enough for tool calling
- stable enough that wrappers do not break every release
- thin enough that each agent ecosystem gets a small adapter instead of a forked implementation

## System Audit

### Current strengths

1. **There is already a good extensibility seam for languages.**
   - `internal/adapters/adapter.go:13-50` defines a real language contract.
   - `internal/adapters/registry.go:20-31` centralizes adapter registration.

2. **There is already a provider abstraction for LLMs.**
   - `internal/llm/provider.go:23-42` gives you a provider interface that can stay below a future service layer.

3. **The core generation flow exists and is testable.**
   - `internal/generator/engine.go:77-176` already encapsulates the generation pipeline.

4. **The codebase is healthy enough to evolve.**
   - `go test ./...` passes on the current branch.

### Current architectural mismatches

1. **The docs say the command layer has no business logic, but the command layer is doing orchestration.**
   - `docs/ARCHITECTURE.md:36-40` says `cmd/` has "No business logic".
   - `cmd/generate.go:119-263` scans files, selects providers, builds the engine, processes results, and controls error/reporting behavior.
   - So the written architecture is cleaner than the real architecture.

2. **Parallel generation exists on paper but is not actually used in the main CLI path.**
   - `internal/generator/worker.go:26-120` defines a worker pool.
   - `cmd/generate.go:280-323` still processes files in a serial loop.
   - This is a credibility gap for any agent workflow that expects scalable batch generation.

3. **Validation is effectively a stub.**
   - `internal/validation/validator.go:46-69` computes coverage from a helper that always returns `false`.
   - That means agent workflows cannot trust `validate` as a postcondition today.

4. **Analyze is heuristic, not grounded in the same parsing pipeline as generation.**
   - `cmd/analyze.go:134-137` estimates function count by line count.
   - For agent tooling, cost and planning data should come from real definitions, not a rough guess.

5. **The TUI duplicates orchestration instead of reusing one shared application service.**
   - `internal/ui/tui/running.go:145-210` reimplements scanning, engine setup, adapter lookup, and generation.
   - This creates drift risk: CLI and TUI can diverge in behavior and output.

6. **There is no stable machine contract yet.**
   - `cmd/generate.go:335-355` returns ad hoc JSON output, but it is not versioned, typed, or documented as an automation contract.
   - Agent ecosystems need a stable response envelope with diagnostics, artifacts, usage, and partial-failure semantics.

7. **Observability exists as a package, but is not wired into the main flow.**
   - `internal/metrics/metrics.go:35-105` exists.
   - Repository search shows it is unused in runtime code.

8. **Provider error handling is too shallow for agent-grade automation.**
   - `internal/llm/openai.go:152-166` and `internal/llm/anthropic.go:144-149` special-case 429, but do not implement retries, backoff, structured classification, or durable diagnostics.
   - An agent calling TestGen needs actionable failures, not generic API errors.

## What already exists

| Sub-problem | Existing code | Reuse decision |
|---|---|---|
| Multi-language generation core | `internal/generator/engine.go`, `internal/adapters/*` | Reuse directly |
| LLM backend abstraction | `internal/llm/provider.go`, provider impls | Reuse, but harden |
| CLI shell entrypoint | `cmd/root.go`, `cmd/generate.go`, `cmd/analyze.go`, `cmd/validate.go` | Keep, thin out |
| Interactive UX | `internal/ui/tui/*` | Keep, rewire to shared service |
| Machine-readable output seed | `cmd/generate.go:335-355` | Replace with versioned contract |
| Metrics package | `internal/metrics/metrics.go` | Wire in, do not rewrite |
| Worker pool | `internal/generator/worker.go` | Either integrate or delete |

## Dream state delta

```text
CURRENT STATE
  Developer-facing CLI with working adapters and providers,
  but orchestration is spread across commands and TUI, validation is weak,
  and machine integration is only ad hoc JSON.

THIS PLAN
  Introduce one reusable application service plus a versioned contract,
  then layer wrappers and MCP on top.

12-MONTH IDEAL
  TestGen is the test-generation engine behind multiple agent ecosystems.
  Humans can run the CLI. Agents can call a stable JSON contract or MCP tool.
  Language packs, provider policies, and post-generation validation are pluggable.
```

## Landscape check

The market is telling you something simple:

- Claude Code supports project and personal `SKILL.md` skills, and says skills follow an open Agent Skills standard that works across multiple AI tools. Source: Anthropic docs, "Extend Claude with skills".
- OpenCode exposes MCP and auto-discovers tools from configured MCP servers. Source: OpenCode README.
- GitHub Copilot/Coding agents are also moving toward reusable agent skills/tooling concepts. Source: GitHub docs on agent skills.

The implication: **portable capability beats tool-specific prompt hacks**.

## Viable approaches

### Approach A: Prompt-wrapper only
**Summary:** Create Claude/Codex/OpenCode skill files that shell out to the existing CLI and parse current text/JSON output.
**Effort:** S
**Risk:** High
**Pros:**
- Fastest to ship
- Minimal code changes
- Good for quick demos
**Cons:**
- Brittle output parsing
- Duplicates behavior per ecosystem
- Breaks as CLI output evolves
**Reuses:** Existing CLI only

### Approach B: Stable JSON contract first, wrappers second, MCP third
**Summary:** Keep the CLI, introduce a versioned request/response contract and a shared app service, then add thin wrappers and MCP.
**Effort:** M
**Risk:** Low-Med
**Pros:**
- Lowest regret path
- Preserves existing users
- Lets all agent ecosystems converge on one execution surface
- Makes testing, metrics, and validation much easier
**Cons:**
- Slightly slower than a wrapper-only demo
- Requires refactor of orchestration seams
**Reuses:** Engine, adapters, providers, CLI, TUI, metrics

### Approach C: MCP-first rewrite
**Summary:** Rebuild TestGen primarily as an MCP server and let the CLI become a thin client.
**Effort:** L
**Risk:** High
**Pros:**
- Elegant long-term tool surface
- Strong fit for agent ecosystems that already prefer MCP
**Cons:**
- Rebuilds too much too early
- Raises delivery risk before the contract is stable
- Makes local human CLI experience a secondary concern
**Reuses:** Some engine logic, but forces bigger restructuring

## Recommendation

Choose **Approach B**.

It is the minimum serious architecture. Approach A is a demo. Approach C is a rewrite. Approach B is the one that actually compounds.

## ADR

### Decision
Adopt a **contract-first portable engine** architecture: one shared application service, one versioned JSON contract, multiple invocation surfaces (CLI, TUI, wrappers, MCP).

### Drivers
1. Preserve the existing working CLI.
2. Avoid tool-specific skill forks.
3. Make agent invocation deterministic and testable.
4. Keep diffs incremental and reversible.

### Alternatives considered
- Wrapper-only skill layer
- MCP-first rewrite

### Why chosen
Because it gives you a portable foundation without throwing away the existing CLI or overcommitting to one ecosystem.

### Consequences
- Short-term refactor cost in command/TUI orchestration
- Stronger long-term reuse across all agent ecosystems
- Cleaner testing, metrics, and validation boundaries

### Follow-ups
- Decide whether MCP ships as `testgen mcp` inside the main binary or as a companion binary.
- Decide whether skills are bundled in this repo or published as a companion repo.

## Target architecture

```text
                        +----------------------+
                        |  Agent Skill Wrapper |
                        | Claude / Codex / etc |
                        +----------+-----------+
                                   |
                                   v
+---------+    +---------+   +------------+   +------------------+
| CLI     |    | TUI     |   | MCP Server |   | Future HTTP API  |
| cmd/*   |    | ui/tui  |   | transport  |   | optional         |
+----+----+    +----+----+   +------+-----+   +---------+--------+
     |              |               |                   |
     +--------------+---------------+-------------------+
                                    |
                                    v
                      +-------------------------------+
                      | internal/app / testgen service |
                      | request validation             |
                      | orchestration                  |
                      | result envelope                |
                      +-----+---------------+----------+
                            |               |
                            v               v
                   +----------------+   +----------------+
                   | generator/      |   | validation/    |
                   | adapters/ llm   |   | metrics        |
                   +----------------+   +----------------+
```

## Request/response contract shape

### Request v1

```json
{
  "contract_version": "v1",
  "mode": "generate",
  "targets": [{"path": "src/foo.py"}],
  "test_types": ["unit", "edge-cases"],
  "framework": "auto",
  "provider": "anthropic",
  "validate": true,
  "dry_run": false,
  "output": {
    "format": "structured-json"
  }
}
```

### Response v1

```json
{
  "contract_version": "v1",
  "status": "success|partial|failed",
  "artifacts": [
    {
      "source_file": "src/foo.py",
      "test_file": "tests/test_foo.py",
      "language": "python",
      "functions_tested": ["foo", "bar"],
      "test_code": "...",
      "validation": {
        "status": "passed|failed|skipped",
        "errors": []
      }
    }
  ],
  "diagnostics": [
    {
      "code": "provider_rate_limited",
      "severity": "error",
      "message": "...",
      "target": "src/foo.py"
    }
  ],
  "usage": {
    "provider": "anthropic",
    "tokens_input": 0,
    "tokens_output": 0,
    "estimated_cost_usd": 0
  }
}
```

## Acceptance criteria

1. **One shared app service**
   - CLI and TUI stop owning orchestration logic directly.
   - `cmd/generate.go` and `internal/ui/tui/running.go` call the same service.

2. **Versioned automation contract**
   - `generate`, `analyze`, and `validate` all support a documented versioned structured JSON output.
   - Contract changes require version bumps.

3. **Portable agent wrappers**
   - Repo includes first-party wrapper examples for Claude-style skills and Codex-style project instructions/tool invocation.
   - Wrappers do not parse human text output.

4. **MCP surface**
   - A tool call can request generation via stdio transport.
   - Result maps 1:1 to the same v1 contract.

5. **Validation is real**
   - `validate` no longer reports fake 0% coverage because of `checkTestFileExists` always returning false.
   - At minimum, existence/path-based checks are language-aware and explicit about what is or is not verified.

6. **Observability is wired**
   - Metrics collector is used in the main generation path.
   - Partial failures are included in diagnostics.

7. **Tests cover the new surfaces**
   - Contract tests for JSON output
   - Service tests for orchestration
   - Wrapper smoke tests
   - MCP transport tests

## Implementation steps

### Step 1. Create the application service boundary
**Files:** `cmd/generate.go`, `cmd/analyze.go`, `cmd/validate.go`, `internal/ui/tui/running.go`, new `internal/app/` package

- Introduce an application-layer service like `internal/app/service.go`.
- Move scan -> plan -> generate -> validate -> summarize orchestration out of commands/TUI.
- CLI and TUI become request builders plus presenters.

### Step 2. Define the versioned contract
**Files:** new `pkg/contracts/` or `internal/contracts/`, `cmd/generate.go`, `cmd/analyze.go`, `cmd/validate.go`, docs

- Add typed request/response structs.
- Add `contract_version` and stable diagnostic codes.
- Replace ad hoc JSON in `cmd/generate.go:335-355` with the typed contract.
- Keep text output for humans, but generate it from the same result object.

### Step 3. Fix truthfulness gaps before exposing automation
**Files:** `internal/validation/validator.go`, `cmd/analyze.go`, `internal/metrics/metrics.go`, `internal/llm/*`

- Make validation explicit and language-aware.
- Make analyze use real parsed definitions where possible.
- Wire metrics collection into generation.
- Add provider retry/backoff/error classification.

### Step 4. Unify concurrency behavior
**Files:** `cmd/generate.go`, `internal/generator/worker.go`, `internal/ui/tui/running.go`

- Either integrate `WorkerPool` into the main service or remove it.
- The advertised `--parallel` flag must actually change execution behavior.
- Ensure TUI and CLI share the same parallelism semantics.

### Step 5. Add wrapper-ready invocation surfaces
**Files:** `docs/CLI_REFERENCE.md`, new `skills/` or `.agents/`, new `examples/agent-integration/`

- Add a stable non-interactive invocation mode for agents.
- Publish wrapper examples:
  - Claude skill invoking the binary with contract JSON
  - Codex project instruction example invoking the binary/tool
  - OpenCode config example using either the CLI or MCP server

### Step 6. Add MCP server mode
**Files:** new `cmd/mcp.go` or `cmd/serve.go`, new `internal/mcp/`

- Expose `generate_tests`, `analyze_codebase`, and `validate_tests` tools.
- Make the handler call the same app service as CLI/TUI.
- Return the same contract payload.

### Step 7. Docs and migration packaging
**Files:** `README.md`, `docs/ARCHITECTURE.md`, `docs/CLI_REFERENCE.md`, new `docs/agent-integration.md`

- Update architecture docs so they match reality.
- Document supported invocation modes and contract guarantees.
- Add migration notes for existing CLI users.

## Risks and mitigations

| Risk | Why it matters | Mitigation |
|---|---|---|
| Wrapper drift | Each agent ecosystem behaves differently | Keep wrappers thin and route all logic through contract v1 |
| Fake confidence from validation | Agents may trust broken results | Fix `internal/validation/validator.go` before calling the tool “agent-ready” |
| Serial execution under batch load | Large repos become slow and expensive | Wire `WorkerPool` or remove the flag claim |
| Provider flakiness | Tool-call UX gets unpredictable | Add retries, classified diagnostics, timeout policies |
| Contract churn | Ecosystem wrappers break | Version contract, test golden payloads, document changes |

## Verification plan

1. `go test ./...`
2. Contract golden tests for v1 JSON output
3. CLI-to-service integration tests for `generate`, `analyze`, `validate`
4. TUI smoke tests around shared service calls
5. MCP transport tests over stdio
6. Wrapper smoke tests that exercise Claude-style and OpenCode-style integrations

## Recommended sequencing

### Phase 1, 1-2 days human / ~30-60 min with strong AI execution
- App service extraction
- Contract v1 structs
- CLI JSON contract stabilization

### Phase 2, 2-3 days human / ~60-90 min with strong AI execution
- Validation truthfulness fixes
- Metrics wiring
- Provider error hardening
- Real parallelism

### Phase 3, 1-2 days human / ~30-60 min with strong AI execution
- Claude/Codex wrapper examples
- MCP mode
- Docs refresh

## Not in scope for this pass

- Full web product or SaaS dashboard
- IDE plugin work
- Rewriting all adapters
- E2E browser UI work
- Public hosted API

## Next execution recommendation

If you want to build this now, the next move is **an engineering plan**, not more strategy.

Recommended next command: **`$plan-eng-review`** focused on the app-service seam, contract package placement, and MCP command shape.
