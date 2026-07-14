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
- Keep one-shot WaterButler uploads explicit; expose a reusable acknowledged
  chunk session contract for providers that support resumable upload recovery.

## Evidence

- `internal/download/resume.go` provides fingerprinted checkpoints, stale-state
  invalidation, range fallback, checksum verification, and atomic finalization.
- `internal/download/resume_upload.go` provides acknowledged-chunk recovery
  without claiming resumability for one-shot WaterButler uploads.
- `internal/download/folder.go` and `internal/cli/files_download.go` expose
  resume state in folder manifests and single-file JSON output.
- `internal/download/*_test.go` covers interruption, restart, checksum failure,
  stale checkpoints, range fallback, folder-tree recovery, upload recovery,
  no-progress rejection, and multi-chunk acknowledgement; race tests cover the
  transfer packages.
- `docs/download-resume.md` documents checkpoint files, conflict behavior,
  provider boundaries, and recovery semantics.

## Out of scope

Background bidirectional synchronization, file watching, and implicit writes.
