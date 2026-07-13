# MCP Roadmap

This repository stays CLI-first while now also shipping a read-only stdio MCP
server through `cmd/osf-mcp`. The original roadmap boundary still applies:
MCP code wraps the reusable OSF API client directly and does not expose Cobra,
stdout formatting, or other terminal presentation internals.

## Boundary Review

The current packages are already split in a way that supports a later MCP track without dragging terminal presentation into the server layer.

### Keep Internal

- `internal/cli`: Cobra wiring, argument parsing, command help, and stdout formatting remain CLI concerns.
- `internal/output`: table and JSON emitters are presentation helpers for the CLI, not the MCP server.
- `internal/auth`: the env-backed token loader, missing-token error, and redaction helpers stay process-local while the secret source remains injected.

### Reused By MCP

- `internal/osfapi`: the OSF API client, typed resource models, pagination helpers, and API error handling back the read-only MCP tools.
- `internal/auth`: env-backed credential loading is reused by `cmd/osf-mcp`.
- `internal/mcpserver`: MCP-facing service and structured output adapter.

### Reusable Candidates Later

- `internal/download`: safe path normalization, atomic file writes, conflict policy handling, and folder-download manifest types are reusable if a later MCP track supports export/download flows.

## Next MCP Milestones

1. Read-only server facade
   - Implemented in `internal/mcpserver` and `cmd/osf-mcp`.
   - The read-only tool set includes `osf_whoami`, `osf_projects_list`, `osf_project_get`, `osf_components_list`, `osf_files_list`, `osf_contributors_list`, `osf_search`, `osf_preprints_list`, and `osf_doi_resolve`.

2. Structured tool results
   - OSF resources are mapped to JSON-style MCP structured output in `internal/mcpserver`.
   - Keep presentation helpers in `internal/cli` and `internal/output`.

3. Shared auth contract
   - Reuse `auth.Source`, `auth.LoadToken`, `auth.MissingTokenError`, and `auth.Redact` for both CLI and MCP without introducing shared secret storage.
   - Continue to pass only a bearer token string into `osfapi.Client`.

4. Download boundary decision
   - Keep `internal/download` internal until a future MCP export/download track is approved.
   - If that track lands, promote only the safe path, conflict, and manifest primitives that the server actually needs.

5. Registry/package distribution
   - Official MCP Registry metadata is prepared in `server.json`.
   - The first public package route is OCI via `Dockerfile.mcp` and GHCR.
   - Local client plugin/config examples live in `.github/mcp.json`, `.vscode/mcp.json`, and `plugins/`.

## Non-Goals

- No write-capable MCP tools until a separate track approves the safety model.
- No CLI command behavior changes are required for the MCP server.
- No package is promoted to public API during this track.
