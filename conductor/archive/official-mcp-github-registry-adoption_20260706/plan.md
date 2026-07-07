# Plan: Official MCP/GitHub Registry Adoption

## Phase 1: Registry State Audit

- [x] Task: Inspect current official registry, GitHub MCP registry, GHCR image,
  workflow, and `server.json` state.
    - [x] Record registry URLs, package identifiers, versions, and evidence.
    - [x] Identify stale metadata, missing fields, or score/quality gaps.
- [x] Task: Conductor - Automated Review and Checkpoint 'Registry State Audit' (Protocol in workflow.md)

## Phase 2: Metadata And Validation Hardening

- [x] Task: Add or update repo-local validators for official registry metadata.
    - [x] Cover `server.json`, GHCR package reference, auth/env metadata, and
      install instructions.
    - [x] Add fixture or snapshot checks where live validation is not safe.
- [x] Task: Improve metadata and docs to close every validated gap.
- [x] Task: Conductor - Automated Review and Checkpoint 'Metadata And Validation Hardening' (Protocol in workflow.md)

## Phase 3: Submission Or Refresh

- [x] Task: Run safe non-interactive publish/refresh path if needed.
    - [x] Use existing GitHub Actions or registry CLI/API where available.
    - [x] Use Chrome for browser-only verification; request user login if auth
      cannot be completed automatically.
- [x] Task: Store receipts, workflow URLs, API responses, and remaining blockers.
- [x] Task: Conductor - Automated Review and Checkpoint 'Submission Or Refresh' (Protocol in workflow.md)

## Phase 4: Final Quality Pass

- [x] Task: Run `go test ./...`, `go vet ./...`, `go run ./tools/checkstubs`,
  `go run ./tools/checkreviews`, registry checks, and lint if available.
- [x] Task: Write final phase-review evidence and mark quality/score outcome.
- [x] Task: Conductor - Automated Review and Checkpoint 'Final Quality Pass' (Protocol in workflow.md)
