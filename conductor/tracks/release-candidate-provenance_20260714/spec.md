# Release-candidate provenance

GitHub issue: [#98](https://github.com/edithatogo/osf-cli-go/issues/98)

## Objective

Produce and independently verify a complete, signed release candidate before a
`v1.0.0` tag.

## Requirements

- Generate binaries, checksums, SBOMs, provenance attestations, Cosign
  signatures, OCI images, MCPB packages, and plugin archives.
- Verify artifacts from a clean environment without the source checkout.
- Validate supported-platform installation and version output.
- Review workflow permissions, pinned actions, retention, and metadata.
- Record exact commit, digest, verification commands, and failures.
