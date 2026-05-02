# Project Workflow

## Guiding Principles

1. The plan is the source of truth. Track scoped work in the relevant `conductor/tracks/*/plan.md`.
2. Tech stack changes must be documented in `conductor/tech-stack.md` before implementation.
3. Prefer test-driven development for CLI behavior, API parsing, auth handling, and file conflict logic.
4. Keep OSF API calls behind testable interfaces and use fixtures for routine validation.
5. Favor non-interactive commands that work in CI and on Windows PowerShell.
6. No stubs may be marked complete. A task is not done if production code still contains placeholder behavior, dummy responses, fake OSF data outside tests, `panic("TODO")`, `not implemented`, or command handlers that only pretend to work.
7. Status language must be precise: use "scaffolded", "offline-tested", "integration-ready", or "live-validated" instead of "finished" when live OSF behavior has not been exercised.

## Development Commands

### Setup

```powershell
go mod tidy
```

### Daily Development

```powershell
go test ./...
go run ./cmd/osf --help
go run ./cmd/osf --version
go run ./tools/checkstubs
go test ./... "-coverprofile=coverage.out"
go tool cover "-func=coverage.out"
```

### Before Committing

```powershell
go fmt ./...
go test ./...
go test -race ./...
go vet ./...
go run ./tools/checkstubs
```

On Windows, `scripts/check.ps1` sets repo-local Go build and module caches to avoid user-profile cache permission issues. Local race tests require a C compiler. The script fails when `gcc` is missing unless `-AllowRaceSkip` is explicitly supplied for local development. GitHub Actions remains the strict race-test gate.

## Task Workflow

1. Select the next unchecked task from the active track plan.
2. Mark it `[~]` before editing code.
3. Add or update tests that define the behavior.
4. Run the targeted tests and confirm the expected failure when practical.
5. Implement the smallest coherent change.
6. Run `go fmt ./...` and `go test ./...`.
7. Update documentation or Conductor files when behavior or scope changes.
8. Run the anti-stub scan before marking anything complete.
9. Mark the task `[x]` with the relevant commit SHA after committing.

## Phase Exit Review Protocol

Every phase must finish with a review-fix-continue loop:

1. Run the required automated quality gates for the phase.
2. Run `$conductor-review` against the completed phase or current track changes.
3. Automatically apply safe fixes from the review: missing tests, incorrect CLI behavior, broken docs examples, lint/vet/format failures, missing error handling, unsafe auth/download behavior, or stubbed implementation.
4. Re-run the same quality gates after fixes.
5. Re-run `$conductor-review` until there are no blocking findings.
6. Write a phase review artifact in the track directory using `conductor/templates/phase-review.md`.
7. Mark the phase complete only after the clean review.
8. Continue to the next phase automatically unless blocked by secrets, external credentials, destructive actions, live OSF writes, dependency/license choices outside this stack, or product-scope changes.

## Quality Gates

- All Go files are formatted with `gofmt`.
- Unit tests pass without live network access.
- Race tests pass for shared packages and command plumbing.
- `go vet ./...` passes.
- `golangci-lint run` passes in CI.
- `govulncheck ./...` passes in CI.
- Production code passes the anti-stub scanner.
- Public functions have useful GoDoc when exported.
- CLI errors include a clear next action.
- Secret material is never printed, committed, or written to project-local config.
- Commands that write to OSF are explicit and have conservative defaults.
- Every completed command task has at least one runnable invocation and expected output in tests or docs.

## Commit Guidelines

Use conventional commit style:

```text
feat(cli): add project listing command
fix(auth): avoid printing token values
test(api): cover OSF pagination parsing
docs(conductor): update MVP scope
```

## Live API Testing

Live OSF integration tests must be opt-in. Use environment variables such as `OSF_TOKEN` and `OSF_TEST_PROJECT` only when explicitly running integration checks. Unit tests should continue to pass offline.
