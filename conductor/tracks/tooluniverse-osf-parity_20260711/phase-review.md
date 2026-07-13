# Phase Review: ToolUniverse OSF Parity

## Review Date

2026-07-13

## Scope

Review of the ToolUniverse comparison, dedicated OSF preprint search API,
CLI/MCP contracts, registry metadata, and closeout documentation.

## Findings And Fixes

- The interrupted implementation was missing `osf_preprints_search` in the
  Docker registry fixture and downstream submission packet; both were added.
- Blank CLI and API search queries could otherwise be accepted; both now return
  actionable validation errors.
- Source provenance is included in table, JSON, and MCP results to match the
  audited upstream contract.

## Validation

- `go test ./internal/osfapi ./internal/cli ./internal/mcpserver`
- `go run ./tools/checkregistries`
- Full repository gates are required before archive: tests, race tests, vet,
  lint, vulnerability, anti-stub, review, feature-matrix, release, registry,
  and whitespace checks.

## Decision

The accepted OSF preprint search gap is implemented and offline-tested. The
broader ToolUniverse agent runtime is explicitly deferred as out of scope for
this OSF-focused repository. No credentials, writes, or live API assumptions
were introduced. Track is eligible for archive after the full gate run.
