# Zenodo sandbox transfers and integrity

## Objective

Implement safe upload and download behavior with disposable sandbox validation
before any production Zenodo write capability is claimed.

## Requirements

- Separate `ZENODO_TOKEN`, `ZENODO_BASE_URL`, and sandbox credentials from OSF.
- Support checksums, conflicts, cancellation, resumable checkpoints, limits,
  retries, cleanup, and truthful partial-failure reporting.
- Keep routine tests offline; require explicit opt-in for disposable sandbox use.

## Completion evidence

Offline failure injection passes and dated sandbox evidence verifies bytes,
checksums, resume behavior, cleanup, and secret redaction.
