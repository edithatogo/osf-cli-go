# Resumable transfers

GitHub issue: [#95](https://github.com/edithatogo/osf-cli-go/issues/95)

## Objective

Make interrupted downloads and uploads restartable and integrity-safe for the
1.0 release without changing the explicit-transfer and conflict-policy model.

## Requirements

- Persist checkpoint state in a deterministic manifest tied to source identity,
  destination, size, and checksum expectations.
- Resume only when the remote and local identities still match.
- Use range requests where supported and restart safely where they are not.
- Verify checksums before atomic finalization.
- Handle cancellation, retry, crash recovery, stale checkpoints, and conflicts.
- Expose resume state through CLI JSON and transfer manifests.
- Test interruption and recovery with deterministic HTTP fixtures, race tests,
  fuzz tests, and supported-platform path behavior.

## Out of scope

Background bidirectional synchronization, file watching, and implicit writes.
