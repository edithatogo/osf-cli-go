# Effective Go Style Guide Summary

## Formatting

- Run `go fmt ./...` before committing.
- Let `gofmt` decide indentation and wrapping.

## Naming

- Use `MixedCaps` or `mixedCaps`; do not use underscores in Go identifiers.
- Package names should be short, lowercase, and singular where practical.
- Exported names must have comments that begin with the exported name.

## Errors

- Return errors explicitly.
- Do not discard errors with `_` unless the reason is obvious and local.
- Avoid `panic` outside unrecoverable programmer errors.

## Interfaces

- Keep interfaces small and define them where they are consumed.
- Use `context.Context` for HTTP/API operations.

## Tests

- Prefer table-driven tests for command parsing and API response handling.
- Use `httptest.Server` for API client behavior.
- Keep live OSF calls out of unit tests.
