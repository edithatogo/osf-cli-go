# Implementation Plan

## Phase 1: Contract and threat model

- [x] Task: Define checkpoint manifest, identity, versioning, and invalidation rules.
- [x] Task: Define range, retry, cancellation, conflict, and atomic-finalization behavior.
- [x] Task: Conductor - User Manual Verification 'Contract and threat model' (Protocol in workflow.md)

## Phase 2: Implementation and tests

- [x] Task: Implement resumable download checkpoints and verified finalization.
- [x] Task: Implement resumable upload checkpoints where the provider supports recovery.
- [x] Task: Add interruption, restart, checksum, stale-state, race, fuzz, and cross-platform tests.
- [x] Task: Conductor - User Manual Verification 'Implementation and tests' (Protocol in workflow.md)

## Phase 3: Release integration

- [x] Task: Update CLI JSON, manifests, documentation, and feature matrix.
- [x] Task: Run full repository and release-candidate gates.
- [x] Task: Conductor - User Manual Verification 'Release integration' (Protocol in workflow.md)
