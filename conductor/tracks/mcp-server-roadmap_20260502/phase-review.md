# Phase Review

## Track

- Track: `mcp-server-roadmap_20260502`
- Phase: Package boundary review and MCP design
- Date: 2026-05-02

## Implemented Behavior

- Roadmap/design only. No MCP server implementation, MCP handlers, or Go package changes were made in this track.
- Reviewed the current CLI/API/auth/download foundation surface and recorded which packages should stay internal versus which could be promoted later for MCP reuse.
- Drafted a read-only MCP tool inventory for OSF metadata operations and the auth-sharing model for CLI and MCP.

## Package Boundary Review

- `internal/cli`: remain internal. Cobra wiring, command routing, help text, and terminal output shape are CLI-specific and should not become part of the MCP surface.
- `internal/output`: remain internal. The current JSON/table emitters are presentation helpers for the CLI; MCP should return structured tool results instead of depending on terminal formatting.
- `internal/auth`: remain internal for now. The env-backed token loader and redaction helpers are useful, but the secret source adapter should stay process-local. A future shared auth core can be split out if MCP needs the same token-loading semantics.
- `internal/osfapi`: candidate to become public later. This is the cleanest shared surface for a future MCP server because it already owns the OSF API client, pagination, typed resources, and API error shaping.
- `internal/download`: candidate to become public later. The safe path normalization, atomic write logic, conflict policy, and folder download manifest are reusable for a later MCP export/download flow, even though no MCP write surface is being built now.

## MCP Tool Inventory

- `osf.whoami`: return the authenticated user profile from `users/me`.
- `osf.projects.list`: list the authenticated user’s accessible projects.
- `osf.projects.get`: fetch a project or component by GUID or OSF URL.
- `osf.components.list`: list child components for a project or component.
- `osf.files.list`: list OSF Storage files for a project or component.
- `osf.contributors.list`: list contributors for a project or component.

## Auth Sharing Model

- CLI and MCP should share token-loading rules, missing-token errors, and redaction logic through a small auth core that does not own any secret store.
- Secret material should enter only through injected `Source`/lookup adapters or an explicit token string at client-construction time.
- The CLI can keep using `auth.EnvSource` for `OSF_TOKEN`; an MCP host can supply its own source implementation or session-bound token resolver without writing secrets into repo files, logs, or tool arguments.
- `osfapi.Client` should continue to receive only the bearer token value, not the source of that token.

## Anti-Stub Evidence

- `go run ./tools/checkstubs` result: passed.
- Production markers found: none.
- Ignored paths verified: `tools/checkstubs` package tests passed.
- Self-scan exclusion verified: `tools/checkstubs` package tests passed.
- Validation evidence link or location: local repository state on 2026-05-02.

## Validation Commands

```powershell
go run ./tools/checkstubs
git diff --check
```

## Conductor Review

- Review command: roadmap/design review performed against the documented package boundaries and MCP tool inventory. There was no Go implementation to code-review in this track.
- Blocking findings: none for the documented roadmap/design scope.
- Fixes applied: clarified package boundaries, MCP tool inventory, and auth-sharing constraints.
- Re-review result: no blocking roadmap/design findings remain.

## Status

- Completion claim: roadmap/design complete
- Completion rule: roadmap/design evidence is present and the anti-stub scan passed.
- Residual risks: the MCP server itself is not implemented; the public-vs-internal package split should be rechecked if the future MCP track changes OSF API or auth shape.
- Next phase: implement the MCP server against the documented tool inventory without widening the CLI surface.
