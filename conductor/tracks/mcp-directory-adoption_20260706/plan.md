# Plan: MCP.Directory Adoption

## Phase 1: Directory Audit

- [ ] Task: Search MCP.Directory for existing OSF CLI Go listing and submission status.
    - [ ] Record listing URL, publisher fields, install path, score, and gaps.
    - [ ] Compare live listing fields to repo-local submission metadata.
- [ ] Task: Conductor - Automated Review and Checkpoint 'Directory Audit' (Protocol in workflow.md)

## Phase 2: Submission Packet Hardening

- [ ] Task: Update submission copy, metadata, categories, install commands, and
  evidence fields.
    - [ ] Add validation for any new MCP.Directory-specific fields.
    - [ ] Run JSON, Go, registry, anti-stub, and review checks.
- [ ] Task: Conductor - Automated Review and Checkpoint 'Submission Packet Hardening' (Protocol in workflow.md)

## Phase 3: Browser Submission And Score Iteration

- [ ] Task: Use Chrome to submit or update the listing when browser-only.
    - [ ] Ask the user to log in if authentication blocks the flow.
    - [ ] Iterate quality/score fixes toward 100% where MCP.Directory exposes a score.
- [ ] Task: Store receipt, listing URL, score outcome, and blockers.
- [ ] Task: Conductor - Automated Review and Checkpoint 'Browser Submission And Score Iteration' (Protocol in workflow.md)

## Phase 4: Final Review

- [ ] Task: Run complete validation and conductor-review.
- [ ] Task: Write final phase-review evidence.
- [ ] Task: Conductor - Automated Review and Checkpoint 'Final Review' (Protocol in workflow.md)
