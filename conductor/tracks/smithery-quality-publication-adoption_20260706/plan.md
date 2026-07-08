# Plan: Smithery Quality Publication Adoption

## Phase 1: Smithery Listing And Score Audit

- [x] Task: Inspect current Smithery listing, release status, MCP URL, and score.
    - [x] Capture score components, warnings, and recommended fixes.
    - [x] Record whether CLI/API or Chrome login is required.
- [x] Task: Conductor - Automated Review and Checkpoint 'Smithery Listing And Score Audit' (Protocol in workflow.md)

## Phase 2: MCPB Quality Hardening

- [x] Task: Validate MCPB manifest and bundle path.
    - [x] Run repo-local `checkregistries` and MCPB CLI validation where available.
    - [x] Fix tool schema, auth, metadata, README, or package issues found.
- [x] Task: Iterate Smithery score improvements toward 100%.
- [x] Task: Conductor - Automated Review and Checkpoint 'MCPB Quality Hardening' (Protocol in workflow.md)

## Phase 3: Publish Or Refresh

- [x] Task: Submit or refresh Smithery release using the safest available path.
    - [x] Prefer CLI/API where authenticated.
    - [x] Use Chrome when browser auth is required; ask user to log in if blocked.
- [x] Task: Store receipts, release IDs, MCP URL, score outcome, and blockers.
- [x] Task: Conductor - Automated Review and Checkpoint 'Publish Or Refresh' (Protocol in workflow.md)

## Phase 4: Final Validation

- [x] Task: Run Go, JSON, registry, anti-stub, review, and lint validation.
- [x] Task: Write final phase-review evidence.
- [x] Task: Conductor - Automated Review and Checkpoint 'Final Validation' (Protocol in workflow.md)
