# TestGen

<p align="center">
  <img src="website/images/logo.png" alt="TestGen Logo" width="120" />
</p>

**AI-powered test generation for humans, CI pipelines, and coding agents.**

TestGen is a multi-language CLI for inspecting code, generating tests, validating coverage, and fitting cleanly into local workflows, CI, and agent tooling.

Supported languages: **JavaScript/TypeScript, Python, Go, Rust, and Java**.

## Why teams use TestGen

- **Start safely** with `testgen analyze` and dry-run generation before writing files
- **Work in the terminal** with either direct CLI commands or the interactive TUI
- **Integrate with agents** through JSON output, patch artifacts, and MCP
- **Keep workflows scriptable** for CI, automation, and repeatable review-first usage

## Quick start

### 1. Install

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

### 5. Write and validate when ready

```bash
testgen generate --path=./src --recursive --type=unit --validate
```

## Where next

- **Full command and flag reference** → [`docs/CLI_REFERENCE.md`](docs/CLI_REFERENCE.md)
- **Agent and MCP integrations** → [`docs/integrations/README.md`](docs/integrations/README.md)
- **Architecture** → [`docs/ARCHITECTURE.md`](docs/ARCHITECTURE.md)
- **Full docs map** → [`docs/INDEX.md`](docs/INDEX.md)
- **Contributing guide** → [`CONTRIBUTING.md`](CONTRIBUTING.md)

## Project links

- Code of conduct: [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md)
- Security policy: [SECURITY.md](SECURITY.md)
- Support: [SUPPORT.md](SUPPORT.md)
- Roadmap: [ROADMAP.md](ROADMAP.md)
- Quality standards: [QUALITY.md](QUALITY.md)
