# MCP integration

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
