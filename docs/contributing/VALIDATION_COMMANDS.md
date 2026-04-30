# Validation Commands

Adapters should expose a native validation command when the project has a standard toolchain.

Examples:

- Go: `go test ./...`
- Python: `pytest`
- JavaScript/TypeScript: `npm test`
- Rust: `cargo test`
- Java/Kotlin: `mvn test` or `gradle test`
- C#: `dotnet test`
- PHP: `vendor/bin/phpunit` or `vendor/bin/pest`
- Ruby: `bundle exec rspec`
- C++: `ctest`

If tooling is missing, return clear metadata instead of pretending validation passed.
