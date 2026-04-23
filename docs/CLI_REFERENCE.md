# TestGen CLI Reference

**Scope:** This document is the current reference for TestGen commands, flags, configuration entry points, and environment variables. Use [`README.md`](../README.md) for the high-level overview and [`docs/ARCHITECTURE.md`](./ARCHITECTURE.md) for implementation details.

Complete reference for all TestGen commands and options.

## Shared machine-readable behavior

- Prefer `--output-format json` for CI, wrappers, and agent callers.
- Prefer `--dry-run --emit-patch` when the caller should inspect artifacts before any file write.
- The CLI, TUI, and MCP server share the same orchestration layer, so JSON payloads and generation behavior stay aligned across surfaces.
- In JSON machine mode, commands should return the shared outer envelope on stdout and suppress Cobra usage/error banners on stderr.
- Cost/usage transparency is additive: enabling `--report-usage` or `--cost-estimate` adds fields and summaries without renaming or removing the existing envelope keys.

## Global Flags

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--config` | | Path to config file | `.testgen.yaml` |
| `--verbose` | `-v` | Enable debug output | `false` |
| `--quiet` | `-q` | Suppress non-error output | `false` |

---

## `testgen generate`

Generate tests for source files.

### Usage
```bash
testgen generate [flags]
```

### Flags

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--path` | `-p` | Source directory | - |
| `--file` | | Single source file | - |
| `--type` | `-t` | Test types (comma-separated) | `unit` |
| `--framework` | `-f` | Target test framework | auto-detect |
| `--output` | `-o` | Output directory | same as source |
| `--recursive` | `-r` | Process recursively | `false` |
| `--parallel` | `-j` | Number of workers | `2` |
| `--request-file` | | Read a machine request from a JSON file (`-` reads stdin) | - |
| `--dry-run` | | Preview without writing | `false` |
| `--emit-patch` | | Include structured patch operations in shared/JSON output | `false` |
| `--interactive` | `-i` | Show interactive results view after generation | `false` |
| `--validate` | | Run tests after generation | `false` |
| `--output-format` | | Output format (text/json) | `text` |
| `--include-pattern` | | Glob pattern to include | - |
| `--exclude-pattern` | | Glob pattern to exclude | - |
| `--batch-size` | | API batch size | `5` |
| `--report-usage` | | Generate usage report | `false` |

### Test Types
- `unit` - Basic unit tests
- `edge-cases` - Boundary conditions
- `negative` - Error handling
- `table-driven` - Parameterized tests (Go)
- `integration` - With mocked dependencies

### Examples
```bash
# Single file
testgen generate --file=./src/utils.py

# Directory with multiple test types
testgen generate --path=./src -r --type=unit,edge-cases

# Dry run with JSON output
testgen generate --path=./src -r --dry-run --output-format=json

# Machine request from file or stdin
testgen generate --request-file=./request.json
cat request.json | testgen generate --request-file=-

# Dry run with agent-ready patch output
testgen generate --path=./src -r --dry-run --emit-patch --output-format=json
```

### Usage reporting (`--report-usage`)

When enabled, TestGen emits a shared usage summary for the current generation run:

- **Text mode** keeps the human summary concise while adding request/count/cost transparency.
- **JSON mode** keeps the existing response envelope and adds an additive usage block so current consumers remain compatible.
- Usage details are intended to cover request count, cache reuse, cached-token savings, batch count, chunk count, selected provider/model, and estimated cost when that data is available.
- Per-run snapshots are also persisted under `.testgen/metrics/` for later review.

Recommended machine-readable example:

```bash
testgen generate --path=./src \
  -r \
  --dry-run \
  --emit-patch \
  --report-usage \
  --output-format=json
```

### Machine-readable / agent-safe examples

```bash
# Review-first JSON output for automation
testgen generate --file=./src/utils.py \
  --type=unit \
  --dry-run \
  --emit-patch \
  --output-format=json

# Explicitly write files and validate them
testgen generate --file=./src/utils.py \
  --type=unit \
  --validate \
  --output-format=json

# Explicit machine-input lane
testgen generate --request-file=./request.json
cat request.json | testgen generate --request-file=-
```

---

## `testgen tui`

Launch the interactive terminal UI.

### Usage
```bash
testgen tui
```

### Notes
- Guided keyboard-driven flow for generate/analyze actions
- Shows config forms, command preview, live progress, and results
- Good default for first-time human users

---

## `testgen validate`

Validate existing tests and coverage.

### Usage
```bash
testgen validate [flags]
```

### Flags

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--path` | `-p` | Directory to validate | `.` |
| `--recursive` | `-r` | Check recursively | `true` |
| `--min-coverage` | | Minimum coverage % | `0` |
| `--fail-on-missing-tests` | | Exit 1 if tests missing | `false` |
| `--report-gaps` | | Show coverage gaps | `false` |
| `--output-format` | | Output format | `text` |

### Examples
```bash
# Basic validation
testgen validate --path=./src

# Enforce 80% coverage
testgen validate --path=./src --min-coverage=80 --fail-on-missing-tests
```

---

## `testgen analyze`

Analyze codebase before generation.

### Usage
```bash
testgen analyze [flags]
```

### Flags

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--path` | `-p` | Directory to analyze | `.` |
| `--cost-estimate` | | Show estimated API cost | `false` |
| `--detail` | | Detail level (`summary`, `per-file`) | `summary` |
| `--recursive` | `-r` | Analyze recursively | `true` |
| `--output-format` | | Output format | `text` |

### Detail Levels
- `summary` - Total counts
- `per-file` - File-by-file breakdown

### Examples
```bash
# Quick cost estimate
testgen analyze --path=./src --cost-estimate

# Detailed per-file analysis
testgen analyze --path=./src --detail=per-file --output-format=json

# Machine-readable validation failure envelope
testgen validate --path=./src --fail-on-missing-tests --output-format=json
```

### Provider-aware cost estimates (`--cost-estimate`)

`testgen analyze --cost-estimate` is designed to stay review-first:

- Runs offline and does not require live provider credentials.
- Uses the same pricing and batching assumptions as generation/reporting so estimates and runtime usage stay aligned.
- Adds provider-aware totals at the top level and can include per-file token estimates when `--detail=per-file` is selected.
- Keeps text output readable and JSON output backward-compatible by adding fields instead of replacing the shared response contract.

---

## `testgen mcp`

Run TestGen as an MCP server over stdio.

### Usage
```bash
testgen mcp
```

### Exposed tools
- `testgen_generate`
- `testgen_analyze`
- `testgen_validate`

### `testgen_generate` arguments

| Argument | Description |
|----------|-------------|
| `path` / `file` | Target directory or single file |
| `types` | Test types array, defaults to `["unit"]` |
| `dry_run` | Preview artifacts without writing files |
| `validate` | Validate generated tests |
| `emit_patch` | Include structured patch operations |
| `parallelism` | Parallel worker count |
| `batch_size` | Provider batch size |
| `provider` | Explicit provider override |
| `write_files` | Required when an MCP client wants writes instead of the safe dry-run default |

### Notes
- Uses the same orchestration path as the CLI/TUI
- Safe dry-run generation is the recommended default for agent clients
- `testgen_generate` stays in dry-run mode unless the caller explicitly sets `write_files: true`
- `testgen_generate` can opt into additive usage transparency with the same `report_usage`/runtime contract used by the CLI layer when available
- MCP tool results return JSON text inside the tool response content

---

## Configuration

Use `--config` to point at a specific config file, or keep project defaults in `.testgen.yaml`.

Example:

```yaml
llm:
  provider: anthropic
  model: claude-3-5-sonnet-20241022
  temperature: 0.3

generation:
  batch_size: 5
  parallel_workers: 4
  timeout_seconds: 30

output:
  format: text
  include_coverage: true
```

## Environment Variables

| Variable | Description |
|----------|-------------|
| `ANTHROPIC_API_KEY` | Anthropic Claude API key |
| `OPENAI_API_KEY` | OpenAI GPT API key |
| `GEMINI_API_KEY` | Google Gemini API key |
| `GROQ_API_KEY` | Groq Cloud API key |
| `TESTGEN_LLM_PROVIDER` | Default LLM provider |
| `TESTGEN_LLM_MODEL` | Default model |

---

## Exit Codes

| Code | Meaning |
|------|---------|
| `0` | Success |
| `1` | Error during execution |
| `2` | Validation/coverage failure |
