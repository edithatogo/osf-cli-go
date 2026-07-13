# OSF API schema discovery

## Overview

Assess whether an offline endpoint, schema, tag, and documentation discovery
surface based on the official OSF API v2 OpenAPI specification belongs in the
CLI or MCP server.

## Requirements

- Use the official `CenterForOpenScience/developer.osf.io` specification as the
  source of truth and document refresh, provenance, and licensing rules.
- Benchmark `hirakinii/osf-api-mcp` without copying incompatible implementation.
- Prefer offline deterministic search and stable machine-readable output.

## Acceptance criteria

- A dated feasibility comparison exists.
- The chosen CLI/MCP/API design is implemented or explicitly deferred with a
  linked issue and rationale.
- Generated or vendored specification handling has validation and update tests.

## Out of scope

- Replacing the runtime OSF API client with an unverified generated client.
