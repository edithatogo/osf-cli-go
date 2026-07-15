# Phase Review: Structured observability

Reviewed: 2026-07-14

## Review scope

- Compared the implementation with issue #96, `spec.md`, `plan.md`, and the
  project workflow.
- Reviewed schema stability, event IDs, redaction, file permissions, opt-in
  destinations, CLI output separation, API request events, transfer events,
  and MCP lifecycle/error events.

## Evidence

- `internal/observability/events.go` defines `osf.event.v1`, stable error
  classes, operation/request IDs, level filtering, JSONL sinks, redaction, and
  environment-controlled destinations.
- `internal/osfapi/client.go` emits low-cardinality API request events without
  retaining URLs or authorization headers.
- CLI command lifecycle, resumable transfers, and MCP lifecycle/tool errors
  emit through the same contract.
- Tests cover schema output, level filtering, secret/path redaction, error
  classes, stdout separation, API events, transfer events, and MCP errors.
- `docs/observability.md` and `docs/operations-runbook.md` document opt-in
  configuration, destinations, retention, troubleshooting, and compatibility.
- Review remediation redacts directly supplied event error messages and forces
  existing file destinations to owner-only `0600` permissions before use.

## Safety boundary

Automatic request retries are not added because retrying authenticated writes
could duplicate remote mutations. `retryCount` is part of every event and is
zero for the current non-retrying client; future safe retry policies can report
their count without changing the schema.

## Validation

- `go test ./...`
- `go test -race ./...`
- `go vet ./...`
- `go run ./tools/checkstubs`
- `go run ./tools/checkfeaturematrix`
- `go run ./tools/checkregistries`
- `go run ./tools/checkreleasecontract`
- `git diff --check`
