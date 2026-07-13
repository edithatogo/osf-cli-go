# OSF CLI Go 1.0 hardening and maturity

## Overview

Raise the repository from release-ready 0.x software to a defensible 1.0
launch candidate through measurable reliability, security, compatibility,
operations, and support gates.

## Requirements

- Define compatibility, deprecation, API/MCP schema, and migration contracts.
- Add fuzz/property, contract, performance, cancellation, and cross-platform
  tests for high-risk paths.
- Harden supply chain, provenance, SBOM, signing, secrets, containers, and
  dependency updates.
- Document observability, support, incident response, rollback, release
  cadence, maintainer duties, and live OSF validation cleanup.

## Acceptance criteria

- Every `docs/v1-launch-roadmap.md` gate has dated evidence or explicit waiver.
- No known release-blocking security or compatibility finding remains.
- A clean release-candidate review passes before `v1.0.0` is tagged.
