# Cursor Directory Submission Evidence

## Audit

- Date: 2026-07-13
- Public submission route: <https://cursor.directory/plugins/new>
- Public directory: <https://cursor.directory/>
- Provider source repository: <https://github.com/cursor/community-plugins>
- Plugin standard: <https://open-plugins.com/plugin-builders/specification>

## Repository Preparation

- `.cursor/mcp.json` provides the Cursor project configuration.
- `.mcp.json` provides the vendor-neutral Open Plugins MCP component.
- `.plugin/plugin.json` provides versioned vendor-neutral plugin metadata.
- `integrations/README.md` documents Cursor installation and the submission route.
- The MCP server remains stdio-based and read-only; credentials are environment-backed and are not stored in the repository.

## Current External State

- The public submission page was reachable.
- The page redirected to `https://cursor.directory/login?next=/plugins/new`.
- The authenticated Chrome session did not have an active Cursor Directory login.
- No submission, listing URL, score, or provider scan result is claimed yet.

## Required Follow-up

Sign in with GitHub or Google in the authenticated Chrome session, return to
the submission route, submit `https://github.com/edithatogo/osf-cli-go`, and
record the resulting receipt, listing URL, scan status, and any completeness
score here.
