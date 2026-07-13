# Plan: Codex Marketplace Adoption

Score target: 100% of applicable criteria in
[`docs/registry-scorecard.md`](../../../docs/registry-scorecard.md). A local
pass is not an OpenAI listing or approval.

## Phase 1: Marketplace Requirements Audit

- [ ] Task: Verify current Codex Marketplace submission and validation rules.
    - [ ] Compare rules against `plugins/codex-osf` and `.agents/plugins`.
    - [ ] Record score/validation output and missing fields.
- [ ] Task: Conductor - Automated Review and Checkpoint 'Marketplace Requirements Audit' (Protocol in workflow.md)

## Phase 2: Plugin Hardening

- [ ] Task: Improve Codex plugin metadata, skill text, MCP config, and archive
  packaging until validators pass.
    - [ ] Add repo-local checks for marketplace JSON and install paths.
    - [ ] Run full local validation.
- [ ] Task: Conductor - Automated Review and Checkpoint 'Plugin Hardening' (Protocol in workflow.md)

## Phase 3: Submission

- [ ] Task: Use Chrome to submit to Codex Marketplace or record blocker.
    - [ ] Ask user to log in if required.
    - [ ] Iterate on score/validation feedback toward 100%.
- [ ] Task: Store receipt, review URL, listing URL, score, and blockers.
- [ ] Task: Re-run the universal scorecard after every metadata or release change.
- [ ] Task: Conductor - Automated Review and Checkpoint 'Submission' (Protocol in workflow.md)

## Phase 4: Final Review

- [ ] Task: Run validation gates and conductor-review.
- [ ] Task: Write final phase-review evidence.
- [ ] Task: Conductor - Automated Review and Checkpoint 'Final Review' (Protocol in workflow.md)
