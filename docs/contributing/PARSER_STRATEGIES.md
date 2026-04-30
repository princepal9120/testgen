# Parser Strategies

Prefer a real parser when a stable Go package exists and the language syntax is difficult to parse with regular expressions.

Use the shared regex adapter only for pragmatic first support when:

- the language has simple function signatures for the target fixtures
- framework detection does not require deep AST data
- false positives are covered by regression tests

Parser tests should cover:

- free functions
- class or receiver methods
- imports
- comments around definitions
- constructors or main functions that should be skipped
- async, error, or generic signatures where supported
