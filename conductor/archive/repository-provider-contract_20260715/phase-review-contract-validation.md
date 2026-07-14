# Phase Review: Contract validation

## Track

- Track: `repository-provider-contract_20260715`
- Phase: Contract validation
- Date: 2026-07-15

## Implemented behavior

- Added a reusable conformance suite for vocabulary completeness, ordering, resolution, model version, and JSON round trips.
- Added fixture cases covering supported, partial, and unsupported behavior across concrete OSF and Zenodo descriptors.
- Added negative tests for malformed catalogs, missing reasons, unknown capabilities, partial errors, invalid identities, metadata mutation, and record-envelope round trips.
- Added an explicit provider-conformance CI gate and documentation navigation.
- Kept all existing CLI, MCP, and OSF API behavior unchanged.

## Review findings and fixes

- Renamed the typed error from an unsupported-only name to `CapabilitySupportError` because it also represents partial support.
- Added parameterized JSON media-type validation and binary metadata round-trip coverage.
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
- Targeted provider/conformance aggregate statement coverage after Codecov remediation: 92.5%.
- `go test ./internal/cli ./internal/mcpserver -run 'CompatibilityFixture|RootContractMatchesCompatibilityFixture'`
- `mkdocs build --strict`
- `git diff --check`

## Status

- Completion claim: offline-tested provider contract.
- Residual risk: live Zenodo behavior remains intentionally deferred to adapter and sandbox tracks.
- Next tracks: Zenodo REST client #104 and OAI-PMH adapter #107.
