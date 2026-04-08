# Agent distribution guide

TestGen is ready for repo-local use in:
- Codex 
- Claude Code
- OpenCode
- MCP-compatible clients via `testgen mcp`

## Install wrappers into another repo

Copy mode:

```bash
./scripts/install-agent-integrations.sh /path/to/target-repo copy
```

Symlink mode:

```bash
./scripts/install-agent-integrations.sh /path/to/target-repo symlink
```

## Print MCP config snippet

```bash
./scripts/print-mcp-config.sh testgen
```

Example output:

```json
{
  "mcpServers": {
    "testgen": {
      "command": "testgen",
      "args": ["mcp"]
    }
  }
}
```

## What is publish-ready vs local-ready

### Local-ready now
- repo-local wrapper files
- shared JSON contract
- MCP stdio server

### Still needed for public registry publishing
- published binary/package artifact
- final registry metadata (`server.json`) tied to a real distribution target
- versioned release process for MCP packaging
