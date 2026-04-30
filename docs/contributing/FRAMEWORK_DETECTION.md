# Framework Detection

Framework detection should inspect repo marker files before generation.

Suggested markers:

- JavaScript/TypeScript: `package.json`, lockfiles, config files
- Python: `pyproject.toml`, `requirements.txt`, `setup.cfg`
- Go: `go.mod`, existing `_test.go` files
- Rust: `Cargo.toml`
- Java/Kotlin: `pom.xml`, `build.gradle`, `build.gradle.kts`
- C#: `.csproj`, `.sln`
- PHP: `composer.json`
- Ruby: `Gemfile`, `.rspec`
- C++: `CMakeLists.txt`, existing test folders

Detection should be best-effort and should always return a safe default when unsure.
