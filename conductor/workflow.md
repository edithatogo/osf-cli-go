# Project Workflow

## Guiding Principles

1. The plan is the source of truth. Track scoped work in the relevant `conductor/tracks/*/plan.md`.
2. Tech stack changes must be documented in `conductor/tech-stack.md` before implementation.
3. Prefer test-driven development for CLI behavior, API parsing, auth handling, and file conflict logic.
4. Keep OSF API calls behind testable interfaces and use fixtures for routine validation.
5. Favor non-interactive commands that work in CI and on Windows PowerShell.

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
```

### Before Committing

```powershell
go fmt ./...
go test ./...
go vet ./...
```

## Task Workflow

1. Select the next unchecked task from the active track plan.
2. Mark it `[~]` before editing code.
3. Add or update tests that define the behavior.
4. Run the targeted tests and confirm the expected failure when practical.
5. Implement the smallest coherent change.
6. Run `go fmt ./...` and `go test ./...`.
7. Update documentation or Conductor files when behavior or scope changes.
8. Mark the task `[x]` with the relevant commit SHA after committing.

## Quality Gates

- All Go files are formatted with `gofmt`.
- Unit tests pass without live network access.
- Public functions have useful GoDoc when exported.
- CLI errors include a clear next action.
- Secret material is never printed, committed, or written to project-local config.
- Commands that write to OSF are explicit and have conservative defaults.

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
