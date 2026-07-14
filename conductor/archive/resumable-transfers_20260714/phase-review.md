# Phase Review: Resumable transfers

Reviewed: 2026-07-14

## Review scope

- Compared the implementation with issue #95, `spec.md`, and the phase plan.
- Reviewed checkpoint identity, stale-state invalidation, range handling,
  checksum verification, atomic finalization, conflict policies, manifests,
  and provider-supported upload recovery.
- Ran targeted, race, static-analysis, stub, feature-matrix, registry, release,
  and whitespace gates.

## Findings and fixes

- Upload acknowledgements now compare against the previous offset, so valid
  multi-chunk progress is accepted while zero-progress sessions are rejected.
- Skipped destinations no longer report a checkpoint path that was not used.
- Upload checkpoint fields store fingerprints rather than raw source identities.
- Folder-tree downloads now have an interruption/restart regression test.

## Residual boundaries

- WaterButler remains a one-shot upload provider; the reusable upload API only
  resumes when a provider explicitly acknowledges chunk offsets.
- Live authenticated OSF validation remains tracked by issue #97 and is not
  represented as local implementation evidence.

## Validation

- `go test ./...`
- `go test -race ./internal/download ./internal/osfapi ./internal/cli`
- `go vet ./...`
- `go run ./tools/checkstubs`
- `go run ./tools/checkfeaturematrix`
- `go run ./tools/checkregistries`
- `go run ./tools/checkreleasecontract`
- `git diff --check`
