# Developer Guide

This document extends the repository-level [contributing guide](https://github.com/edithatogo/osf-cli-go/blob/main/CONTRIBUTING.md) with implementation conventions for CLI, API, and documentation work.

## Local Setup

```powershell
go mod tidy
```

On Windows, prefer the project script because it sets repo-local Go caches:

```powershell
.\scripts\check.ps1
```

If `gcc` is unavailable locally, `.\scripts\check.ps1 -AllowRaceSkip` skips the local race test. GitHub Actions remains the strict race-test gate.

## Normal Validation

```powershell
$env:GOTELEMETRY='off'
$env:GOCACHE=(Join-Path (Get-Location) '.gocache')
$env:GOMODCACHE=(Join-Path (Get-Location) '.gomodcache')
go fmt ./...
go test ./...
go vet ./...
go run ./tools/checkstubs
go run ./tools/checkreviews
```

Run coverage when closing a test or coverage track:

```powershell
go test ./... "-coverprofile=coverage.out"
go tool cover "-func=coverage.out"
```

## Implementation Rules

- Keep command routing and terminal presentation in `internal/cli`.
- Keep HTTP and JSON:API behavior in `internal/osfapi`.
- Keep local path safety and atomic file writes in `internal/download`.
- Keep token loading and redaction in `internal/auth`.
- Keep JSON/table formatting in `internal/output`.
- Use `context.Context` for all HTTP-facing API methods.
- Use `httptest.Server` and fixtures for unit tests; live OSF calls must stay opt-in.

## Documentation Rules

- Update `docs/commands.md` when the command surface changes.
- Update `docs/examples.md` when a common workflow changes.
- Update `docs/architecture.md` when package ownership or command routing changes.
- Add Go doc comments for exported package symbols.
- Keep destructive/write examples conservative and explicit.

## Conductor Closeout

When closing a Conductor phase:

1. Confirm every checked task is implemented and tested.
2. Run the relevant quality gates.
3. Run `$conductor-review` using the local review protocol.
4. Write `phase-review.md` in the track directory.
5. Mark the plan and `conductor/tracks.md` only after evidence exists.
