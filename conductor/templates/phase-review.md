# Phase Review

## Track

- Track:
- Phase:
- Date:

## Implemented Behavior

- 

## Anti-Stub Evidence

- `go run ./tools/checkstubs` result:
- Production markers found:
- Ignored paths verified:
- Self-scan exclusion verified:
- Validation evidence link or location:

## Validation Commands

```powershell
go fmt ./...
go test ./...
go test -race ./...
go vet ./...
go run ./tools/checkstubs
go test ./... "-coverprofile=coverage.out"
go tool cover "-func=coverage.out"
```

## Conductor Review

- Review command: `$conductor-review`
- Blocking findings:
- Fixes applied:
- Re-review result:

## Status

- Completion claim: scaffolded | offline-tested | integration-ready | live-validated
- Completion rule: do not select a claim unless Anti-Stub Evidence is filled and the current branch passed `go run ./tools/checkstubs`.
- Residual risks:
- Next phase:
