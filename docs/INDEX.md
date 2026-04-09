# TestGen Documentation Map

**Scope:** This page explains which document to read for which purpose and which files are the current sources of truth.

## Start here

If you are new to the project:
1. Read [`README.md`](../README.md) for the product overview and quick start.
2. Read [`docs/CLI_REFERENCE.md`](./CLI_REFERENCE.md) for commands and flags.
3. Read [`docs/integrations/README.md`](./integrations/README.md) if you are using TestGen from an agent or MCP client.

## Source of truth

| Topic | Canonical doc | Notes |
|------|---------------|-------|
| Product overview and onboarding | [`README.md`](../README.md) | High-level only |
| Commands and flags | [`docs/CLI_REFERENCE.md`](./CLI_REFERENCE.md) | Validate against `cmd/*.go` |
| Architecture | [`docs/ARCHITECTURE.md`](./ARCHITECTURE.md) | Implementation-focused |
| Agent and MCP usage | [`docs/integrations/README.md`](./integrations/README.md) | Shared contract and per-tool guides |
| Contributor workflow | [`CONTRIBUTING.md`](../CONTRIBUTING.md) | Build, test, lint, PR expectations |
| Testing approach | [`docs/TESTING_STRATEGY.md`](./TESTING_STRATEGY.md) | Contributor-facing verification strategy |
| Quality standards | [`QUALITY.md`](../QUALITY.md) | Engineering principles and release bar |

## Documentation catalog

### Current user-facing docs

| Document | Audience | Purpose |
|----------|----------|---------|
| [`README.md`](../README.md) | New users | What TestGen is and how to get started |
| [`docs/CLI_REFERENCE.md`](./CLI_REFERENCE.md) | CLI users | Detailed command and flag reference |
| [`docs/integrations/README.md`](./integrations/README.md) | Agent and tool users | Shared integration overview |
| [`docs/integrations/codex.md`](./integrations/codex.md) | Codex / oh-my-codex users | Codex-specific setup |
| [`docs/integrations/claude-code.md`](./integrations/claude-code.md) | Claude Code users | Claude-specific setup |
| [`docs/integrations/opencode.md`](./integrations/opencode.md) | OpenCode users | OpenCode-specific setup |
| [`docs/integrations/mcp.md`](./integrations/mcp.md) | MCP clients | MCP server usage |
| [`docs/ARCHITECTURE.md`](./ARCHITECTURE.md) | Maintainers | Package and runtime architecture overview |

### Contributor and maintenance docs

| Document | Audience | Purpose |
|----------|----------|---------|
| [`CONTRIBUTING.md`](../CONTRIBUTING.md) | Contributors | Setup, workflow, and PR expectations |
| [`docs/TESTING_STRATEGY.md`](./TESTING_STRATEGY.md) | Contributors | Testing approach for the codebase |
| [`QUALITY.md`](../QUALITY.md) | Maintainers and contributors | Quality principles |
| [`SECURITY.md`](../SECURITY.md) | Security reporters | Responsible disclosure |
| [`SUPPORT.md`](../SUPPORT.md) | Users | Support channels |
| [`ROADMAP.md`](../ROADMAP.md) | Users and maintainers | Future direction |

### Historical / design-context docs

| Document | Status | Why keep it |
|----------|--------|-------------|
| [`PRD-TestGen.md`](../PRD-TestGen.md) | Historical / aspirational | Product vision and requirements thinking |
| [`TechSpec-TestGen.md`](../TechSpec-TestGen.md) | Historical / aspirational | Broader design exploration and technical direction |

## Rules for updating docs

- If command behavior, flags, or output change, update [`docs/CLI_REFERENCE.md`](./CLI_REFERENCE.md).
- If the onboarding story changes, update [`README.md`](../README.md).
- If architecture boundaries change, update [`docs/ARCHITECTURE.md`](./ARCHITECTURE.md).
- If an integration surface changes, update the relevant file under [`docs/integrations/`](./integrations/README.md).
- Do not present historical design documents as the current implementation source of truth unless they are fully synchronized first.
