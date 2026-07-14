# Release Candidate Provenance Evidence

Reviewed: 2026-07-14

## Local candidate

- Candidate version: `1.0.0-rc.1`
- Source commit: `026f3c8fe147205986952b2f08c7a585039edd92`
- Build output: ignored local directory `dist/release-candidate/`
- Build command: cross-compiled `cmd/osf` and `cmd/osf-mcp` for Linux,
  macOS, and Windows amd64/arm64 with `CGO_ENABLED=0` and embedded version,
  commit, and build date metadata.
- Checksum command: `go run ./tools/checksums dist/release-candidate/binaries dist/release-candidate/SHA256SUMS-binaries.txt`
- Checksum verification: all 12 binary entries passed `shasum -a 256 -c`.
- Clean-directory verification: copied the macOS arm64 CLI binary to a
  temporary directory without the source checkout; `--version` and `--help`
  passed.

## Package evidence

- SBOM: `dist/release-candidate/osf-cli-go-sbom.spdx.json`, SPDX parsed with
  `jq`; SHA-256 `32bbff5bb5d1d70c3956e7a24aea6873988b6e9a1fc58439d4934b5e9c61534f`.
- MCPB: `osf-cli-go-1.0.0-rc.1-osx-arm64.mcpb`; SHA-256
  `2860de7891f204f3dd9315c3ddf2045103e4be4ec87f9a8729014401ffab79b8`;
  required manifest, MCP configuration, README, and server binary present.
- Plugin archives: five macOS arm64 archives were built and inspected for
  their required manifests, MCP configuration, documentation, skills, and
  server binary. Their checksums are recorded in the local build output.

## Hosted candidate verification

- Immutable tag: `v1.0.0-rc.1` at the source commit above.
- Release Artifacts workflow: run
  [29339418104](https://github.com/edithatogo/osf-cli-go/actions/runs/29339418104)
  passed all six binary matrix jobs, native Linux/macOS/Windows version checks,
  checksum generation, and release publication.
- MCPB workflow: run
  [29339418373](https://github.com/edithatogo/osf-cli-go/actions/runs/29339418373)
  passed Linux, macOS, and Windows bundle builds and release upload.
- Plugin workflow: run
  [29339418952](https://github.com/edithatogo/osf-cli-go/actions/runs/29339418952)
  passed Linux, macOS, and Windows plugin builds and release upload.
- Release Security workflow: run
  [29339418109](https://github.com/edithatogo/osf-cli-go/actions/runs/29339418109)
  published the OCI image with BuildKit SBOM/provenance metadata and signed
  immutable digest `sha256:7a34e65fab61c969ba144b05a0086a5f931da0d48ed3c48cf43034536bf08407`.
- Independent Cosign verification passed against that digest with the
  expected GitHub Actions OIDC identity.
- The final GitHub release contains 36 clean RC assets: 12 binaries, 6
  checksum files, 15 plugin archives, and 3 MCPB packages; no stale `0.3.2`
  assets remain.

Residual risk: this macOS environment cannot execute Windows/Linux binaries,
so cross-platform verification relies on the native hosted runner checks and
published checksum verification rather than local execution of every target.
