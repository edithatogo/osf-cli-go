# Implementation Plan

## Phase 1: Contract and threat model

- [ ] Task: Freeze CLI JSON and MCP schemas and write migration fixtures
- [ ] Task: Complete threat model, write approval model, secret-handling review, and support policy
- [ ] Task: Conductor - Automated Review and Checkpoint 'Contract and threat model' (Protocol in workflow.md)

## Phase 2: Reliability and supply chain

- [ ] Task: Add fuzz/property, race, cancellation, retry, pagination, and large-transfer tests
- [ ] Task: Add cross-platform clean-install and rollback verification
- [ ] Task: Harden SBOM, provenance, signing, container, dependency, and artifact verification
- [ ] Task: Conductor - Automated Review and Checkpoint 'Reliability and supply chain' (Protocol in workflow.md)

## Phase 3: Operations and live readiness

- [ ] Task: Define structured logging, metrics, incident response, support, and release runbooks
- [ ] Task: Run opt-in live OSF validation and record sanitized evidence
- [ ] Task: Reconcile all 1.0 launch gates with issue #52
- [ ] Task: Conductor - Automated Review and Checkpoint 'Operations and live readiness' (Protocol in workflow.md)
