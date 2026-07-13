# Plan: Codex Marketplace Adoption

Score target: 100% of applicable criteria in
[`docs/registry-scorecard.md`](../../../docs/registry-scorecard.md). A local
pass is not an OpenAI listing or approval.

## Phase 1: Marketplace Requirements Audit

- [x] Task: Verify current Codex Marketplace submission and validation rules.
    - [x] Compare rules against `plugins/codex-osf` and `.agents/plugins`.
    - [x] Record score/validation output and missing fields.
- [x] Task: Conductor - Automated Review and Checkpoint 'Marketplace Requirements Audit' (Protocol in workflow.md)

## Phase 2: Plugin Hardening

- [x] Task: Improve Codex plugin metadata, skill text, MCP config, and archive
  packaging until validators pass.
    - [x] Add repo-local checks for marketplace JSON and install paths.
    - [x] Run full local validation.
- [x] Task: Conductor - Automated Review and Checkpoint 'Plugin Hardening' (Protocol in workflow.md)

## Phase 3: Submission

- [x] Task: Use the available provider surface to submit to Codex Marketplace or record blocker.
    - [x] Confirm local marketplace discovery and installation with an isolated Codex home.
    - [x] Record that public Plugin Directory publication requires the provider/workspace publication gate; no submission is claimed.
- [x] Task: Store receipt, review URL, listing URL, score, and blockers.
- [x] Task: Re-run the universal scorecard after every metadata or release change.
- [x] Task: Conductor - Automated Review and Checkpoint 'Submission' (Protocol in workflow.md)

## Phase 4: Final Review

- [x] Task: Run validation gates and conductor-review.
- [x] Task: Write final phase-review evidence.
- [x] Task: Conductor - Automated Review and Checkpoint 'Final Review' (Protocol in workflow.md)
