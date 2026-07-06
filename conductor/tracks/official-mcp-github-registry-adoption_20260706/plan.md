# Plan: Official MCP/GitHub Registry Adoption

## Phase 1: Registry State Audit

- [ ] Task: Inspect current official registry, GitHub MCP registry, GHCR image,
  workflow, and `server.json` state.
    - [ ] Record registry URLs, package identifiers, versions, and evidence.
    - [ ] Identify stale metadata, missing fields, or score/quality gaps.
- [ ] Task: Conductor - Automated Review and Checkpoint 'Registry State Audit' (Protocol in workflow.md)

## Phase 2: Metadata And Validation Hardening

- [ ] Task: Add or update repo-local validators for official registry metadata.
    - [ ] Cover `server.json`, GHCR package reference, auth/env metadata, and
      install instructions.
    - [ ] Add fixture or snapshot checks where live validation is not safe.
- [ ] Task: Improve metadata and docs to close every validated gap.
- [ ] Task: Conductor - Automated Review and Checkpoint 'Metadata And Validation Hardening' (Protocol in workflow.md)

## Phase 3: Submission Or Refresh

- [ ] Task: Run safe non-interactive publish/refresh path if needed.
    - [ ] Use existing GitHub Actions or registry CLI/API where available.
    - [ ] Use Chrome for browser-only verification; request user login if auth
      cannot be completed automatically.
- [ ] Task: Store receipts, workflow URLs, API responses, and remaining blockers.
- [ ] Task: Conductor - Automated Review and Checkpoint 'Submission Or Refresh' (Protocol in workflow.md)

## Phase 4: Final Quality Pass

- [ ] Task: Run `go test ./...`, `go vet ./...`, `go run ./tools/checkstubs`,
  `go run ./tools/checkreviews`, registry checks, and lint if available.
- [ ] Task: Write final phase-review evidence and mark quality/score outcome.
- [ ] Task: Conductor - Automated Review and Checkpoint 'Final Quality Pass' (Protocol in workflow.md)
