# Technical Stack

## Language And Runtime

- Go 1.26.x, matching the local toolchain currently available in this workspace.
- Standard library first for the initial scaffold to avoid premature dependency and packaging decisions.
- Public module path is currently local: `osf-cli-go`. Replace this with the canonical repository module path before the first public release.

## CLI Architecture

- Entry point: `cmd/osf/main.go`.
- Command orchestration: `internal/cli`.
- Future API client package: `internal/osfapi`.
- Future auth/token package: `internal/auth`.
- Future output rendering package: `internal/output`.

## API Direction

- Target OSF API v2 at `https://api.osf.io/v2/`.
- Model OSF entities explicitly around nodes, files, users, contributors, registrations, and metadata.
- Implement pagination and relationship traversal as core client primitives, not per-command afterthoughts.
- Prefer context-aware HTTP functions and typed errors that preserve status code, endpoint, and OSF error details.

## Authentication

- Primary auth input: OSF personal access token from `OSF_TOKEN`.
- Future config support may use an OS-specific credential store, but project-local config must never contain secrets.
- Commands that can operate on public content should not require authentication.

## Output

- Default output: concise tables for humans.
- `--json` output: stable JSON with clear schemas for automation.
- Error output: short actionable messages by default, with future verbose/debug mode for HTTP detail.

## Testing

- Unit tests use Go's standard `testing` package.
- HTTP tests should use `httptest.Server` and fixture JSON instead of live OSF calls.
- Live API tests, when added, must be opt-in and skipped unless explicit environment variables are present.

## Tooling

- Format: `go fmt ./...`
- Test: `go test ./...`
- Static checks: add `go vet ./...` as the codebase grows.
- Release packaging: later use GoReleaser or a similarly conventional Go release pipeline once command behavior stabilizes.

## Constraints

- No hardcoded OSF credentials, personal tokens, or private project identifiers.
- Avoid network-dependent unit tests.
- Keep destructive API operations behind explicit command names and confirmation or force flags.
