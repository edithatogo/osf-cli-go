# Spec: Official MCP/GitHub Registry Adoption

## Overview

Keep the OSF MCP server discoverable and high-quality in the Official MCP
Registry and GitHub MCP Registry ecosystem. This track treats the official
registry as the primary source of truth for downstream MCP discovery.

## Functional Requirements

- Verify `server.json`, OCI package metadata, GHCR image references, and
  GitHub workflow evidence are current.
- Validate public registry visibility using registry APIs/pages and repo-local
  checks.
- Improve metadata, README installation snippets, category/keyword coverage,
  and auth documentation where validation or registry display exposes gaps.
- Trigger safe non-interactive publication/update paths where available.
- Use Chrome only for browser-only verification or account/auth flows; if login
  fails or requires user interaction, stop and ask the user to log in.

## Acceptance Criteria

- `go run ./tools/checkregistries`, JSON validation, Go tests, vet, anti-stub,
  and review checks pass.
- Official registry page/API shows the current package/version or a precise
  blocker is recorded.
- Any score/quality hints available from the registry are iteratively improved
  as close to 100% as the repo and target allow.
- Receipts, workflow URLs, API responses, screenshots, or blockers are stored
  under the track.

## Out Of Scope

- Publishing credentials or secrets.
- Building a new hosted Streamable HTTP service unless the registry requires it
  and a separate implementation track is created.
