# Spec: Release Packaging

Prepare the CLI for repeatable local and GitHub Actions release builds.

## Outcomes

- `go build` produces a versioned `osf` binary with commit/date metadata.
- Release checks are scriptable on Windows and CI.
- Shell completions are available through Cobra.
- GoReleaser or equivalent release configuration is present but conservative, with publishing disabled until explicitly enabled.

## Non-Goals

- Publishing a public release.
- Signing or notarization unless added in a later release-hardening track.
