# MCP integration

**Scope:** This page covers direct MCP usage. For the shared integration model and safe defaults, start with the [integrations index](./README.md).

TestGen includes an experimental MCP stdio server.

## Run

```bash
testgen mcp
```

## Exposed tools

- `testgen_generate`
- `testgen_analyze`
- `testgen_validate`

## Transport

- stdio
- JSON-RPC messages framed with `Content-Length`

## Notes

- Tool handlers call the shared `internal/app` layer.
- `testgen_generate` defaults to safe dry-run behavior unless the caller explicitly requests file writes.
- Structured outputs are returned as JSON text inside MCP tool results.
