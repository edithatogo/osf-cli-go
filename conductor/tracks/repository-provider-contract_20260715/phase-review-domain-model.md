# Phase Review: Domain model

## Track

- Track: `repository-provider-contract_20260715`
- Phase: Domain model
- Date: 2026-07-15

## Implemented behavior

- Added provider-qualified, provider/kind-validated, reversible identities.
- Added a lossless native metadata envelope with defensive copies, media-type-aware JSON validation, and JSON round trips.
- Added common plus native lifecycle state, explicit permissions, links, checksums, and separate native/DOI/concept-DOI version identity.
- Added a complete capability vocabulary with concrete OSF and Zenodo supported, partial, and unsupported decisions.
- Added typed partial/unsupported errors without introducing a generic network client interface.
- Documented semantic non-equivalences and preserved the frozen OSF public surfaces.

## Review findings and fixes

- Rejected impossible provider/resource-kind combinations instead of accepting structurally valid but semantically false identities.
- Exposed native metadata in lossless JSON rather than silently omitting it.
- Applied JSON validation to structured media types with parameters.
- Made permission validation diagnostics deterministic.
- No blocking findings remain.

## Validation

- `go test ./internal/repository/...`
- `go vet ./internal/repository/...`
- `go test ./internal/cli ./internal/mcpserver -run 'CompatibilityFixture|RootContractMatchesCompatibilityFixture'`
- `git diff --check`

## Status

- Completion claim: offline-tested domain contract.
- Residual risk: no Zenodo network adapter consumes the contract yet; that belongs to issue #104.
- Next phase: reusable conformance fixtures and full compatibility validation.
