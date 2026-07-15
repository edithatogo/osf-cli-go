# Structured observability

GitHub issue: [#96](https://github.com/edithatogo/osf-cli-go/issues/96)

## Objective

Provide a stable, redacted operational event contract for CLI, API, transfer,
and MCP execution.

## Requirements

- Define JSON event fields for operation ID, request ID, duration, retry count,
  outcome, endpoint class, and classified errors.
- Keep human-readable stderr and machine-readable output separate.
- Redact tokens, passwords, authorization headers, secrets, and sensitive paths.
- Document levels, destinations, retention, troubleshooting, and compatibility.
- Test redaction, cancellation, retries, MCP errors, and schema evolution.

## Out of scope

External telemetry collection, user tracking, and mandatory hosted observability.
