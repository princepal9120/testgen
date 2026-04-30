# Prompt Templates

Adapter prompts should tell the LLM:

- target language and framework
- test type requested
- existing code context
- naming and assertion style
- validation expectations
- whether to include mocks, fixtures, and edge cases

Prompts should avoid hidden file writes. The engine owns writing, patch output, and validation.
