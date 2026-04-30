# TestGen JSON Schemas

These schemas document the agent-readable JSON surfaces used by TestGen commands.

Compatibility rule:

- Required top-level fields stay stable within `api_version: v1`.
- New optional fields may be added without a breaking version bump.
- Agents should ignore unknown fields.
- Error responses should use the shared error envelope shape.

Current schemas:

- `common.schema.json`
- `doctor.response.schema.json`
- `capabilities.response.schema.json`
- `languages.response.schema.json`
- `generate.response.schema.json`
- `analyze.response.schema.json`
- `validate.response.schema.json`
- `error-envelope.schema.json`
