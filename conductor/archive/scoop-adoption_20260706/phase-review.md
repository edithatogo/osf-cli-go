# Scoop Adoption Review

## Result

Complete for the project-owned bucket route. Scoop Main closed PR #8293 because
the project does not meet its popularity threshold; the project-owned bucket
is the supported immediate distribution route.

## Evidence

- Bucket: https://github.com/edithatogo/scoop-osf
- Manifest: `bucket/osf-cli-go.json`
- Install command: `scoop bucket add osf https://github.com/edithatogo/scoop-osf`
- Release: `v0.3.2`
- Hashes verified against the public Windows release assets.
- Main-bucket blocker: https://github.com/ScoopInstaller/Main/pull/8293

## Review Validation

- Project bucket `main` commit `70743f258c995d8f47252c8ef5c0da3daf4834c2`
  contains `bucket/osf-cli-go.json`.
- `jq empty packaging/scoop/osf-cli-go.json`: passed.
- Manifest URLs and hashes match the public `v0.3.2` Windows release checksum
  files for x64 and arm64.
- `go test ./...`, `go test -race ./...`, `go vet ./...`, anti-stub, review,
  feature-matrix, registry, release-contract, `govulncheck`, and diff checks:
  passed.

## Archive Decision

The project-owned Scoop bucket is published and satisfies the local adoption
contract. The Main-bucket popularity threshold remains an external blocker,
but is out of scope for the supported project-owned route; the track is
archive-eligible.
