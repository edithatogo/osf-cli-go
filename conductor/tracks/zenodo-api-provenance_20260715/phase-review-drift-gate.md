# Phase Review: Drift gate

## Track

- Track: `zenodo-api-provenance_20260715`
- Phase: Drift gate
- Date: 2026-07-15

## Implemented behavior

- Added a strict JSON decoder and deterministic SHA-256 digest over the reviewed capability snapshot.
- Validated official source hosts, source ordering, retrieval dates, evidence markers, authentication policy, limits, capability ordering, required operations, and write-risk authentication.
- Added actionable online drift diagnostics without credentials or repository API operations.
- Added an offline pull-request gate and a credential-free scheduled/manual online workflow.
- Added Make, PowerShell, development, and MkDocs integration.

## Review findings and fixes

- Corrected the official repository-policy access marker to its exact source ordering.
- Confirmed the online probe validates records, depositions, scopes, bucket limits, OAI-PMH, rate-limit headers, terms, and policy evidence.
- No blocking findings remain.

## Validation

- `go test ./...`
- `go test -race ./...`
- `go vet ./...`
- `golangci-lint run` (`0 issues`)
- `govulncheck ./...` (`No vulnerabilities found`)
- `go run ./tools/checkstubs`
- `go run ./tools/checkreviews`
- `go run ./tools/checkregistries`
- `go run ./tools/checkfeaturematrix`
- `go run ./tools/checkreleasecontract`
- `go run ./tools/checkzenodoapi`
- `go run ./tools/checkzenodoapi -online`
- `mkdocs build --strict`
- `git diff --check`

## Status

- Completion claim: live-validated documentation drift gate.
- Residual risk: scheduled online validation depends on official documentation availability and remains subject to hosted GitHub Actions execution.
- Next track: repository provider contract, issue #103.
