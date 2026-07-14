# Implementation Plan

## Phase 1: Contract and threat model

- [ ] Task: Define checkpoint manifest, identity, versioning, and invalidation rules.
- [ ] Task: Define range, retry, cancellation, conflict, and atomic-finalization behavior.
- [ ] Task: Conductor - User Manual Verification 'Contract and threat model' (Protocol in workflow.md)

## Phase 2: Implementation and tests

- [ ] Task: Implement resumable download checkpoints and verified finalization.
- [ ] Task: Implement resumable upload checkpoints where the provider supports recovery.
- [ ] Task: Add interruption, restart, checksum, stale-state, race, fuzz, and cross-platform tests.
- [ ] Task: Conductor - User Manual Verification 'Implementation and tests' (Protocol in workflow.md)

## Phase 3: Release integration

- [ ] Task: Update CLI JSON, manifests, documentation, and feature matrix.
- [ ] Task: Run full repository and release-candidate gates.
- [ ] Task: Conductor - User Manual Verification 'Release integration' (Protocol in workflow.md)
