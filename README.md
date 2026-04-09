# TestGen

<p align="center">
  <img src="website/images/logo.png" alt="TestGen Logo" width="120" />
</p>

**AI-Powered Multi-Language Test Generation CLI**

TestGen automatically generates production-ready tests for source code across JavaScript/TypeScript, Python, Go, Rust, and Java using LLM APIs (Anthropic Claude, OpenAI GPT, Google Gemini, Groq).

```
 ████████╗███████╗███████╗████████╗ ██████╗ ███████╗███╗   ██╗
 ╚══██╔══╝██╔════╝██╔════╝╚══██╔══╝██╔════╝ ██╔════╝████╗  ██║
    ██║   █████╗  ███████╗   ██║   ██║  ███╗█████╗  ██╔██╗ ██║
    ██║   ██╔══╝  ╚════██║   ██║   ██║   ██║██╔══╝  ██║╚██╗██║
    ██║   ███████╗███████║   ██║   ╚██████╔╝███████╗██║ ╚████║
    ╚═╝   ╚══════╝╚══════╝   ╚═╝    ╚═════╝ ╚══════╝╚═╝  ╚═══╝
 
                     Universal TEST Generator

  ```
## Features

- 🖥️ **Interactive TUI Mode**: Full terminal UI with visual forms and live progress
- 🌍 **Multi-Language Support**: JavaScript/TypeScript, Python, Go, Rust, Java
- 🧪 **Multiple Test Types**: Unit, edge-cases, negative, table-driven, integration
- 🔌 **Framework Aware**: Jest, Vitest, pytest, Go testing, cargo test, JUnit
- 💰 **Cost Optimized**: Semantic caching, request batching
- 🔧 **CI/CD Ready**: JSON output, meaningful exit codes, quiet mode
- 🤖 **Agent Ready**: Shared JSON contract with structured artifacts and patch operations
- 🏗️ **Clean Architecture**: Extensible adapter pattern

## Installation

### Quick Install (Recommended)

**macOS / Linux:**

```bash
curl -fsSL https://raw.githubusercontent.com/princepal9120/testgen/main/install.sh | bash
```

**Windows (PowerShell):**

```powershell
irm https://raw.githubusercontent.com/princepal9120/testgen/main/install.ps1 | iex
```

### From Source

```bash
# Clone the repository
git clone https://github.com/princepal9120/testgen.git
cd testgen

# Build
go build -o testgen .

# Install globally (optional)
go install .
```

### Binary Releases

Download pre-built binaries from [GitHub Releases](https://github.com/princepal9120/testgen/releases).

Supported platforms:
- **Linux**: x86_64, aarch64
- **macOS**: x86_64, aarch64 (Apple Silicon)
- **Windows**: x86_64

> Source, installers, and release artifacts live in [`princepal9120/testgen`](https://github.com/princepal9120/testgen). The Go module path remains `github.com/princepal9120/testgen-cli`, so `go install github.com/princepal9120/testgen-cli@latest` is still the correct Go-based install command.

## Quick Start

### Step 1: Get an API Key

Choose a provider and get your API key:

| Provider | Get API Key | Best For |
|----------|-------------|----------|
| **Anthropic Claude** | [console.anthropic.com](https://console.anthropic.com/) | Best quality |
| **OpenAI GPT** | [platform.openai.com](https://platform.openai.com/api-keys) | Most popular |
| **Google Gemini** | [aistudio.google.com](https://aistudio.google.com/app/apikey) | Free tier |
| **Groq** | [console.groq.com](https://console.groq.com/keys) | Fastest, free tier |

### Step 2: Set Your API Key

```bash
# Choose ONE provider and set the environment variable:

# Anthropic Claude (recommended)
export ANTHROPIC_API_KEY="sk-ant-api03-xxxxx"

# OpenAI GPT
export OPENAI_API_KEY="sk-xxxxx"

# Google Gemini (free tier available)
export GEMINI_API_KEY="AIzaSyxxxxx"

# Groq (fastest, free tier)
export GROQ_API_KEY="gsk_xxxxx"
```

> 💡 **Tip**: Add this to your `~/.bashrc` or `~/.zshrc` to persist across sessions.

### Step 3: Generate Tests

```bash
# Launch interactive TUI mode
testgen tui

# Or use CLI commands directly:

# Generate tests for a single file
testgen generate --file=./src/utils.py --type=unit

# Generate tests for a directory recursively
testgen generate --path=./src --recursive --type=unit,edge-cases

# Preview without writing files
testgen generate --path=./src --dry-run

# Preview with structured agent-friendly patch output
testgen generate --path=./src --dry-run --emit-patch --output-format json

# Analyze cost before generation
testgen analyze --path=./src --cost-estimate
```

### Choose how you want to use TestGen

You can use TestGen in four practical ways:

1. **Interactive TUI**
   - Run `testgen tui`
   - Best when you want guided prompts, keyboard navigation, and live progress

2. **Direct CLI commands**
   - Run `testgen generate`, `testgen analyze`, and `testgen validate`
   - Best for local development, scripts, and CI

3. **Agent wrappers**
   - Use the repo-local Codex, Claude Code, or OpenCode wrappers
   - Best when an AI coding agent should inspect dry-run artifacts before writing files

4. **MCP server**
   - Run `testgen mcp`
   - Best when your MCP client prefers tool calling over shelling out directly

Recommended first run:

```bash
# Inspect a codebase before spending tokens
testgen analyze --path=./examples --cost-estimate --recursive

# Generate review-first artifacts without writing files
testgen generate --file=./examples/python/calculator.py \
  --type=unit \
  --dry-run \
  --emit-patch \
  --output-format json

# Write files only when you want validation feedback
testgen generate --path=./src --recursive --type=unit --validate
```

## Agent integrations

TestGen now exposes a shared machine-readable contract for agent wrappers.

- Codex example skill: `.codex/skills/testgen/SKILL.md`
- Claude Code command: `.claude/commands/testgen.md`
- OpenCode command: `.opencode/commands/testgen.md`
- OpenCode notes: `docs/integrations/opencode.md`
- MCP server notes: `docs/integrations/mcp.md`

Recommended safe mode for agents:

```bash
testgen generate --file=./src/utils.py --type=unit --dry-run --emit-patch --output-format json
```

Experimental MCP server:

```bash
testgen mcp
```

Agent integration surfaces:

1. **Codex / oh-my-codex**
   - add or vendor `.codex/skills/testgen/SKILL.md`
   - invoke TestGen through the shared JSON CLI contract

2. **Claude Code**
   - use `.claude/commands/testgen.md`
   - run review-first dry-run generation by default

3. **OpenCode**
   - use `.opencode/commands/testgen.md`
   - or connect through `testgen mcp`

4. **Direct MCP clients**
   - run `testgen mcp`
   - call `testgen_generate`, `testgen_analyze`, and `testgen_validate` over stdio

Install those wrappers into another repo:

```bash
./scripts/install-agent-integrations.sh /path/to/target-repo copy
```

That installs the exact files each agent surface expects:

- **Codex / oh-my-codex** → `.codex/skills/testgen/SKILL.md`
- **Claude Code** → `.claude/commands/testgen.md`
- **OpenCode** → `.opencode/commands/testgen.md`

Manual install also works if you only want one surface:

```bash
# Codex
mkdir -p /path/to/repo/.codex/skills/testgen
cp .codex/skills/testgen/SKILL.md /path/to/repo/.codex/skills/testgen/SKILL.md

# Claude Code
mkdir -p /path/to/repo/.claude/commands
cp .claude/commands/testgen.md /path/to/repo/.claude/commands/testgen.md

# OpenCode
mkdir -p /path/to/repo/.opencode/commands
cp .opencode/commands/testgen.md /path/to/repo/.opencode/commands/testgen.md
```

Print an MCP client config snippet:

```bash
./scripts/print-mcp-config.sh testgen
```

## Commands

### `testgen tui`

Launch the interactive Terminal User Interface.

```bash
testgen tui
```

**Features:**
- Visual home screen to choose actions
- Interactive config forms (path, types, parallel, dry-run, validate)
- Command preview before execution
- Live progress with spinner and file-by-file updates
- Results summary with generated file paths

**Controls:**
| Key | Action |
|-----|--------|
| Tab / Shift+Tab | Navigate fields |
| Space | Toggle options |
| Enter | Confirm / Select |
| Esc | Go back |
| q / Ctrl+C | Quit |
| Ctrl+X | Cancel operation |

### `testgen generate`

Generate tests for source files.

```bash
testgen generate [OPTIONS]

Options:
  -p, --path string           Source directory to generate tests for
      --file string           Single source file to generate tests for
  -t, --type strings          Test types: unit, edge-cases, negative, table-driven, integration (default [unit])
  -f, --framework string      Target test framework (auto-detected by default)
  -o, --output string         Output directory for generated tests
  -r, --recursive             Process directories recursively
  -j, --parallel int          Number of parallel workers (default 2)
      --dry-run               Preview output without writing files
      --validate              Run generated tests after creation
      --output-format string  Output format: text, json (default "text")
      --include-pattern       Glob pattern for files to include
      --exclude-pattern       Glob pattern for files to exclude
      --batch-size int        Batch size for API requests (default 5)
      --report-usage          Generate usage/cost report
```

### `testgen validate`

Validate existing tests and coverage.

```bash
testgen validate [OPTIONS]

Options:
  -p, --path string           Directory to validate (default ".")
  -r, --recursive             Check recursively (default true)
      --min-coverage float    Minimum coverage percentage (0-100)
      --fail-on-missing-tests Exit with error if tests missing
      --report-gaps           Show coverage gaps per file
      --output-format string  Output format: text, json (default "text")
```

### `testgen analyze`

Analyze codebase for test generation cost estimation.

```bash
testgen analyze [OPTIONS]

Options:
  -p, --path string           Directory to analyze (default ".")
      --cost-estimate         Show estimated API costs
      --detail string         Detail level: summary, per-file, per-function (default "summary")
  -r, --recursive             Analyze recursively (default true)
      --output-format string  Output format: text, json (default "text")
```

## Configuration

Create a `.testgen.yaml` file in your project root:

```yaml
llm:
  provider: anthropic        # anthropic, openai, gemini, or groq
  model: claude-3-5-sonnet-20241022
  # Models per provider:
  #   anthropic: claude-3-5-sonnet-20241022
  #   openai: gpt-4-turbo-preview
  #   gemini: gemini-1.5-pro, gemini-1.5-flash
  #   groq: llama-3.3-70b-versatile, mixtral-8x7b-32768
  temperature: 0.3

generation:
  batch_size: 5
  parallel_workers: 4
  timeout_seconds: 30

output:
  format: text
  include_coverage: true

languages:
  javascript:
    frameworks: [jest, vitest]
    default_framework: jest
  python:
    frameworks: [pytest, unittest]
    default_framework: pytest
  go:
    frameworks: [testing]
  rust:
    frameworks: [cargo-test]
```

## Environment Variables

| Variable | Description |
|----------|-------------|
| `ANTHROPIC_API_KEY` | Anthropic Claude API key |
| `OPENAI_API_KEY` | OpenAI GPT API key |
| `GEMINI_API_KEY` | Google Gemini API key |
| `GROQ_API_KEY` | Groq Cloud API key |
| `TESTGEN_LLM_PROVIDER` | Default LLM provider (anthropic, openai, gemini, groq) |
| `TESTGEN_LLM_MODEL` | Default model |

## Supported Languages

| Language | Extensions | Default Framework | Test Types |
|----------|------------|-------------------|------------|
| JavaScript/TypeScript | `.js`, `.ts`, `.jsx`, `.tsx` | Jest | unit, edge-cases, negative |
| Python | `.py` | pytest | unit, edge-cases, negative |
| Go | `.go` | testing + testify | unit, table-driven, edge-cases, negative |
| Rust | `.rs` | cargo test | unit, edge-cases, negative |
| Java | `.java` | JUnit 5 | unit, edge-cases, negative |

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | Internal/generation error |
| 2 | Validation/coverage failure |

## CI/CD Integration

### GitHub Actions

```yaml
name: Generate Tests
on: [pull_request]

jobs:
  testgen:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.25.9'
      - name: Install TestGen
        run: go install github.com/princepal9120/testgen-cli@latest
      - name: Generate tests
        env:
          ANTHROPIC_API_KEY: ${{ secrets.ANTHROPIC_API_KEY }}
        run: |
          testgen generate --path=./src \
            --recursive \
            --type=unit \
            --output-format=json
```

## Development

```bash
# Run tests
go test ./... -v

# Build
go build -o testgen .

# Run linter
golangci-lint run

# Run local CI quality checks
make ci
```

## Community and Governance

- Contributing guide: [CONTRIBUTING.md](CONTRIBUTING.md)
- Code of conduct: [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md)
- Security policy: [SECURITY.md](SECURITY.md)
- Support channels: [SUPPORT.md](SUPPORT.md)
- Roadmap: [ROADMAP.md](ROADMAP.md)

## Engineering Quality

Quality and release standards are documented in:

- [QUALITY.md](QUALITY.md)
- [docs/TESTING_STRATEGY.md](docs/TESTING_STRATEGY.md)

## License

Apache 2.0 - See [LICENSE](LICENSE) for details.
