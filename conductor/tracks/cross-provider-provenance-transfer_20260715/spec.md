# Cross-provider copy, provenance, and failure recovery

## Objective

Design explicit OSF/Zenodo copy workflows that preserve identity and provenance,
surface semantic loss before writes, and recover safely from partial failure.

## Requirements

- Provide dry-run mapping reports for metadata, files, access, embargo, license,
  identifiers, versions, and unsupported fields.
- Record source-qualified identity and every transformation; never silently mirror.
- Use deterministic idempotency keys, checkpoints, conflict policy, and a saga-
  style compensation model for multi-step writes.
- Require explicit direction, destination, authorization, and publish intent.

## Completion evidence

Failure-injection and replay tests prove truthful partial results, idempotent
recovery, provenance completeness, and no accidental publication.
