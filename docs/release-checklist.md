# Release Checklist

Use this checklist before tagging a release or publishing binaries.

## Versioning

- Confirm the release tag follows semantic versioning.
- Confirm the version reported by `osf --version` matches the release tag or release metadata.
- Record any versioning notes in the release notes before publishing.

## Binary Matrix

- Build a Windows binary.
- Build a macOS binary.
- Build a Linux binary.
- Verify the expected archive names or executables are present for each platform.
- For local Windows checks, use `.\scripts\build.ps1` or `make build` to produce `bin\osf.exe`.

## Checksums

- Generate checksums for every release binary.
- Publish the checksum file with the release artifacts.
- Verify the published checksums match the built binaries before release.

## Validation

- Run `go run ./cmd/osf --help`.
- Run `go run ./cmd/osf --version`.
- Run `go run ./cmd/osf completion bash`.
- Run `go run ./cmd/osf completion powershell`.
- Run `go run ./tools/checkstubs`.
- Run `go run ./tools/checkproviderrelease -report docs/multi-provider-validation-report.md` and review every validation level and resource disposition.
- Run `go test ./internal/cli ./internal/mcpserver -run 'CompatibilityFixture|RootContractMatchesCompatibilityFixture'`.
- Run `go test ./...` from a clean checkout.
- Confirm the README, contributing guide, and security notes still match the released command surface and auth rules.

## Safety Review

- Confirm the release docs do not describe planned commands as already implemented.
- Confirm `OSF_TOKEN` remains the preferred auth path and that `OSF_USERNAME`/`OSF_PASSWORD` fallback limitations, SSO/2FA caveats, and guided token bootstrap behavior are documented.
- Confirm live OSF tests remain opt-in and are not implied by default release validation.
- Confirm Zenodo/cross-provider live jobs remain manual and false by default, and review deletion or retained-record evidence for every sandbox resource.
- Confirm `server.json` and marketplace metadata advertise only implemented public read surfaces; internal provider writes must not appear as MCP tools.
- Review `docs/compatibility-policy.md`, `docs/support-policy.md`, and
  `docs/live-validation-matrix.md`, `docs/provider-release-operations.md`, and
  `docs/adr-001-multi-provider-release-contract.md`.
- Attach `docs/release-candidate-evidence.md` with every release-candidate
  decision and record any explicit waivers.

## Publishing

```powershell
# Tag the release
git tag v1.0.0
git push origin v1.0.0

# The GoReleaser config (publishing disabled) is in .goreleaser.yaml
# Enable by setting `release.disable: false` or using --release-notes
```
