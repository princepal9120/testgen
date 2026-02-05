# TestGen Roadmap

This roadmap focuses on reliability, extensibility, and community adoption.
Timelines are directional and may shift based on contributor capacity.

## 2026 Q1: Open Source Foundation

- Add contributor governance and security policies
- Establish issue and PR templates with triage labels
- Publish engineering quality and testing standards
- Define release process and ownership model

## 2026 Q2: Reliability and Coverage

- Increase unit coverage in `internal/generator`, `internal/llm`, `internal/validation`
- Introduce deterministic tests for prompt and output parsing
- Add regression tests for known edge cases per language adapter
- Harden error handling and CLI exit code consistency

## 2026 Q3: Architecture and Extensibility

- Refine adapter contract for easier language onboarding
- Add stronger configuration validation and diagnostics
- Improve provider abstraction for retries, backoff, and rate-limit behavior
- Extend usage reporting for cost and token analytics

## 2026 Q4: Ecosystem and Adoption

- Improve website and docs discoverability
- Add end-to-end sample projects for each supported language
- Improve contributor onboarding with guided starter issues
- Prepare v1.0 hardening checklist and compatibility policy

## Continuous Tracks

- Security scanning and dependency hygiene
- Documentation quality and examples maintenance
- Faster CI feedback and stable developer experience
