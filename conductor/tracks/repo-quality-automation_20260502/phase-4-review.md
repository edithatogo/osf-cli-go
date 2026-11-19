# Phase 4 review: bounded provider fuzzing

Date: 2026-11-20. Base: `61627dbbc358b99292cf1b1d7a291aa316361c3c`.
Status: local verification passed; hosted checks and default-branch admission pending.

## Scope and behavior

The repository already had five fuzz targets in ordinary CI. This extension
keeps those targets, adds native JSON evidence preservation and resumption-token
properties, and moves the mutation commands into one budgeted workflow. Ordinary
tests continue to replay seed cases. Production code, provider calls, credentials
and the module/toolchain manifests are unchanged.

PR campaigns run for 30 seconds per target; weekly/manual campaigns run for 300
seconds. Two jobs and two workers per job may run concurrently. The Go memory
limit is a soft target, not a process quota. Inputs are bounded before parsing.
Failing inputs and source/toolchain logs are retained for 14 days.

## Executed local evidence

Compiler: official Go `go1.26.11 linux/amd64`, selected explicitly from the local
installation. This does not change the repository's toolchain pin; PR #152 owns
that separate update.

- `go test ./...`: passed.
- `go test -race ./...`: passed with the race detector enabled.
- `go vet ./...`: passed.
- `go run ./tools/checkstubs`: passed.
- All seven targets passed 15-second mutation campaigns with `GOMAXPROCS=2`,
  `GOMEMLIMIT=768MiB`, `-parallel=2`, `-fuzzminimizetime=15s`, and `-timeout=2m`.
  The commands used `-run='^$'` and an exact `-fuzz='^<target>$'` selector.

| Package | Target | Executions reported |
| --- | --- | ---: |
| osfapi | FuzzResolveReference | 207769 |
| zenodoapi | FuzzResolvePagination | 273452 |
| zenodoapi | FuzzDecodeSearchPage | 149750 |
| zenodoapi | FuzzDecodeFiles | 168022 |
| zenodoapi | FuzzRecordNativePreservation | 141784 |
| zenodooai | FuzzDecodeEnvelope | 114046 |
| zenodooai | FuzzResumptionToken | 196381 |

## Self-review and limits

The new JSON oracle independently checks retained bytes and alias isolation;
the XML/token properties check namespace, determinism, opaque data and expiry
semantics. The review found no production changes or live provider interaction.
No minimized failure was produced. This is a bounded local result, not complete
input-space coverage, hosted CI acceptance or a release-readiness claim.
