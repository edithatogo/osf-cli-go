# Phase Review: Client primitives

## Track

- Track: `zenodo-readonly-client_20260715`
- Phase: Client primitives
- Date: 2026-07-15

## Implemented behavior

- Added production, sandbox, and localhost-test configuration with optional bearer-header authentication.
- Added bounded response, pagination, retry, redirect, and concurrency policies for GET-only requests.
- Added typed redacted API errors, rate-limit state, cancellation, retry-after handling, and provider-tagged observability.
- Added current and legacy Zenodo file-shape decoding and lossless raw-record conversion to the provider envelope.
- Added synthetic provenance-linked fixtures with no account or live user data.

## Review findings and fixes

- Restricted configured hosts to Zenodo production, sandbox, or localhost tests.
- Rejected cross-origin redirects before bearer headers can be forwarded.
- Added explicit concurrency bounds and semantic validation for missing search hits.
- Added parser and pagination fuzz targets.
- No blocking findings remain.

## Validation

- `go test ./internal/zenodoapi`
- `go test -race ./internal/zenodoapi`
- `go vet ./internal/zenodoapi`
- Targeted coverage: 81.0% before the read-operation boundary expansion.
- `git diff --check`

## Status

- Completion claim: offline-tested client primitives.
- Residual risk: boundary coverage must increase before track completion; live public API returned a non-JSON intermediary response and is not used as fixture evidence.
- Next phase: read-operation boundary coverage and full repository compatibility review.
