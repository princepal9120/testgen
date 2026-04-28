<p align="center">
  <img src="website/images/logo.png" alt="TestGen logo" width="140">
</p>

<p align="center">
  <strong>AI-powered test generation built for developers, CI, and coding agents</strong>
</p>

<p align="center">
  <a href="https://github.com/princepal9120/testgen/actions"><img src="https://github.com/princepal9120/testgen/actions/workflows/ci.yml/badge.svg" alt="CI"></a>
  <a href="https://github.com/princepal9120/testgen/releases"><img src="https://img.shields.io/github/v/release/princepal9120/testgen" alt="Release"></a>
  <a href="https://www.apache.org/licenses/LICENSE-2.0"><img src="https://img.shields.io/badge/License-Apache%202.0-blue.svg" alt="License: Apache 2.0"></a>
  <a href="https://github.com/princepal9120/testgen"><img src="https://img.shields.io/github/stars/princepal9120/testgen?style=social" alt="GitHub stars"></a>
</p>

<p align="center">
  <a href="#installation">Install</a> &bull;
  <a href="#quick-start">Quick Start</a> &bull;
  <a href="#agent-onboarding">Agent Onboarding</a> &bull;
  <a href="docs/CLI_REFERENCE.md">CLI Reference</a> &bull;
  <a href="docs/integrations/README.md">Integrations</a>
</p>

---

TestGen is a multi-language CLI that inspects code, generates tests, validates coverage, and fits cleanly into local workflows, CI pipelines, and AI agent tooling.

The CLI, TUI, repo-local agent skills, and MCP server all use the same shared application layer. That gives humans and agents one review-first backend with machine-readable output, dry-run patches, validation, and cost-aware usage reporting.

Supported languages: **JavaScript/TypeScript, Python, Go, Rust, and Java**.

## Why TestGen

| Need | What TestGen gives you |
|------|-------------------------|
| Safe agent workflows | Dry-run generation, patch artifacts, JSON output, explicit write controls |
| Cost-aware planning | Offline estimates with `testgen analyze --cost-estimate` |
| Multi-language coverage | JS/TS, Python, Go, Rust, and Java adapters |
| CI-friendly output | Scriptable commands and machine-readable result envelopes |
| Human review loop | TUI, CLI summaries, generated artifacts, and validation metadata |
| Agent onboarding | Codex skill, Claude command, OpenCode command, and MCP stdio server |

## Installation

### Quick Install, Linux/macOS

```bash
curl -fsSL https://raw.githubusercontent.com/princepal9120/testgen/main/install.sh | bash
```

Installs to `~/.local/bin`. Add it to PATH if needed:

```bash
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.bashrc
source ~/.bashrc
```

For zsh, use `~/.zshrc` instead of `~/.bashrc`.

### Windows PowerShell

```powershell
irm https://raw.githubusercontent.com/princepal9120/testgen/main/install.ps1 | iex
```

The installer places `testgen.exe` in `%USERPROFILE%\.local\bin` and can add that directory to your user PATH.

### Go install

```bash
go install github.com/princepal9120/testgen-cli@latest
```

### Build from source

```bash
git clone https://github.com/princepal9120/testgen.git
cd testgen
go build -o testgen .
./testgen --help
```

### Pre-built binaries

Download from [releases](https://github.com/princepal9120/testgen/releases):

- macOS: `testgen-macos-x86_64` / `testgen-macos-aarch64`
- Linux: `testgen-linux-x86_64` / `testgen-linux-aarch64`
- Windows: `testgen-windows-x86_64.exe`

### Verify installation

```bash
testgen --version
testgen --help
```

## Quick Start

### 1. Set one provider API key

```bash
export ANTHROPIC_API_KEY="..."
# or OPENAI_API_KEY / GEMINI_API_KEY / GROQ_API_KEY
```

### 2. Inspect the codebase before generating

```bash
testgen analyze --path=./src --cost-estimate --output-format json
```

`--cost-estimate` is offline and does not require an API key. It uses the same provider pricing and batching assumptions as generation, so estimates stay aligned with runtime usage reporting.

### 3. Generate review-first output

```bash
testgen generate --file=./src/utils.py \
  --type=unit \
  --dry-run \
  --emit-patch \
  --report-usage \
  --output-format json
```

This is the safest default for agents and automation because it returns artifacts and patches without writing files.

### 4. Write and validate when ready

```bash
testgen generate --file=./src/utils.py \
  --type=unit \
  --validate \
  --output-format json
```

For a full folder:

```bash
testgen generate --path=./src \
  --recursive \
  --type=unit \
  --validate \
  --output-format json
```

## Agent Onboarding

TestGen ships repo-local wrappers for AI coding tools.

```bash
# From the TestGen repo
./scripts/install-agent-integrations.sh /path/to/target-repo copy
```

Installed files:

- `.codex/skills/testgen/SKILL.md`
- `.claude/commands/testgen.md`
- `.opencode/commands/testgen.md`

Use `symlink` mode when developing the wrappers locally:

```bash
./scripts/install-agent-integrations.sh /path/to/target-repo symlink
```

### Codex

```bash
mkdir -p /path/to/target-repo/.codex/skills/testgen
cp .codex/skills/testgen/SKILL.md /path/to/target-repo/.codex/skills/testgen/SKILL.md
```

Then ask Codex to use the TestGen skill inside that repo.

Recommended Codex flow:

```bash
testgen analyze --path=./src --cost-estimate --output-format json
testgen generate --path=./src --recursive --type=unit --dry-run --emit-patch --report-usage --output-format json
```

### Claude Code and OpenCode

The same install script copies command wrappers into `.claude/commands` and `.opencode/commands`. See [agent integrations](docs/integrations/README.md) for tool-specific usage.

### MCP

Run TestGen as a stdio MCP server:

```bash
testgen mcp
```

Print a config snippet:

```bash
./scripts/print-mcp-config.sh testgen
```

## How It Works

```
Developer or agent
       |
       v
 testgen analyze          offline scan, language detection, cost estimate
       |
       v
 testgen generate         provider call, generated tests, patch artifacts
       |
       v
 testgen validate         coverage and test validation checks
       |
       v
 JSON envelope            results, artifacts, patches, usage, errors
```

Core design choices:

1. **Review-first generation**. Dry-run and patch artifacts before writes.
2. **Shared application layer**. CLI, TUI, wrappers, and MCP use the same orchestration logic.
3. **Machine-readable output**. JSON envelopes are stable for agents and CI.
4. **Cost transparency**. Analyze and generate can report estimated tokens, provider costs, cache reuse, and batching behavior.
5. **Thin integrations**. Agent wrappers do not duplicate TestGen logic.

## Common Commands

### Analyze

```bash
testgen analyze --path=./src --cost-estimate
testgen analyze --path=. --cost-estimate --output-format json
```

### Generate

```bash
testgen generate --file=./src/utils.py --type=unit --dry-run --emit-patch
testgen generate --path=./src --recursive --type=unit --dry-run --emit-patch --output-format json
testgen generate --request-file=./request.json
cat request.json | testgen generate --request-file=-
```

### Validate

```bash
testgen validate --path=./src
testgen generate --file=./src/utils.py --type=unit --validate
```

### Interactive and MCP

```bash
testgen tui
testgen mcp
```

## Cost-efficiency Reporting

TestGen exposes one cost-efficiency story across analyze, generate, and saved run metrics:

- `testgen analyze --cost-estimate` reports provider-aware token and cost estimates without live API calls.
- `testgen generate --report-usage` surfaces request counts, cache reuse, batching/chunking activity, and estimated cost.
- `.testgen/metrics/*.json` stores per-run accounting snapshots for later inspection.

Recommended bulk agent command:

```bash
testgen generate --path=./src \
  --recursive \
  --dry-run \
  --emit-patch \
  --report-usage \
  --output-format json
```

## Documentation

- [CLI reference](docs/CLI_REFERENCE.md)
- [Agent and MCP integrations](docs/integrations/README.md)
- [Codex integration](docs/integrations/codex.md)
- [MCP integration](docs/integrations/mcp.md)
- [Release and distribution guide](docs/release/AGENT_DISTRIBUTION.md)
- [Architecture](docs/ARCHITECTURE.md)
- [Docs index](docs/INDEX.md)
- [Contributing](CONTRIBUTING.md)

## Project Links

- [Code of conduct](CODE_OF_CONDUCT.md)
- [Security policy](SECURITY.md)
- [Support](SUPPORT.md)
- [Roadmap](ROADMAP.md)
- [Quality standards](QUALITY.md)

## License

Apache 2.0. See [LICENSE](LICENSE).
