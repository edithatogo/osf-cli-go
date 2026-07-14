# Phase Review

## Track

- Track: `zenodo-publication-state_20260715`
- Phase: Publication workflows
- Date: 2026-07-15

## Implemented Behavior

- Added a sandbox-only executor for automatic DOI-reservation verification, metadata application, publication, new-version creation, and draft discard.
- Required exact dry-run confirmation for publication/discard and explicit client-declared token scopes for each network action.
- Rejected production/cross-origin targets, bounded responses, preserved redacted typed errors, disabled automatic retries for non-idempotent actions, and documented recovery inspection.
- Added an opt-in harness that created and published sandbox record `565256`, verified the public record and file, created version draft `565257`, and discarded only the unpublished version draft.
- Used a one-use token with `deposit:write` and `deposit:actions`, excluded `user:email`, removed its local copy, and verified revocation in the sandbox account.

## Anti-Stub Evidence

- `go run ./tools/checkstubs` result: pass
- Production markers found: none
- Ignored paths verified: existing scanner policy unchanged
- Self-scan exclusion verified: pass
- Validation evidence link or location: `docs/zenodo-publication-validation-evidence.md`

## Validation Commands

```powershell
go fmt ./...
go test ./...
go test -race ./...
go vet ./...
golangci-lint run
go run ./tools/checkstubs
go run ./tools/checkfeaturematrix
go run ./tools/checkzenodoapi
```

All commands passed after the review fixes.

## Conductor Review

- Review command: `$conductor-review` publication-workflows phase
- Blocking findings: legacy DOI reservation payload rejected by the live sandbox; ambiguous post-action failure recovery; ignored cleanup errors
- Fixes applied: verify the automatically reserved DOI, add typed partial-publication and inspect-before-retry errors, and surface draft/workspace cleanup failures
- Re-review result: no blocking findings (`2e2b361`)

## Status

- Completion claim: live-validated
- Residual risks: the published sandbox record is intentionally irreversible; production publication remains rejected and no stable CLI/MCP write is advertised
- Next phase: track closeout and cross-provider provenance transfer (#110)
