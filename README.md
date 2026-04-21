# TestGen

<p align="center">
  <img src="website/images/logo.png" alt="TestGen Logo" width="120" />
</p>

**AI-powered test generation for humans, CI pipelines, and coding agents.**

TestGen is a multi-language CLI for inspecting code, generating tests, validating coverage, and fitting cleanly into local workflows, CI, and agent tooling. The CLI, TUI, and MCP server all ride on the same shared application layer, so teams can use one review-first backend across human and machine callers.

Supported languages: **JavaScript/TypeScript, Python, Go, Rust, and Java**.

## Why teams use TestGen

- **Start safely** with `testgen analyze` and dry-run generation before writing files
- **Work in the terminal** with either direct CLI commands or the interactive TUI
- **Integrate with agents** through shared JSON output, optional patch artifacts, and MCP
- **Keep workflows scriptable** for CI, automation, and repeatable review-first usage

## Quick start

### 1. Install the latest release or build from source

**macOS / Linux**

```bash
curl -fsSL https://raw.githubusercontent.com/princepal9120/testgen/main/install.sh | bash
```

**Windows (PowerShell)**

```powershell
irm https://raw.githubusercontent.com/princepal9120/testgen/main/install.ps1 | iex
```

Or build from source:

```bash
git clone https://github.com/princepal9120/testgen.git
cd testgen
go build -o testgen .
```

Upgrade story:

- Re-run the platform installer to fetch the latest GitHub release.
- Or use Go directly: `go install github.com/princepal9120/testgen-cli@latest`
- If you copied repo-local agent wrapper files into another repo, re-run `./scripts/install-agent-integrations.sh` after upgrading so those wrapper assets stay aligned.

### 2. Set one provider API key

```bash
export ANTHROPIC_API_KEY="..."
# or OPENAI_API_KEY / GEMINI_API_KEY / GROQ_API_KEY
```

### 3. Inspect the codebase first

```bash
testgen analyze --path=./src --cost-estimate
```

### 4. Generate review-first output

```bash
testgen generate --file=./src/utils.py \
  --type=unit \
  --dry-run \
  --emit-patch \
  --output-format json
```

This is the recommended safe default for agents and automation because it keeps file writes reviewable.

Explicit machine-input lane:

```bash
cat request.json | testgen generate --request-file=-
# or: testgen generate --request-file=./request.json
```

In machine mode, TestGen writes the shared JSON envelope to stdout and suppresses human-oriented Cobra banners on stderr.

### 5. Write and validate when ready

```bash
testgen generate --path=./src --recursive --type=unit --validate
```

For MCP and repo-local agent wrappers, see the integration docs for the same review-first flow and explicit write controls.

## Where next

- **Full command and flag reference** → [`docs/CLI_REFERENCE.md`](docs/CLI_REFERENCE.md)
- **Agent and MCP integrations** → [`docs/integrations/README.md`](docs/integrations/README.md)
- **AI tool install overview** → [`docs/integrations/INSTALL_OVERVIEW.md`](docs/integrations/INSTALL_OVERVIEW.md)
- **AI tool implementation guide** → [`docs/integrations/INSTALL_IMPLEMENTATION.md`](docs/integrations/INSTALL_IMPLEMENTATION.md)
- **Release and distribution guide** → [`docs/release/AGENT_DISTRIBUTION.md`](docs/release/AGENT_DISTRIBUTION.md)
- **Architecture** → [`docs/ARCHITECTURE.md`](docs/ARCHITECTURE.md)
- **Full docs map** → [`docs/INDEX.md`](docs/INDEX.md)
- **Contributing guide** → [`CONTRIBUTING.md`](CONTRIBUTING.md)

## Project links

- Code of conduct: [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md)
- Security policy: [SECURITY.md](SECURITY.md)
- Support: [SUPPORT.md](SUPPORT.md)
- Roadmap: [ROADMAP.md](ROADMAP.md)
- Quality standards: [QUALITY.md](QUALITY.md)
