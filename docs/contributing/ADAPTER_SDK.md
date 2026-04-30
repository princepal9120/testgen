# Adapter SDK

Use this guide when adding or improving a TestGen language adapter.

A language adapter should implement the `LanguageAdapter` contract in `internal/adapters/adapter.go`:

- detect whether a file belongs to the language
- parse source into functions, methods, classes, imports, and metadata
- generate the expected test path
- describe supported frameworks
- provide prompts for unit, edge-case, negative, and integration tests
- expose the default validation command where possible

Minimum PR checklist:

1. Add or update the adapter under `internal/adapters/`.
2. Register it in `internal/adapters/registry.go`.
3. Add scanner extensions and test-file ignore rules.
4. Add parser fixtures for functions, methods, imports, classes, errors, and async behavior where relevant.
5. Add example source under `examples/` or `examples/projects/`.
6. Update `testgen languages`, docs, and website copy.
7. Run `go test ./...` and `golangci-lint run ./...`.
