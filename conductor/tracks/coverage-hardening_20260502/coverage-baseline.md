# Coverage Baseline

Date: 2026-05-16

## Baseline Source

- Current command: `go test ./... "-coverprofile=coverage.out"` followed by `go tool cover "-func=coverage.out"`.
- Current total coverage: 75.0% of statements.
- Previous numeric baseline: unavailable. The checked-in `coverage-before` artifact contains only `mode: set`, so it cannot be used as a numeric before value.

## Current Package Coverage

- `internal/auth`: 90.9%
- `internal/cli`: 80.4%
- `internal/download`: 87.1%
- `internal/osfapi`: 87.9%
- `internal/output`: 90.9%
- `tools/checkreviews`: 16.3%
- `tools/checkstubs`: 46.8%
- `tools/livevalidation`: 39.5%
- `cmd/osf`: 0.0%

## Targeted Risk Areas Covered

- CLI usage and argument errors, including invalid output modes, conflicting flags, missing arguments, invalid download options, and write-command confirmation behavior.
- Auth redaction and missing-token paths.
- OSF API error parsing, malformed JSON, fallback error bodies, endpoint resolution, and WaterButler error responses.
- Download conflict, skip, overwrite, traversal, symlink escape, open failure, and folder-manifest failure paths.
- Output helper JSON/table behavior and writer error propagation.

## Remaining Low-Value Or Deliberate Gaps

- `cmd/osf/main.go` remains at 0.0%; it is a thin process entrypoint over `internal/cli.Run`.
- Tool `main` functions remain low coverage because their behavior is exercised through command-level invocation and direct helper tests where useful.
- Live validation execution paths remain low coverage by design; live OSF calls are opt-in and excluded from unit tests.
- Default client wrapper methods are partly uncovered because most command tests use fake clients and API package tests cover HTTP behavior directly.
