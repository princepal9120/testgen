# MCP integration

**Scope:** This page covers direct MCP usage. For the shared integration model and safe defaults, start with the [integrations index](./README.md).

TestGen includes a repo-local MCP stdio server.

## Run

```bash
testgen mcp
```

You can print a ready-to-paste config snippet with:

```bash
./scripts/print-mcp-config.sh testgen
```

## Exposed tools

- `testgen_generate`
- `testgen_analyze`
- `testgen_validate`

## Tool behavior notes

- `testgen_generate` accepts `path` or `file`, `types`, `dry_run`, `validate`, `emit_patch`, `parallelism`, `batch_size`, `provider`, and `write_files`.
- `testgen_generate` remains in safe dry-run mode unless the caller explicitly sets `write_files: true`.
- When usage transparency is enabled in the shared generate path, MCP callers receive the same additive request/cache/batch/cost fields as CLI JSON mode.
- `testgen_analyze` can be used with cost estimation for offline, provider-aware budget previews before a caller decides to generate anything.
- Tool handlers return structured JSON as text in the MCP tool result payload.

## Transport

- stdio
- JSON-RPC messages framed with `Content-Length`

## Notes

- Tool handlers call the shared `internal/app` layer.
- `testgen_generate` defaults to safe dry-run behavior unless the caller explicitly requests file writes.
- Structured outputs are returned as JSON text inside MCP tool results.
- Usage/cost reporting is additive and should not break existing MCP clients that already consume the shared envelope.
