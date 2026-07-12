# Plan: Scoop Adoption

## Phase 1: Scoop Requirements Audit

- [x] Task: Determine Scoop route: project bucket, Extras, or blocker.
    - [x] Audit release artifact URLs, hashes, binary layout, and autoupdate needs.
    - [x] Record validation tooling and submission target.
- [x] Task: Conductor - Automated Review and Checkpoint 'Scoop Requirements Audit' (Protocol in workflow.md)

## Phase 2: Manifest Implementation

- [x] Task: Add Scoop manifest or submission packet.
    - [x] Validate manifest JSON, hash fields, bin entries, and autoupdate.
    - [x] Update Windows install docs.
- [x] Task: Conductor - Automated Review and Checkpoint 'Manifest Implementation' (Protocol in workflow.md)

## Phase 3: Submission

- [x] Task: Submit PR or publish bucket route.
    - [x] Use GitHub CLI; browser was not necessary.
    - [x] Store bucket URL, validation output, and the Main-bucket blocker.
- [x] Task: Conductor - Automated Review and Checkpoint 'Submission' (Protocol in workflow.md)

## Phase 4: Final Review

- [x] Task: Run validation gates and conductor-review.
- [x] Task: Write final phase-review evidence.
- [x] Task: Conductor - Automated Review and Checkpoint 'Final Review' (Protocol in workflow.md)
