# MCP Roadmap

This repository stays CLI-first. The MCP work described here is design-only and is intended to preserve the current command-line shape while identifying reusable OSF core code for a later server track.

## Boundary Review

The current packages separate cleanly into CLI presentation, auth/token handling, OSF API access, and download safety.

### Keep Internal

- `internal/cli`: Cobra wiring, argument parsing, command help, and stdout formatting remain CLI concerns.
- `internal/output`: table and JSON emitters are presentation helpers for the CLI, not the MCP server.
- `internal/auth`: the env-backed token loader and redaction helpers should stay internal while the secret source remains process-local.

### Candidate Public Packages Later

- `internal/osfapi`: the OSF API client, typed resource models, pagination helpers, and API error handling are the best reuse candidate for MCP.
- `internal/download`: safe path normalization, atomic file writes, conflict policy handling, and folder-download manifest types are reusable if a later MCP track supports export/download flows.

## Read-Only MCP Tool Inventory

The first MCP server should expose read-only OSF metadata tools only.

- `osf.whoami`
- `osf.projects.list`
- `osf.projects.get`
- `osf.components.list`
- `osf.files.list`
- `osf.contributors.list`

Each tool should return structured JSON-style results rather than terminal tables.

## Auth Sharing Model

CLI auth and MCP auth should share the same token-loading and redaction behavior without sharing any secret storage.

- Shared core: token lookup abstraction, missing-token error, and redaction helpers.
- CLI adapter: `OSF_TOKEN` via `auth.EnvSource`.
- MCP adapter: host-provided lookup or session resolver, passed in at runtime.
- Secret boundary: only the bearer token string crosses into `osfapi.Client`; no token should be written to files, configs, or logs.

## Non-Goals

- No MCP server implementation in this repository track.
- No CLI command changes are required for the roadmap work.
- No package is promoted to public API during this track; this document only identifies the likely future split points.
