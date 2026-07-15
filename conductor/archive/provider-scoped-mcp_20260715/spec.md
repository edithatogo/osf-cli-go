# Provider-scoped MCP tools and compatibility fixtures

## Objective

Expose stable, discoverable Zenodo read capabilities to agents without changing
existing OSF MCP semantics or advertising deferred write operations.

## Requirements

- Provide capability discovery and provider-qualified arguments or tool names.
- Keep schemas stable, errors typed, outputs lossless, and observability redacted.
- Update registry claims only for executable, fixture-backed behavior.

## Completion evidence

MCP fixtures cover negotiation, success, unsupported operations, validation,
errors, redaction, and unchanged OSF compatibility.
