# Plan: MCP.Directory Adoption

## Phase 1: Directory Audit

- [x] Task: Search MCP.Directory for existing OSF CLI Go listing and submission status.
    - [x] Record listing URL, publisher fields, install path, score, and gaps.
    - [x] Compare live listing fields to repo-local submission metadata.
- [x] Task: Conductor - Automated Review and Checkpoint 'Directory Audit' (Protocol in workflow.md)

## Phase 2: Submission Packet Hardening

- [x] Task: Update submission copy, metadata, categories, install commands, and
  evidence fields.
    - [x] Add validation for any new MCP.Directory-specific fields.
    - [x] Run JSON, Go, registry, anti-stub, and review checks.
- [x] Task: Conductor - Automated Review and Checkpoint 'Submission Packet Hardening' (Protocol in workflow.md)

## Phase 3: Browser Submission And Score Iteration

- [x] Task: Use Chrome to submit or update the listing when browser-only.
    - [x] Ask the user to log in if authentication blocks the flow.
    - [x] Iterate quality/score fixes toward 100% where MCP.Directory exposes a score.
- [x] Task: Store receipt, listing URL, score outcome, and blockers.
- [x] Task: Conductor - Automated Review and Checkpoint 'Browser Submission And Score Iteration' (Protocol in workflow.md)

## Phase 4: Final Review

- [x] Task: Run complete validation and conductor-review.
- [x] Task: Write final phase-review evidence.
- [x] Task: Conductor - Automated Review and Checkpoint 'Final Review' (Protocol in workflow.md)
