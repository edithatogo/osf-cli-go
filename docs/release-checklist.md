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

## Checksums

- Generate checksums for every release binary.
- Publish the checksum file with the release artifacts.
- Verify the published checksums match the built binaries before release.

## Validation

- Run `go run ./cmd/osf --help`.
- Run `go run ./cmd/osf --version`.
- Run `go run ./tools/checkstubs`.
- Run `go test ./...` from a clean checkout.
- Confirm the README, contributing guide, and security notes still match the released command surface and auth rules.

## Safety Review

- Confirm the release docs do not describe planned commands as already implemented.
- Confirm `OSF_TOKEN` is documented as an environment variable only.
- Confirm live OSF tests remain opt-in and are not implied by default release validation.
