# Zenodo DOI publication state machine

## Objective

Make DOI reservation, publication, versioning, discard, access, embargo, and
license behavior explicit and safe at every CLI, API, and MCP boundary.

## Requirements

- Model valid and terminal transitions and reject invalid transitions locally.
- Require explicit authorization, dry-run where meaningful, confirmation, and
  auditable evidence for publication or destructive actions.
- Validate access, embargo, license, metadata completeness, and token scopes.
- Never imply that published files can be casually replaced or removed.

## Completion evidence

State-machine and command tests cover every transition, denial, confirmation,
redaction, recovery boundary, and sandbox-validated lifecycle path.
