<!-- /autoplan restore point: /Users/prince/.gstack/projects/princepal9120-testgen/main-autoplan-restore-20260421-221458.md -->
# PRD: TestGen as a Trusted Agent Backend

Date: 2026-04-09
Branch: main
Commit: 0910c73
Planning mode: SELECTIVE EXPANSION
Status: Autoplan reviewed, pending final approval gate

## Requirements Summary

Turn TestGen from a useful CLI with early agent surfaces into a **trusted test-generation backend** that teams can rely on from Codex, Claude Code, OpenCode, CI, and MCP clients.

This is no longer a “build wrappers later” plan. The repo already contains:
- a shared application layer in `internal/app/`
- repo-local wrapper assets for Codex / Claude Code / OpenCode
- an MCP stdio command at `testgen mcp`

So the real job is not inventing agent support. The real job is making the existing claim **trustworthy, versioned, measurable, and distributable**.

### Buyer and user
- **Buyer / chooser:** the team or maintainer deciding whether to add TestGen to a repo, CI lane, or agent workflow.
- **Primary user:** a developer or coding agent operating inside an existing repository and needing review-first test generation with low blast radius.
- **Why they switch:** not because TestGen supports more wrapper names, but because it produces safer, more reviewable, more machine-readable results than generic built-in agent generation.

### Product job to be done
- A developer or agent points TestGen at a repo or file.
- TestGen returns a stable contract with artifacts, patches, validation outcomes, and explicit failure codes.
- The caller can safely choose to inspect, apply, validate, retry, or abort.
- The team can measure whether generated tests were trusted, applied, and accepted.

## CEO Review Summary

### Current strengths
- The repo already has strong reusable seams in `internal/scanner/`, `internal/adapters/`, `internal/llm/`, and `internal/generator/`.
- `internal/app/service.go` already centralizes generate / analyze / validate orchestration for CLI, TUI, and wrappers.
- Repo-local integration docs and wrapper assets already exist for Codex, Claude Code, OpenCode, and MCP.
- `go test ./...` passes on the current branch, so this is a stable base to harden instead of a rescue mission.

### Current strategic gaps
1. **The plan was stale relative to the repo.** It described shared service, wrappers, and MCP as future work even though they already exist in code and docs.
2. **The buyer was missing.** The original framing centered an “agent” as the actor, but teams choose tools, budgets, and trust policies.
3. **The moat was misframed.** Wrapper count is not the wedge. Trust, safe review-first operation, and distributable install paths are.
4. **Distribution was underweighted.** `docs/release/AGENT_DISTRIBUTION.md` already admits the repo is only “local-ready now” and still lacks public publish-ready metadata/process.
5. **Outcome metrics were absent.** The prior plan talked about usage data, but not trust metrics like patch-apply rate, validation-pass-after-write rate, or rerun rate.

### Current architectural gaps
1. **The machine contract exists, but is not actually versioned or failure-coded.** `internal/app/types.go` has structured request/response types, but no `api_version`, `request_id`, `failure_code`, or compatibility guarantees.
2. **Analyze is still heuristic.** `internal/app/service.go` estimates functions as `lines/20` and costs from static constants instead of adapter-backed parsing or provider-backed estimates.
3. **Validation is still only partially trustworthy.** `internal/validation/validator.go` begins with file-existence checks and upgrades coverage opportunistically only when language runners happen to produce usable output.
4. **Metrics exist but are not wired into the execution path.** `internal/metrics/metrics.go` persists run files, but the shared application layer does not use it.
5. **Distribution and compatibility posture are not part of the implementation sequence.** The old plan waited too long to make installation and release credibility first-class.

### Premise challenge
Accepted premise reframe from the approval gate:
- Old premise: “make TestGen portable across agent ecosystems.”
- New premise: “make TestGen’s existing agent-facing surfaces trustworthy, measurable, and distributable.”

### Dream state delta

CURRENT STATE
`A working CLI plus early wrapper/MCP surfaces, but with weak versioning, partial validation trust, heuristic analysis, and incomplete distribution/release posture.`

THIS PLAN
`Harden the shared contract, validation/analyze trust, failure taxonomy, install/release path, and outcome metrics across the existing CLI, wrapper, and MCP surfaces.`

12-MONTH IDEAL
`TestGen is the trusted backend teams prefer over generic agent-native test generation because it is safer, more reviewable, easier to install, and measurable in real workflow outcomes.`

## RALPLAN-DR Summary

### Principles
1. One execution core, many shells.
2. Trust before wrapper proliferation.
3. Distribution before surface-area expansion.
4. Explicit contracts over implied JSON shape.
5. Measure trusted outcomes, not just transport success.

### Decision Drivers
1. Increase trust in generated tests and machine callers.
2. Make public docs honest about what is already implemented versus what is still missing.
3. Reduce future wrapper and protocol drift by versioning the current surfaces.

### Viable Options

#### Approach A: Keep the current story, add missing hardening around it
- Summary: Treat the existing shared service and wrappers as the product baseline, then add versioning, failure taxonomy, validation hardening, metrics, and release/distribution support.
- Effort: M
- Risk: Low-Medium
- Pros:
  - Works with the codebase you already have.
  - Fixes trust and packaging without another rewrite story.
  - Gives near-term user value quickly.
- Cons:
  - Requires revising stale docs/plan language.
  - Forces compatibility discipline immediately.
- Reuses:
  - `internal/app`, `internal/mcp`, wrapper docs/assets, `internal/generator`, `internal/validation`

#### Approach B: Re-open architecture extraction as if nothing shipped yet
- Summary: Pretend shared service / wrappers / MCP are still future work and rebuild the sequence from scratch.
- Effort: M
- Risk: High
- Pros:
  - Clean narrative on paper.
- Cons:
  - Fights the repo’s current state.
  - Risks duplicate work and stale planning.
  - Delays the real bottlenecks.
- Reuses:
  - Limited, because the plan is written against a past snapshot.

#### Approach C: Narrow to one primary channel first, defer the rest
- Summary: Pick one primary agent surface plus MCP, and de-emphasize named wrappers until trust and distribution are proven.
- Effort: S-M
- Risk: Medium
- Pros:
  - Sharpest focus.
  - Avoids overcommitting to ecosystem-specific glue.
- Cons:
  - Creates a product-positioning decision that changes current docs.
  - May underuse the wrapper assets already present.
- Reuses:
  - Shared service, MCP, one wrapper lane

**RECOMMENDATION:** Choose **Approach A** because it matches the actual repo, fixes the real bottlenecks, and preserves the complete product path without pretending shipped surfaces do not exist.

## ADR

### Decision
Adopt **Approach A**: keep the current shared-service / wrapper / MCP baseline, then harden it into a trusted and distributable contract.

### Drivers
- Shared service, wrappers, and MCP already exist.
- The biggest missing pieces are trust, compatibility discipline, and distribution credibility.
- Rewriting the story from scratch would spend effort on narration instead of product value.

### Alternatives considered
- Rewrite the plan as if shared service and wrappers were still future work.
- Narrow to a single primary channel and treat all wrapper assets as premature.

### Why chosen
It is the smallest move that addresses the real product gap. It fixes trust and distribution without discarding current surfaces or multiplying architecture churn.

### Consequences
- Existing public docs and acceptance criteria must be updated to match reality.
- Contract evolution becomes a first-class compatibility problem.
- Validation, analysis, and failure semantics now matter as product surface, not implementation detail.

### Follow-ups
- Add backward-compatibility tests for JSON and MCP responses.
- Decide whether future ecosystem expansion should stay repo-local or grow into publishable registry integrations.

## What already exists

| Sub-problem | Existing code | Reuse decision |
|---|---|---|
| File discovery | `internal/scanner/scanner.go` | Reuse directly |
| Language-specific parsing and formatting | `internal/adapters/*.go` + `internal/adapters/registry.go` | Reuse directly |
| Multi-provider LLM abstraction | `internal/llm/provider.go`, provider implementations | Reuse, but classify failures and retry posture explicitly |
| Shared orchestration | `internal/app/service.go`, `internal/app/types.go` | Reuse and harden instead of re-introducing |
| Parallelism primitive | `internal/generator/worker.go` | Reuse, already wired through `internal/app` |
| Human CLI output | `cmd/generate.go`, `cmd/analyze.go`, `cmd/validate.go` | Keep as wrapper |
| TUI flow | `internal/ui/tui/*.go` | Keep as wrapper over shared service |
| MCP server | `cmd/mcp.go`, `internal/mcp/server.go` | Keep and harden, not phase-2 greenfield |
| Repo-local wrappers | `.codex/skills/testgen/SKILL.md`, `.claude/commands/testgen.md`, `.opencode/commands/testgen.md` | Keep thin and align with the hardened contract |
| Distribution scaffolding | `docs/release/AGENT_DISTRIBUTION.md`, install scripts | Reuse and finish |

## NOT in scope
- Full SaaS dashboard.
- IDE plugin ecosystem.
- Automatic production-code refactoring.
- Multi-turn autonomous bug fixing beyond test generation.
- Hosted playground / public demo environment in the same PR.
- Broad MCP tool catalog beyond the current `generate`, `analyze`, and `validate` tools.

## Acceptance Criteria

1. **Versioned machine contract**
   - `GenerateResponse`, `AnalyzeResponse`, `ValidateResponse`, and MCP tool payloads expose an explicit `api_version`.
   - Add stable `failure_code` values for unsupported language, missing API key, provider timeout, malformed provider output, validation failure, write failure, and no-source-files.
2. **Stable machine modes**
   - `testgen generate` supports the current flag-based JSON mode and one explicit machine-input path (`--stdin-json` or `--request-file`).
   - Safe dry-run remains the default for agent callers unless `write_files` is explicitly requested.
3. **Trustworthy validation/analyze**
   - Validation moves beyond file existence into language-aware compilation/discovery semantics.
   - Analyze stops relying on `lines/20` as the main estimate and marks remaining heuristics explicitly.
4. **Compatibility discipline**
   - Add golden tests for CLI JSON output and MCP tool payloads.
   - Existing keys do not disappear silently across one release.
5. **Distribution credibility**
   - The plan defines how teams install TestGen beyond local source builds, including binary/package release shape and MCP publish-ready metadata path.
6. **Metrics and trust outcomes**
   - Machine-mode runs persist run metrics.
   - Track at least one trust-oriented metric: patch-apply rate, validation-pass-after-write rate, or rerun rate.
7. **Docs stay honest**
   - README, CLI reference, integrations docs, and release guide align with the actual implementation and release posture.
8. **Verification passes**
   - `go test ./...` passes.
   - Contract/service tests, MCP tests, and failure-path tests cover the hardened semantics.

## Implementation Steps

### Step 1: Audit and rename the current truth
**Files:**
- `.omx/plans/prd-testgen-agent-skill.md`
- `README.md`
- `docs/integrations/README.md`
- `docs/release/AGENT_DISTRIBUTION.md`

**Work:**
- Rewrite the product story around trust hardening, not greenfield wrapper creation.
- Call out that shared service, wrappers, and MCP already exist.
- Separate “local-ready now” from “publish-ready next.”

### Step 2: Harden the shared contract that already exists
**Files:**
- `internal/app/types.go`
- `internal/app/service.go`
- `cmd/generate.go`
- `cmd/analyze.go`
- `cmd/validate.go`

**Work:**
- Add explicit `api_version`, `failure_code`, request correlation, and write-mode semantics.
- Add a machine-input lane (`--stdin-json` or `--request-file`).
- Keep existing JSON output stable while extending it additively.

### Step 3: Make trust claims real
**Files:**
- `internal/validation/validator.go`
- `internal/validation/coverage.go`
- `internal/generator/engine.go`
- `internal/llm/reliable.go`
- adapter validation hooks as needed

**Work:**
- Replace placeholder validation logic with language-aware checks.
- Classify provider and formatting failures explicitly.
- Decide how malformed model output and partial failures surface in the contract.

### Step 4: Make analysis believable
**Files:**
- `internal/app/service.go`
- `cmd/analyze.go`
- adapters / provider helpers as needed

**Work:**
- Use parser-backed definition counts where possible.
- Mark heuristic estimates explicitly when exact counts are unavailable.
- Keep JSON and text outputs aligned.

### Step 5: Finish the install / release / distribution lane
**Files:**
- `install.sh`
- `install.ps1`
- `scripts/install-agent-integrations.sh`
- `scripts/print-mcp-config.sh`
- `docs/release/AGENT_DISTRIBUTION.md`

**Work:**
- Define publish-ready binary/package targets.
- Define release ownership, versioning, and registry metadata expectations.
- Keep repo-local wrappers, but stop pretending they are the whole distribution story.

### Step 6: Harden wrapper and MCP convergence
**Files:**
- `.codex/skills/testgen/SKILL.md`
- `.claude/commands/testgen.md`
- `.opencode/commands/testgen.md`
- `cmd/mcp.go`
- `internal/mcp/server.go`

**Work:**
- Keep wrappers thin and contract-dependent.
- Treat MCP as an existing surface that needs compatibility hardening, not phase-2 invention.
- Align all examples around the same safe defaults and failure semantics.

### Step 7: Wire metrics and outcome tracking
**Files:**
- `internal/metrics/metrics.go`
- `internal/app/service.go`
- docs / release notes as needed

**Work:**
- Persist run metrics for machine-mode calls.
- Add at least one trust-oriented metric and define where it is recorded.

## Risks and Mitigations

| Risk | Why it matters | Mitigation |
|---|---|---|
| Plan/doc drift repeats | Public docs keep promising futures that already shipped or never shipped | Make doc alignment part of the implementation checklist |
| Contract churn breaks wrappers | Thin wrappers still fail if payload shape changes | Add `api_version`, additive schema changes, and golden tests |
| Validation remains weak | Teams will not trust generated tests | Make validation hardening a phase-1 gate |
| Distribution stays local-only | Product remains demo-friendly but adoption-hostile | Finish publish-ready release metadata and install path |
| Wrapper count becomes vanity metric | Engineering optimizes surface area instead of trust | Measure trust outcomes, not just invocation counts |
| MCP surface diverges from CLI JSON | Two compatibility stories emerge | Reuse shared response types and golden-test both |

## Verification Steps

1. `go test ./...`
2. Contract tests for JSON request/response serialization with additive-field compatibility.
3. Golden tests for CLI JSON output.
4. Golden tests for MCP `tools/call` payload contents.
5. Failure-path tests for missing API key, unsupported language, malformed provider output, validation failure, and write failure.
6. Manual proof:
   - Codex wrapper calls local binary and gets the hardened JSON envelope.
   - Claude Code wrapper does the same.
   - OpenCode wrapper does the same.
   - MCP `testgen_generate` stays dry-run unless `write_files=true`.
7. Trust proof:
   - at least one run metric persists,
   - at least one trust metric is emitted or stored.

## Follow-up Staffing Guidance

Recommended lanes if you execute this next:
- `architect`, high: contract/versioning and compatibility boundaries.
- `executor`, high: validation/analyze hardening plus install/release wiring.
- `test-engineer`, medium: JSON/MCP golden tests and failure-path coverage.
- `writer`, medium: doc alignment and release/distribution instructions.

## AUTOPLAN REVIEW — Phase 1 CEO

### 0A. Premise Challenge
- The original plan solved for “agent portability,” but the real chooser is the team installing and trusting the tool.
- Doing nothing leaves the repo in an awkward state: it already claims agent/MCP readiness publicly, but lacks enough versioning, trust signals, and publish-ready distribution to make those claims durable.
- The stronger framing is trust hardening of existing surfaces, not new wrapper proliferation.

### 0B. Existing Code Leverage
- `internal/app/service.go` already centralizes orchestration.
- `cmd/generate.go`, `cmd/analyze.go`, `cmd/validate.go`, and `internal/ui/tui/running.go` already consume that shared layer.
- `cmd/mcp.go` + `internal/mcp/server.go` already expose MCP tools.
- `.codex`, `.claude`, and `.opencode` wrapper assets already exist.

### 0C-bis. Implementation Alternatives
- See **Viable Options** above. Minimal viable = Approach C. Ideal architecture = Approach A. Chosen = Approach A.

### 0D. Selective expansion decisions
Accepted into scope:
- Add distribution/release credibility, not just repo-local wrappers.
- Add trust-oriented metrics, not just usage counts.
- Treat MCP as an existing surface to harden, not future greenfield.

Deferred to `TODOS.md`:
- Zero-setup hosted playground.
- Public registry / package ecosystem growth beyond initial release-hardening.

Skipped:
- IDE ecosystem expansion.
- Hosted SaaS dashboard.

### 0E. Temporal interrogation
- **Hour 1:** implementer needs the real contract owner and compatibility rules.
- **Hour 2-3:** they will hit ambiguity around failure codes, partial failures, and write semantics.
- **Hour 4-5:** they will discover install/release gaps if distribution is not planned up front.
- **Hour 6+:** they will wish trust metrics and golden compatibility tests existed before the first wrapper regression.

### 0F. Mode selection
- Confirmed mode: **SELECTIVE EXPANSION**
- Chosen approach under this mode: **Approach A**

### CODEX SAYS (CEO — strategy challenge)
- The plan had no real buyer, only an execution actor.
- The PRD was strategically stale relative to shipped repo state.
- Distribution and trust matter more than wrapper count.
- The moat is trusted output, not portability theater.
- Outcome metrics were missing.

### PRIMARY REVIEW (CEO — strategy synthesis)
- Agrees with Codex on reframe toward trust + distribution.
- Agrees that “ship three wrappers” cannot remain a phase-1 success criterion.
- Adds that MCP should be hardened as an existing surface, not deferred as new build-out.

### CEO DUAL VOICES — CONSENSUS TABLE
═══════════════════════════════════════════════════════════════
  Dimension                           Primary  Codex  Consensus
  ──────────────────────────────────── ───────  ─────  ─────────
  1. Premises valid?                   Risk     Risk   CONFIRMED
  2. Right problem to solve?           Reframe  Reframe CONFIRMED
  3. Scope calibration correct?        Mixed    Mixed  CONFIRMED
  4. Alternatives sufficiently explored? Gap    Gap    CONFIRMED
  5. Competitive / market risks covered? Gap    Gap    CONFIRMED
  6. 6-month trajectory sound?         Only if reframe  Only if reframe CONFIRMED
═══════════════════════════════════════════════════════════════

### Error & Rescue Registry
| Method / Codepath | What can go wrong | Exception class / failure code | Rescued? | Rescue action | User sees |
|---|---|---|---|---|---|
| CLI / machine request parse | invalid JSON / bad flags | `invalid_request` | Partial | fail fast with structured message | exact field/flag error |
| scan target path | no files / unreadable path | `no_source_files`, `path_unreadable` | Yes | stop before provider call | actionable scan failure |
| provider completion | timeout / 429 / malformed payload | `provider_timeout`, `provider_rate_limited`, `provider_output_invalid` | Partial | classify + retry guidance + preserve partial results | explicit provider failure |
| validation | compile/discovery failure | `validation_failed` | Yes | return generated artifact + failed validation state | partial success with fix hint |
| write materialization | cannot create/write test path | `write_failed` | Yes | preserve dry-run artifact for manual application | no silent loss |
| MCP tools/call | write_files omitted | `safe_dry_run_enforced` | Yes | force dry-run | safe result, no write |

### Failure Modes Registry
| Codepath | Failure mode | Rescued? | Test? | User sees? | Logged? |
|---|---|---:|---:|---|---:|
| `GenerateResponse` evolution | wrapper breaks on field churn | No today | Partial | confusing parse break | Partial |
| `AnalyzeResponse` heuristics | misleading effort estimate | Partial | No | false confidence | No |
| validation runner | reports coverage but not trust | Partial | Partial | green-looking but weak validation | Partial |
| MCP tool call | text-wrapped JSON payload drifts from CLI expectations | Partial | Partial | client glue code break | Partial |
| metrics collector | run metrics never persisted | No | No | no visibility into trust/adoption | No |

## AUTOPLAN REVIEW — Phase 3 Engineering

### Step 0. Scope Challenge
The smallest change that achieves the real goal is **not** “add wrappers.” It is:
1. harden the shared contract,
2. harden validation/analyze,
3. wire metrics,
4. finish distribution.
Anything else before that is scope theater.

### Architecture ASCII diagram
```text
                        INSTALL / RELEASE LANE
     install.sh / install.ps1 / release docs / package metadata / MCP config
                                 |
                                 v
+--------------------+    +-------------------------+    +----------------------+
| CLI / TUI / Skills |--> | internal/app Service    |--> | generator.Engine     |
| cmd/*, ui/tui,     |    | request/response owner  |    | parse -> prompt ->   |
| wrapper assets     |    | version/failure policy  |    | artifact             |
+--------------------+    +-------------------------+    +----------+-----------+
                                                                      |
                                 +------------------------------------+----------------+
                                 |                                     |                |
                                 v                                     v                v
                           scanner / adapters                    validation          metrics
                           discovery + parse                     trust gate          persisted run +
                                                                                     trust outcomes
                                 |
                                 v
                            MCP server
                     (same service, same contract,
                      same safe dry-run semantics)
```

### Data flow (including shadow paths)
```text
REQUEST
  -> validate request / flags / stdin-json
      -> [invalid] fail with failure_code=invalid_request
  -> scan source files
      -> [none found] fail with failure_code=no_source_files
  -> adapter parse + definition extraction
      -> [unsupported / parse fail] classify per-file error
  -> provider completion
      -> [timeout / malformed / rate limit] classify + partial result
  -> artifact format / patch generation
      -> [format fail] fallback or classify provider_output_invalid
  -> optional write
      -> [write_files omitted] stay dry-run
      -> [write fail] preserve artifact + failure_code=write_failed
  -> optional validation
      -> [validation fail] partial success with validation_failed
  -> metrics persist
      -> [persist fail] do not hide primary result, log metrics failure
```

### State machine
```text
request_received
  -> scanned
  -> generated
    -> dry_run_complete
    -> wrote_files
       -> validated
          -> success
          -> partial_failure
    -> partial_failure
  -> failed_fast
```

### Deployment sequence
```text
1. add additive schema fields
2. ship golden tests for CLI + MCP
3. update wrappers/docs to new fields
4. publish binary / package
5. publish MCP metadata / registry docs
```

### Rollback flow
```text
bad schema / wrapper break
  -> keep previous binary available
  -> revert additive field consumer changes
  -> retain prior api_version compatibility window
```

### Code quality review
- `cmd/generate.go` is already thin enough that “move orchestration out of command handlers” should be rewritten as “finish contract ownership and remove remaining presentation-only leakage.”
- `internal/app/types.go` is under-specified for long-term compatibility. It has structure, but not versioning, failure taxonomy, or trust/result metadata.
- `internal/mcp/server.go` returns JSON text inside tool content. That is workable, but the plan must treat this as a compatibility surface with tests, not an implementation detail.
- `internal/metrics/metrics.go` is a dead capability until wired into service-level execution.

### Test review diagram
```text
NEW UX FLOWS / CLI FLOWS
  [+] install binary -> set provider key -> analyze -> generate dry-run json
  [+] wrapper invokes binary in repo-local workflow
  [+] MCP client initialize -> tools/list -> tools/call

NEW DATA FLOWS
  [+] request -> app types -> JSON envelope
  [+] app response -> wrapper consumer
  [+] app response -> MCP text payload wrapper
  [+] metrics -> persisted run file

NEW CODEPATHS / BRANCHES TO COVER
  [+] additive schema version fields
  [+] stdin-json or request-file input mode
  [+] failure_code classification for all major failures
  [+] partial failure after validation / write
  [+] safe dry-run enforcement when write_files not set

NEW INTEGRATIONS / EXTERNAL CALLS
  [+] provider completion
  [+] language-specific validation runners
  [+] install / release scripts

NEW ERROR / RESCUE PATHS
  [+] no_source_files
  [+] provider_timeout
  [+] provider_output_invalid
  [+] validation_failed
  [+] write_failed
```

### Test coverage mapping
| Item | Test type | Exists? | Gap |
|---|---|---|---|
| CLI JSON additive schema | golden / integration | Partial | add backward-compat snapshots |
| MCP `tools/call` payload shape | integration / golden | Partial | add schema compatibility assertions |
| safe dry-run enforcement | MCP integration | Partial | add negative write-path cases |
| validation partial failure semantics | unit + integration | No | add explicit partial-success tests |
| analyze heuristic confidence marking | unit | No | add deterministic coverage |
| metrics persistence | integration | No | add run artifact assertions |
| install / release path docs/scripts | smoke | No | add doc/script smoke checks |

### Test plan artifact
- Written to: `/Users/prince/.gstack/projects/princepal9120-testgen/prince-main-eng-review-test-plan-20260421-222359.md`

### Performance review
- Main risk is not CPU latency, it is repeated provider cost and retry behavior without visible trust metrics.
- `internal/generator/worker.go` is already wired through `internal/app`, so the old plan’s “wire worker pool” item should be deleted.
- Compatibility churn is the real scaling risk: more wrappers and clients multiply breakage cost unless schema drift is controlled.

### Worktree parallelization strategy
| Step | Modules touched | Depends on |
|---|---|---|
| Contract hardening | `internal/app`, `cmd/`, `internal/mcp` | — |
| Trust hardening | `internal/validation`, `internal/generator`, adapters | Contract hardening |
| Distribution / release | `install*`, `scripts/`, `docs/release/` | Contract hardening |
| Docs alignment | `README.md`, `docs/`, wrapper assets | Contract hardening |
| Metrics wiring | `internal/metrics`, `internal/app` | Contract hardening |

Parallel lanes:
- **Lane A:** contract hardening -> metrics wiring (shared `internal/app`, sequential)
- **Lane B:** trust hardening (`internal/validation`, adapters) after Lane A interface decisions
- **Lane C:** distribution / release docs and scripts (mostly independent after schema decisions)
- **Lane D:** docs + wrapper alignment (parallel with Lane C once contract fields are finalized)

Execution order:
- Launch Lane A first.
- Then launch B + C + D in parallel.
- Merge A before the rest to minimize `internal/app` conflicts.

Conflict flags:
- Lanes A and D both touch wrapper-facing contract language.
- Lanes A and B both depend on final failure taxonomy.

### Engineering completion summary
- Architecture Review: 5 issues found
- Code Quality Review: 4 issues found
- Test Review: diagram produced, 7 major gaps identified
- Performance Review: 3 risks flagged
- Failure modes: 5 total, 2 critical gaps (`metrics`, `validation/trust semantics`)
- Parallelization: 4 lanes, 3 parallel after contract decisions

## AUTOPLAN REVIEW — Phase 3.5 DX

### Product type classification
- Primary type: **CLI Tool**
- Secondary types: **Documentation**, **Claude Code Skill / agent integration surface**, **MCP tool**

### Developer Persona Card
```text
TARGET DEVELOPER PERSONA
========================
Who:       Maintainer or staff-level developer integrating TestGen into an existing repo,
           CI lane, or coding-agent workflow.
Context:   They already have code and want safer, review-first test generation without
           teaching each agent a bespoke flow.
Tolerance: 5 minutes to first useful result, maybe 1 doc hop, very low tolerance for
           vague failures or unstable JSON.
Expects:   Copy-paste install, one provider key, stable JSON/MCP output, and no hidden writes.
```

### Developer Empathy Narrative
I land on the README and immediately see that TestGen claims to work for humans, CI pipelines, and coding agents. Good start. The quick start gives me an install script, then asks for a provider API key, then suggests `analyze`, then a dry-run `generate` call with JSON patch output. That is a credible flow for an already-motivated CLI user, but it is not yet a magical first run. I still need the binary installed, a provider key configured, and enough trust in the contract to wire it into my repo or agent.

Then I click into integrations docs. They tell me the shared contract exists and give me repo-local wrapper commands. That helps. But if I am evaluating this against stronger developer tools, I still do not know the compatibility policy, whether MCP/CLI payloads are versioned, or how safe upgrades are. If something breaks, I can probably recover, but I do not yet see a migration or release-confidence story. I would try it, but I would not yet standardize it across multiple repos without more trust signals.

### Competitive DX Benchmark
| Tool | TTHW | Notable DX choice | Source |
|---|---:|---|---|
| Stripe | ~2-5 min | strong quickstarts and CLI-first setup path | https://docs.stripe.com/development/quickstart |
| Vercel | ~2-5+ min | guided getting-started flow and clear account / CLI path | https://vercel.com/docs/getting-started-with-vercel |
| Firebase | ~5-10+ min | explicit project setup, SDK install, and onboarding steps | https://firebase.google.com/docs/web/setup |
| TestGen (current) | ~5-8 min | install script + API key + analyze + dry-run generate | README + integrations docs |

### Magical Moment Specification
- **Chosen vehicle:** copy-paste demo command
- **Moment:** the developer runs one review-first command and receives a stable JSON envelope with generated artifacts and patch operations, without writing files unexpectedly.
- **Why this vehicle:** it matches the product’s CLI/tooling identity and is much cheaper than building a hosted playground right now.

### Developer Journey Map
| Stage | Developer does | Friction points | Status |
|---|---|---|---|
| Discover | reads README and sees agent/MCP positioning | claims are ahead of compatibility policy | partial |
| Install | uses install script or builds from source | release / package posture unclear | gap |
| Hello World | sets provider key and runs dry-run JSON command | good path, but not fully trust-calibrated | partial |
| Real Usage | installs repo-local wrapper or MCP config | no explicit compatibility or migration story | gap |
| Debug | hits API key / provider / validation issues | structured failure docs and failure codes missing | gap |
| Upgrade | updates binary or wrapper docs | deprecation / migration path unspecified | gap |

### First-Time Developer Confusion Report
```text
FIRST-TIME DEVELOPER REPORT
============================
Persona: repo maintainer integrating a trusted CLI/tool backend
Attempting: TestGen getting started and agent integration

CONFUSION LOG:
T+0:00  I read the README and the product sounds ready for agents and MCP.
T+0:45  I install the binary and set a provider API key. Fine.
T+1:30  I can run the dry-run JSON example, but I do not know the compatibility promise of the payload.
T+2:30  I open the integrations docs and see wrapper assets plus MCP. Good, but I still do not know what breaks on upgrade.
T+3:30  I can probably try this in one repo, but I am not yet ready to standardize it broadly because failure semantics, release discipline, and migration posture are underspecified.
```

### DX Scorecard
+====================================================================+
|              DX PLAN REVIEW — SCORECARD                            |
+====================================================================+
| Dimension            | Score | Notes                               |
|----------------------|-------|-------------------------------------|
| Getting Started      | 7/10  | good copy-paste path, no magic tier |
| API/CLI/SDK          | 7/10  | names are sensible, contract thin   |
| Error Messages       | 4/10  | lacks explicit failure taxonomy     |
| Documentation        | 7/10  | docs are findable, trust story thin |
| Upgrade Path         | 3/10  | no migration / deprecation posture  |
| Dev Environment      | 7/10  | repo-local wrappers are practical   |
| Community            | 5/10  | adequate OSS basics, little more    |
| DX Measurement       | 3/10  | usage mentioned, trust metrics absent |
+--------------------------------------------------------------------+
| TTHW                 | ~5-8 min | target: <5 min                    |
| Competitive Rank     | Competitive for motivated CLI users, not champion |
| Magical Moment       | present in concept, not yet fully productized     |
| Product Type         | CLI Tool / agent integration surface               |
| Mode                 | DX POLISH                                          |
| Overall DX           | 5.4/10                                            |
+====================================================================+

### DX Implementation Checklist
- [ ] Time to hello world under 5 minutes from install to trusted dry-run output
- [ ] One explicit machine-input path in addition to flags
- [ ] Every failure path returns problem + cause + fix + failure code
- [ ] CLI JSON and MCP payloads have versioning and compatibility tests
- [ ] Upgrade / migration policy documented
- [ ] Install and release path documented beyond repo-local wrapper copying
- [ ] Trust metrics persisted for machine-mode runs

## Cross-phase themes
- **Theme: The plan was stale against the repo.** Flagged in CEO and Eng.
- **Theme: Trust beats wrapper count.** Flagged in CEO, Eng, and DX.
- **Theme: Distribution is load-bearing.** Flagged in CEO and DX, reinforced by release docs.
- **Theme: Compatibility/versioning is the real architecture work.** Flagged in Eng and DX.

## Decision Audit Trail
| # | Phase | Decision | Classification | Principle | Rationale | Rejected |
|---|---|---|---|---|---|---|
| 1 | CEO | Reframe from portability to trusted backend | Mechanical | completeness | The repo already has wrappers/MCP; trust is the missing value | keep old portability-first framing |
| 2 | CEO | Keep shared service as baseline, do not re-introduce it as future work | Mechanical | explicit over clever | `internal/app` already exists and is in use | pretend extraction is still phase 1 |
| 3 | CEO | Pull distribution/release into core scope | Taste | boil lakes | Local-ready without publish-ready distribution leaves adoption blocked | leave distribution as postscript |
| 4 | CEO | Treat wrapper count as secondary to trust metrics | Mechanical | pragmatic | wrapper proliferation without trust is local-demo theater | three-wrapper success criterion |
| 5 | Eng | Remove “wire worker pool” from main plan narrative | Mechanical | explicit over clever | `internal/app` already uses `WorkerPool` | keep stale step |
| 6 | Eng | Add versioning / failure taxonomy to shared contract | Mechanical | completeness | current types are structured but not durable | keep unversioned JSON |
| 7 | Eng | Require golden tests for CLI + MCP payloads | Mechanical | explicit over clever | compatibility is the real integration risk | rely on ad-hoc integration tests |
| 8 | DX | Use copy-paste demo command as magical moment | Taste | pragmatic | fits CLI identity with much lower effort than a playground | build hosted playground now |
| 9 | DX | Defer playground and public registry growth to TODOs | Mechanical | dry | trust/distribution basics are more urgent | widen scope now |

## GSTACK REVIEW REPORT
| Review | Runs | Status | Findings | Verdict |
|---|---:|---|---:|---|
| CEO | 1 | completed | 6 | REFRAME REQUIRED |
| Design | 0 | skipped | 0 | NO UI SCOPE |
| Eng | 1 | completed | 7 | MAJOR HARDENING NEEDED |
| DX | 1 | completed | 8 | NEEDS POLISH |
| Voices | 1 | codex-only for CEO, primary-led elsewhere | 4 consensus themes | ACTIONABLE |
