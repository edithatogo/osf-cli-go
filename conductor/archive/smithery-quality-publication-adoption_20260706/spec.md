# Spec: Smithery Quality Publication Adoption

## Overview

Validate, improve, and if appropriate republish the OSF MCP server on Smithery
using the existing MCPB route, aiming for the highest available Smithery quality
score without weakening security or overstating live behavior.

## Functional Requirements

- Verify current Smithery listing, release, MCP URL, manifest interpretation,
  quality score, and validation feedback.
- Build/validate MCPB artifacts locally or through CI.
- Align manifest metadata, tool schemas, auth fields, descriptions, categories,
  and README material with Smithery expectations.
- Use Chrome for Smithery login/publication flows when CLI/API is insufficient;
  ask the user to log in if Chrome auth cannot proceed.
- Iterate score improvements until it reaches 100% or all remaining score gaps
  are external/unavailable and documented.

## Acceptance Criteria

- Smithery listing is published or a precise authentication/platform blocker is
  recorded.
- MCPB validation, registry checks, Go tests, vet, anti-stub, and review checks
  pass.
- Score evidence and all improvement attempts are stored under the track.

## Out Of Scope

- Public Streamable HTTP implementation.
- Committing generated bundles, credentials, or screenshots containing secrets.
