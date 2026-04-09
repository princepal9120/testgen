# TestGen Architecture

**Scope:** This document explains the implementation architecture of TestGen. Use [`README.md`](../README.md) for the high-level product overview and [`docs/CLI_REFERENCE.md`](./CLI_REFERENCE.md) for command usage.

## Overview

TestGen follows a layered architecture with a shared application service between human UIs and agent-facing wrappers.

```
┌─────────────────────────────────────────────────────┐
│            Presentation / Wrapper Layer             │
│ ┌──────────┐ ┌──────────┐ ┌──────────────┐          │
│ │ CLI      │ │ TUI      │ │ Agent skills │          │
│ └────┬─────┘ └────┬─────┘ └──────┬───────┘          │
└──────┼────────────┼──────────────┼──────────────────┘
       └────────────┴───────┬──────┘
                            ▼
┌─────────────────────────────────────────────────────┐
│          Application Service (internal/app/)        │
│  Shared generate / analyze / validate orchestration │
└──────────────────────┬──────────────────────────────┘
                       ▼
┌─────────────────────────────────────────────────────┐
│              Core Engine (internal/generator/)      │
│     Generates artifacts, then materializes writes   │
└──────────────────────┬──────────────────────────────┘
                       │
       ┌───────────────┼───────────────┐
       ▼               ▼               ▼
┌───────────┐   ┌───────────┐   ┌───────────┐
│  Scanner  │   │  Adapters │   │    LLM    │
│(internal/)│   │(internal/)│   │(internal/)│
└───────────┘   └───────────┘   └───────────┘
```

---

## Package Responsibilities

### `cmd/`
- Cobra command definitions
- Flag parsing and validation
- Calls `internal/app`
- **Minimal business logic**

### `internal/app/`
- Shared application-layer request/response contracts
- Generate/analyze/validate orchestration
- Shared machine-readable output for CLI, TUI, and agent wrappers

### `internal/ui/tui/`
- Bubble Tea TUI application
- Screen models (Home, Config, Preview, Running, Results)
- State machine for navigation
- Uses lipgloss for styling

### `internal/ui/`
- Shared UI components (spinner, banner, progress)
- Style definitions

### `internal/scanner/`
- File discovery
- Language detection
- Ignore pattern handling

### `internal/adapters/`
- `LanguageAdapter` interface
- Language-specific implementations (Go, Python, JavaScript/TypeScript, Rust, Java)
- Parsing, prompts, formatting

### `internal/llm/`
- `Provider` interface
- Anthropic/OpenAI implementations
- Caching, rate limiting, batching

### `internal/generator/`
- Core orchestration
- Worker pool for parallelism
- Output handling

### `internal/validation/`
- Test compilation checks
- Coverage parsing

### `pkg/models/`
- Shared data structures
- DTOs between packages

---

## Key Interfaces

### LanguageAdapter
```go
type LanguageAdapter interface {
    ParseFile(content string) (*models.AST, error)
    GetPromptTemplate(testType string) string
    GenerateTestPath(sourcePath string, outputDir string) string
    FormatTestCode(code string) (string, error)
    ValidateTests(testCode string, testPath string) error
    RunTests(testDir string) (*models.TestResults, error)
}
```

### LLM Provider
```go
type Provider interface {
    Complete(ctx context.Context, req CompletionRequest) (*CompletionResponse, error)
    BatchComplete(ctx context.Context, reqs []CompletionRequest) ([]*CompletionResponse, error)
}
```

---

## Data Flow

```
Caller → App Service → Scanner → Adapter.Parse → Engine → LLM → Adapter.Format → Artifact → Write/Validate
```

1. **App service** resolves the request shape
2. **Scanner** discovers source files
3. **Adapter** parses file into AST
4. **Engine** builds prompts using adapter templates
5. **LLM** generates test code
6. **Engine** returns artifacts first
7. **App service / engine** writes and validates when requested

---

## Adding a New Language

1. Create `internal/adapters/<lang>.go`
2. Implement `LanguageAdapter` interface
3. Register in `internal/adapters/registry.go`:
   ```go
   defaultRegistry.Register(NewRubyAdapter())
   ```

No changes needed in CLI, engine, or LLM layers.
