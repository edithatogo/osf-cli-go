# MCP Roadmap

This repository stays CLI-first. The MCP work described here is design-only and is intended to preserve the current command-line shape while identifying reusable OSF core code for a later server track.

## Boundary Review

The current packages are already split in a way that supports a later MCP track without dragging terminal presentation into the server layer.

### Keep Internal

- `internal/cli`: Cobra wiring, argument parsing, command help, and stdout formatting remain CLI concerns.
- `internal/output`: table and JSON emitters are presentation helpers for the CLI, not the MCP server.
- `internal/auth`: the env-backed token loader, missing-token error, and redaction helpers stay process-local while the secret source remains injected.

### Reusable Candidates Later

- `internal/osfapi`: the OSF API client, typed resource models, pagination helpers, and API error handling are the best reuse candidate for MCP.
- `internal/download`: safe path normalization, atomic file writes, conflict policy handling, and folder-download manifest types are reusable if a later MCP track supports export/download flows.

## Next MCP Milestones

1. Read-only server facade
   - Define a small MCP-facing OSF service layer that wraps `internal/osfapi` without exposing Cobra, stdout writers, or table formatting.
   - Keep the first tool set read-only: `osf.whoami`, `osf.projects.list`, `osf.projects.get`, `osf.components.list`, `osf.files.list`, and `osf.contributors.list`.

2. Structured tool results
   - Map OSF resources to JSON-style tool responses in a new server adapter instead of reusing CLI record printers.
   - Keep presentation helpers in `internal/cli` and `internal/output`.

3. Shared auth contract
   - Reuse `auth.Source`, `auth.LoadToken`, `auth.MissingTokenError`, and `auth.Redact` for both CLI and MCP without introducing shared secret storage.
   - Continue to pass only a bearer token string into `osfapi.Client`.

4. Download boundary decision
   - Keep `internal/download` internal until a future MCP export/download track is approved.
   - If that track lands, promote only the safe path, conflict, and manifest primitives that the server actually needs.

## Non-Goals

- No MCP server implementation in this repository track.
- No CLI command changes are required for the roadmap work.
- No package is promoted to public API during this track; this document only identifies the likely future split points and the next sequence of server milestones.
