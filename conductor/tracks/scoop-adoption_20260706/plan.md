# Plan: Scoop Adoption

## Phase 1: Scoop Requirements Audit

- [ ] Task: Determine Scoop route: project bucket, Extras, or blocker.
    - [ ] Audit release artifact URLs, hashes, binary layout, and autoupdate needs.
    - [ ] Record validation tooling and submission target.
- [ ] Task: Conductor - Automated Review and Checkpoint 'Scoop Requirements Audit' (Protocol in workflow.md)

## Phase 2: Manifest Implementation

- [ ] Task: Add Scoop manifest or submission packet.
    - [ ] Validate manifest JSON, hash fields, bin entries, and autoupdate.
    - [ ] Update Windows install docs.
- [ ] Task: Conductor - Automated Review and Checkpoint 'Manifest Implementation' (Protocol in workflow.md)

## Phase 3: Submission

- [ ] Task: Submit PR or publish bucket route.
    - [ ] Prefer GitHub CLI; use Chrome for login if necessary.
    - [ ] Store PR URL, bucket URL, validation output, or blocker.
- [ ] Task: Conductor - Automated Review and Checkpoint 'Submission' (Protocol in workflow.md)

## Phase 4: Final Review

- [ ] Task: Run validation gates and conductor-review.
- [ ] Task: Write final phase-review evidence.
- [ ] Task: Conductor - Automated Review and Checkpoint 'Final Review' (Protocol in workflow.md)
