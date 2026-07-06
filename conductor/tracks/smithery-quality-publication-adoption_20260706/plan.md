# Plan: Smithery Quality Publication Adoption

## Phase 1: Smithery Listing And Score Audit

- [ ] Task: Inspect current Smithery listing, release status, MCP URL, and score.
    - [ ] Capture score components, warnings, and recommended fixes.
    - [ ] Record whether CLI/API or Chrome login is required.
- [ ] Task: Conductor - Automated Review and Checkpoint 'Smithery Listing And Score Audit' (Protocol in workflow.md)

## Phase 2: MCPB Quality Hardening

- [ ] Task: Validate MCPB manifest and bundle path.
    - [ ] Run repo-local `checkregistries` and MCPB CLI validation where available.
    - [ ] Fix tool schema, auth, metadata, README, or package issues found.
- [ ] Task: Iterate Smithery score improvements toward 100%.
- [ ] Task: Conductor - Automated Review and Checkpoint 'MCPB Quality Hardening' (Protocol in workflow.md)

## Phase 3: Publish Or Refresh

- [ ] Task: Submit or refresh Smithery release using the safest available path.
    - [ ] Prefer CLI/API where authenticated.
    - [ ] Use Chrome when browser auth is required; ask user to log in if blocked.
- [ ] Task: Store receipts, release IDs, MCP URL, score outcome, and blockers.
- [ ] Task: Conductor - Automated Review and Checkpoint 'Publish Or Refresh' (Protocol in workflow.md)

## Phase 4: Final Validation

- [ ] Task: Run Go, JSON, registry, anti-stub, review, and lint validation.
- [ ] Task: Write final phase-review evidence.
- [ ] Task: Conductor - Automated Review and Checkpoint 'Final Validation' (Protocol in workflow.md)
