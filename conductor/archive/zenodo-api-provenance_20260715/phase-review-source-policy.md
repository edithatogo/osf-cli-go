# Phase Review: Source policy

## Track

- Track: `zenodo-api-provenance_20260715`
- Phase: Source policy
- Date: 2026-07-15

## Implemented behavior

- Recorded official developer, sandbox, terms, and policy sources with retrieval dates and structural evidence markers.
- Adopted a documentation-date policy because the documented depositions REST surface has no published semantic API version or pinned OpenAPI schema.
- Separated production REST, sandbox REST, and OAI-PMH boundaries and rejected access tokens in query parameters.
- Recorded scopes, limits, lifecycle risks, file constraints, and uploader policy inputs without claiming implementation support.

## Validation

- `go test ./tools/checkzenodoapi`
- `go run ./tools/checkzenodoapi`
- `go run ./tools/checkzenodoapi -online`
- `go vet ./tools/checkzenodoapi`
- `git diff --check`

## Conductor review

- Blocking findings: none.
- Fixes applied: corrected the exact access-policy evidence marker ordering to match the official source.
- Re-review result: offline and online source validation pass.

## Status

- Completion claim: live-validated documentation provenance.
- Residual risk: official documentation can change without a semantic version; the next phase supplies scheduled structural drift detection.
- Next phase: deterministic snapshot and CI drift gate.
