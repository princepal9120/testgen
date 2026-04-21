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
- Tool handlers return structured JSON as text in the MCP tool result payload.

## Transport

- stdio
- JSON-RPC messages framed with `Content-Length`

## Notes

- Tool handlers call the shared `internal/app` layer.
- `testgen_generate` defaults to safe dry-run behavior unless the caller explicitly requests file writes.
- Structured outputs are returned as JSON text inside MCP tool results.
