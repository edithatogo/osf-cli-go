# Phase Review

## Track

- Track: #98 Release-candidate provenance
- Phase: Candidate build, independent verification, and launch decision
- Date: 2026-07-15

## Implemented Behavior

- Release candidate workflows now publish versioned binaries, checksums,
  MCPB bundles, plugin archives, SBOM/provenance metadata, and an immutable
  keyless-signed OCI image for every `v*` tag.
- Candidate package publication is idempotent and derives package versions
  from the tag, preventing stale-version assets.
- Native runner version checks cover runnable Linux, macOS, and Windows
  targets; cross-platform checksum manifests are independently verified.

## Anti-Stub Evidence

- `go run ./tools/checkstubs` result: pass.
- Production markers found: none outside approved test/documentation paths.
- Ignored paths verified: local candidate output is under ignored
  `dist/release-candidate/`; no generated release artifact was committed.
- Self-scan exclusion verified: the repository anti-stub checker passed.
- Validation evidence: `provenance-evidence.md` and hosted workflow runs
  29339729495, 29339729375, 29339729457, 29339729423, and 29339729662.

## Validation Commands

```text
go test ./...
go test -race ./...
go vet ./...
go run ./tools/checkstubs
go run ./tools/checkfeaturematrix
go run ./tools/checkregistries
go run ./tools/checkreleasecontract
git diff --check
```

All commands passed. The six published checksum manifests passed with
`shasum -a 256 -c`, isolated-config Cosign verification passed for the
immutable digest `sha256:1d6655b24d832e47782123c41743067562fbefaac5b8a85f0837b7f537d7aa6c`,
and public GHCR OCI inspection verified the SPDX and SLSA provenance layers
and their BuildKit source revision.

## Conductor Review

- Review command: `$conductor-review #98`
- Blocking findings: the initial review identified missing registry-level
  attestation evidence; this was fixed and rechecked.
- Fixes applied: corrected MCPB integrity metadata, idempotent release
  uploads, tag-derived package versions, immutable digest signing, and native
  version checks; synchronized evidence with the final hosted runs.
- Re-review result: clean local gates, clean hosted candidate verification,
  and auditable OCI attestation evidence.

## Status

- Completion claim: live-validated
- Completion rule: anti-stub evidence is complete and the current branch
  passes `go run ./tools/checkstubs`.
- Residual risks: live OSF behavior remains tracked separately by #97;
  compatibility freeze and broader external marketplace gates remain tracked
  by #99 and the launch review.
- Next phase: archive this completed track and continue #97/#99.
