# Reproducibility Protocol

## Frozen input

- Software release: `v0.3.1`
- Repository: <https://github.com/edithatogo/osf-cli-go>
- Go version: the version declared in `go.mod`
- Network-free baseline: repository fixtures and local package metadata

## Required local commands

```text
go run ./tools/checkreleasecontract
go test ./...
go test -race ./...
go vet ./...
go run ./tools/checkregistries
go run ./tools/checkstubs
go run ./tools/checkreviews
git diff --check
```

## Optional live campaign

Live OSF checks require `OSF_TOKEN` or the documented fallback credentials and
a disposable OSF project. They must remain opt-in and record redacted results,
HTTP outcomes, checksums, cleanup decisions, and the exact software revision.
No credentials belong in this repository or manuscript artifacts.

## Comparative evaluation

The comparative evaluation is a source-backed capability matrix, not an
uncontrolled performance claim. Each reference tool is recorded with its
repository, license, supported OSF entities, authentication model, transfer
semantics, automation surface, tests, release evidence, and maintenance state.
Unimplemented parity tracks remain explicit rather than being presented as
completed results.
