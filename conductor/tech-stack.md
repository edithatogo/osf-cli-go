# Technical Stack

## Language And Runtime

- Go 1.26.x, matching the local toolchain currently available in this workspace.
- Cobra is the approved CLI router for nested commands, help output, command-specific flags, and future completions.
- Keep dependencies deliberate and documented. Do not add a dependency unless it supports the CLI contract, OSF API integration, testing, release automation, or security/quality gates.
- Public module path is currently local: `osf-cli-go`. Replace this with the canonical repository module path before the first public release.

## CLI Architecture

- Entry point: `cmd/osf/main.go`.
- Command orchestration: `internal/cli`.
- Future Cobra command tree root: `internal/cli`.
- Future API client package: `internal/osfapi`.
- Future auth/token package: `internal/auth`.
- Future output rendering package: `internal/output`.
- Future download safety package: `internal/download`.
- Provider-qualified identity, metadata, lifecycle, and capability contracts:
  `internal/repository`; this package does not define a generic network client.
- Read-only Zenodo published-record REST adapter: `internal/zenodoapi`; writes,
  depositions, and publication remain outside this package.
- Public Zenodo OAI-PMH adapter: `internal/zenodooai`; XML metadata, sets,
  schemas, protocol errors, and opaque continuation remain separate from REST.
- Authenticated Zenodo draft transfer adapter: `internal/zenodotransfer`;
  writes are sandbox-only, whole-file uploads never claim partial resume, and
  verified byte-range downloads reuse `internal/download` checkpoints.
- Provider-scoped Zenodo CLI commands consume the concrete REST and OAI-PMH
  clients; write-shaped commands consult `repository.ZenodoContract` only.
- Provider-scoped MCP tools consume the same concrete clients and shared
  `internal/zenodoid` parser; registry inventories are checked against the
  executable read-only tool list.
- Future reusable core packages should begin under `internal/`. Promote public packages only after the CLI behavior stabilizes and an MCP server track proves the package boundary.

## API Direction

- Target OSF API v2 at `https://api.osf.io/v2/`.
- Pin future Zenodo support to the dated, digested capability contract in
  `docs/zenodo-api-source.json`; treat REST and OAI-PMH as separate adapters.
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
- Structured observability: the standard-library `internal/observability` package emits opt-in `osf.event.v1` JSONL events with operation/request IDs, redaction, level filtering, and classified errors.
- Event destinations are local-only and owner-readable; stdout is never used for structured events.

## Testing

- Unit tests use Go's standard `testing` package.
- HTTP tests should use `httptest.Server` and fixture JSON instead of live OSF calls.
- Live API tests, when added, must be opt-in and skipped unless explicit environment variables are present.
- Concrete repository descriptors must pass the reusable
  `internal/repository/conformancetest` suite before CLI or MCP exposure.
- Zenodo REST tests use dated synthetic fixtures, bounded `httptest` transports,
  race tests, and dedicated parser/pagination fuzz targets without live network.
- Zenodo draft transfers use offline failure injection plus the opt-in
  `tools/zenodosandboxvalidation` disposable live harness; live credentials are
  scoped, ephemeral, and never required by routine tests.
- Zenodo OAI-PMH tests use synthetic XML fixtures, deterministic expiry clocks,
  strict parsing, and a dedicated parser fuzz target without live harvesting.

## Tooling

- Format: `go fmt ./...`
- Test: `go test ./...`
- Race tests: `go test -race ./...`
- Static checks: `go vet ./...`
- Lint: `golangci-lint run`
- Vulnerability scan: `govulncheck ./...`
- Anti-stub scan: `go run ./tools/checkstubs`
- Zenodo API contract: `go run ./tools/checkzenodoapi` offline in normal CI;
  `-online` only in the scheduled/manual credential-free drift workflow.
- Release packaging: later use GoReleaser or a similarly conventional Go release pipeline once command behavior stabilizes.
- Dependency updates: Renovate for Go modules and GitHub Actions, without automerge initially.

## Constraints

- No hardcoded OSF credentials, personal tokens, or private project identifiers.
- Avoid network-dependent unit tests.
- Keep destructive API operations behind explicit command names and confirmation or force flags.
