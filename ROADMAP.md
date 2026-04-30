# TestGen Roadmap

This roadmap focuses on reliability, extensibility, agent-native adoption, and community growth.
Timelines are directional and may shift based on contributor capacity.

For LLM-tool-specific adoption work, see [`docs/LLM_AGENT_ADOPTION.md`](docs/LLM_AGENT_ADOPTION.md).

## 2026 Q1: Open Source Foundation

- Add contributor governance and security policies
- Establish issue and PR templates with triage labels
- Publish engineering quality and testing standards
- Define release process and ownership model

## 2026 Q2: Reliability and Coverage

- Add `testgen doctor` for repo readiness checks, provider-key checks, framework detection, and suggested safe next commands
- Add `testgen capabilities --output-format=json` so LLM hosts can discover supported commands, flags, schemas, and limitations
- Increase unit coverage in `internal/generator`, `internal/llm`, `internal/validation`
- Introduce deterministic tests for prompt and output parsing
- Add regression tests for known edge cases per language adapter
- Harden error handling and CLI exit code consistency

## 2026 Q3: Architecture and Extensibility

- Refine adapter contract for easier language onboarding
- Publish JSON schemas for analyze, cost, generation, patch, validation, capabilities, and error responses
- Add framework detection for every supported language family
- Add stronger configuration validation and diagnostics
- Improve provider abstraction for retries, backoff, and rate-limit behavior
- Extend usage reporting for cost and token analytics

## 2026 Q4: Ecosystem and Adoption

- Add Cursor, Cline, Continue, Roo Code, and Gemini CLI integration docs/install targets
- Improve website and docs discoverability
- Add end-to-end sample projects for each supported language and framework
- Improve contributor onboarding with guided starter issues
- Prepare v1.0 hardening checklist and compatibility policy

## Continuous Tracks

- Security scanning and dependency hygiene
- Documentation quality and examples maintenance
- Faster CI feedback and stable developer experience
