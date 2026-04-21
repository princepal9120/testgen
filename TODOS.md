# TODOS

## P1

### Publish installable distribution metadata beyond repo-local wrappers
- **What:** Add public release metadata for package/distribution channels after the hardened contract ships.
- **Why:** Repo-local wrappers work for contributors, but broader adoption still stalls if teams cannot discover and install TestGen through a standard package or registry path.
- **Pros:** Lowers adoption friction, makes docs honest, and turns local-ready demos into repeatable installs.
- **Cons:** Adds release-maintenance burden and compatibility commitments.
- **Context:** `docs/release/AGENT_DISTRIBUTION.md` already distinguishes local-ready from publish-ready. Finish the publish-ready side only after contract/versioning is stable.
- **Depends on / blocked by:** Versioned JSON/MCP contract, release process, binary packaging.

### Measure trust outcomes, not just transport success
- **What:** Track patch-apply rate, write-success rate, validation-pass-after-write rate, and rerun rate for machine-mode generation.
- **Why:** The real product wedge is trusted generated tests. Counting wrapper invocations alone will optimize the wrong thing.
- **Pros:** Gives a real adoption loop and tells you whether users trust the output.
- **Cons:** Requires schema changes and careful privacy boundaries.
- **Context:** Current plan already mentions persisted metrics, but not outcome metrics that prove developer trust.
- **Depends on / blocked by:** Metrics wiring in `internal/app`, stable response schema, privacy policy for telemetry.

## P2

### Add a zero-setup try-it experience
- **What:** Create a minimal playground, demo repo, or one-command sample flow that shows the “review-first test generation” moment without asking users to assemble everything manually.
- **Why:** Current Time to Hello World is competitive only for already-motivated CLI users. A lighter first-run path would widen the funnel.
- **Pros:** Better DX, easier demos, clearer magical moment.
- **Cons:** More product surface area and sample maintenance.
- **Context:** Defer until trust and distribution are solid. The current plan should not take this on in the same change set.
- **Depends on / blocked by:** Hardened contract, distribution path, example repos.
