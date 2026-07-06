# Spec: Cursor Directory Adoption

## Overview

Prepare, submit, and verify a Cursor Directory listing for the OSF MCP server,
using the current Cursor plugin/directory submission path.

## Functional Requirements

- Verify Cursor Directory submission requirements and current MCP/server listing
  state.
- Prepare Cursor-compatible MCP install metadata, docs, and repository examples.
- Use Chrome for `cursor.directory` submission or login flows; ask the user to
  log in if authentication blocks progress.
- Improve any listing completeness or quality signal toward 100%.
- Record receipts, listing URLs, validation output, and blockers.

## Acceptance Criteria

- Cursor Directory listing is submitted/verified or a precise blocker is stored.
- Cursor install docs and repo metadata are validated.
- Go, JSON, registry, anti-stub, and review checks pass.

## Out Of Scope

- Building a separate Cursor extension if a directory listing accepts MCP config.
