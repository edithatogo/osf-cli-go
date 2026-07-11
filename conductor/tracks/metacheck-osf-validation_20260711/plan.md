# Implementation Plan

## Phase 1: Evidence and Gap Analysis

- [ ] Task: Audit scienceverse/metacheck capabilities, activity, license, tests, release maturity, and user workflows
    - [ ] Capture source URLs and dated evidence
    - [ ] Compare against current CLI, API, MCP, packaging, and documentation
- [ ] Task: Update the competitive comparison table and classify every gap
- [ ] Task: Conductor - Automated Review and Checkpoint 'Evidence and Gap Analysis' (Protocol in workflow.md)

## Phase 2: Test-Driven Parity Work

- [ ] Task: Add failing offline tests for each accepted capability gap
- [ ] Task: Implement accepted capabilities using existing repository patterns
    - [ ] Preserve conservative write and authentication behavior
    - [ ] Add CLI and MCP exposure where applicable
- [ ] Task: Document rejected or deferred capabilities with rationale and linked issues
- [ ] Task: Conductor - Automated Review and Checkpoint 'Test-Driven Parity Work' (Protocol in workflow.md)

## Phase 3: Validation and Closeout

- [ ] Task: Update user docs, examples, release metadata, and the comparison matrix
- [ ] Task: Run full local and CI quality gates
- [ ] Task: Reconcile GitHub issue #20 with evidence and remaining external validation
- [ ] Task: Conductor - Automated Review and Checkpoint 'Validation and Closeout' (Protocol in workflow.md)

