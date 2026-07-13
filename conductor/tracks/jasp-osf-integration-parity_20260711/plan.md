# Implementation Plan

## Phase 1: Evidence and Gap Analysis

- [x] Task: Audit jasp-stats/jasp-desktop capabilities, activity, license, tests, release maturity, and user workflows
    - [x] Capture source URLs and dated evidence
    - [x] Compare against current CLI, API, MCP, packaging, and documentation
- [x] Task: Update the competitive comparison table and classify every gap
- [x] Task: Conductor - Automated Review and Checkpoint 'Evidence and Gap Analysis' (Protocol in workflow.md)

## Phase 2: Test-Driven Parity Work

- [x] Task: Add failing offline tests for each accepted capability gap
- [x] Task: Implement accepted capabilities using existing repository patterns
    - [x] Preserve conservative write and authentication behavior
    - [x] Add CLI and MCP exposure where applicable
- [x] Task: Document rejected or deferred capabilities with rationale and linked issues
- [x] Task: Conductor - Automated Review and Checkpoint 'Test-Driven Parity Work' (Protocol in workflow.md)

## Phase 3: Validation and Closeout

- [x] Task: Update user docs, examples, release metadata, and the comparison matrix
- [x] Task: Run full local and CI quality gates
- [x] Task: Reconcile GitHub issue #18 with evidence and remaining external validation
- [x] Task: Conductor - Automated Review and Checkpoint 'Validation and Closeout' (Protocol in workflow.md)
