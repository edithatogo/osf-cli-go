# Release Candidate Provenance Evidence

Reviewed: 2026-07-14

## Local candidate

- Candidate version: `1.0.0-rc.1`
- Source commit: `cb988fe25eb61a88a5af4b0d39447413d8273c1a`
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

## Unresolved release gates

- No immutable `v1.0.0-rc.1` tag or GitHub release was created.
- OCI image build and digest verification were not run because Docker is not
  available in this environment.
- Cosign signature generation and verification were not run; keyless signing
  requires the GitHub Actions OIDC environment.
- GitHub build-provenance attestations were not generated for this local build.
- Windows/macOS/Linux package installation was not independently exercised on
  each supported runner.

The candidate is locally buildable and checksum-verifiable, but this evidence
does not satisfy #98 until the hosted release workflow produces and verifies
the signed OCI, SBOM, provenance, and package artifacts from an immutable
release-candidate tag.
