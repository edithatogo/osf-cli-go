# Provider boundary fuzzing

The seven Go fuzz targets exercise local URL, JSON, XML and token conversion.
They never call OSF or Zenodo. Seed inputs are synthetic; private provider
responses and credentials must not be added to the corpus.

Ordinary `go test ./...` replays seed inputs and any checked-in regressions.
The separate **Bounded provider fuzzing** workflow mutates each target for
30 seconds on relevant pull requests and 300 seconds on Tuesday/manual runs.
At most two jobs and two fuzz workers per job execute concurrently. Each
campaign has a seven-minute process timeout and a twelve-minute job timeout.
`GOMEMLIMIT=768MiB` is a Go soft memory target, not an OS memory quota.
Individual byte/string inputs are capped at 64 KiB; token fields have smaller
limits. Larger-input performance and native sanitizer qualification are separate.

## Run and reproduce

Use the Go toolchain selected by `go.mod`:

```sh
go test ./internal/zenodoapi -run='^FuzzRecordNativePreservation$' -count=1
go test ./internal/zenodoapi -run='^$' -fuzz='^FuzzRecordNativePreservation$' -fuzztime=30s -parallel=2 -timeout=2m
```

On failure Go writes the minimized input under
`internal/<package>/testdata/fuzz/<target>/`. CI retains those files and source,
compiler and campaign logs for 14 days. Replay an artifact with the exact source
commit and compiler recorded alongside it:

```sh
go test ./internal/zenodoapi -run='FuzzRecordNativePreservation/<input-hash>' -count=1
```

After fixing a defect, check the minimized synthetic input into the owning
package's `testdata/fuzz` directory. Coverage-growth counts indicate explored
paths; a clean bounded campaign does not establish complete coverage.
