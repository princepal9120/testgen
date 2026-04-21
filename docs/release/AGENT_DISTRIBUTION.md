# Agent distribution guide

This guide covers the current installation, upgrade, and distribution story for TestGen's agent-facing surfaces.

TestGen is ready for repo-local use in:
- Codex
- Claude Code
- OpenCode
- MCP-compatible clients via `testgen mcp`

## Install or upgrade the TestGen binary

### Latest GitHub release installers

macOS / Linux:

```bash
curl -fsSL https://raw.githubusercontent.com/princepal9120/testgen/main/install.sh | bash
```

Windows PowerShell:

```powershell
irm https://raw.githubusercontent.com/princepal9120/testgen/main/install.ps1 | iex
```

Both installers fetch the latest GitHub release for the current platform.

### Go install alternative

```bash
go install github.com/princepal9120/testgen-cli@latest
```

### Build from source

```bash
git clone https://github.com/princepal9120/testgen.git
cd testgen
go build -o testgen .
```

### Upgrade guidance

- Re-run the installer to pick up the newest published binary.
- Or rerun `go install github.com/princepal9120/testgen-cli@latest`.
- If a target repo copied wrapper files instead of symlinking them, rerun `./scripts/install-agent-integrations.sh` after upgrading so those repo-local assets stay aligned with the current docs and wrapper behavior.

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

If your binary is not on `PATH`, pass the explicit binary path instead:

```bash
./scripts/print-mcp-config.sh /absolute/path/to/testgen
```

## What is publish-ready vs local-ready

### Local-ready now
- GitHub release installers for the CLI binary
- `go install` and source-build paths
- repo-local wrapper files
- shared JSON contract
- MCP stdio server

### Still needed for public registry publishing
- package-manager distribution targets such as Homebrew/Scoop/etc. if broader install channels are desired
- final registry metadata (`server.json`) tied to a real distribution target
- versioned release process for MCP packaging
