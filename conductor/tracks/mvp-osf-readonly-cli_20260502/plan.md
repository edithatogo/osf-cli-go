# Plan: MVP OSF Read-Only CLI Roadmap

This track is the parent roadmap. Implementation is split across subsystem tracks listed in `conductor/tracks.md` so multiple agents can work in parallel without colliding.

## Phase 1: Foundation And Contracts

- [x] Task: Create Go module and baseline `osf` command scaffold
- [x] Task: Complete `cli-contract-routing_20260502`
- [x] Task: Complete `api-client-fixtures_20260502`
- [ ] Task: Complete `repo-quality-automation_20260502`
- [x] Task: Complete `quality-review-automation_20260502`

## Phase 2: Read-Only MVP

- [ ] Task: Complete `auth-public-access_20260502`
- [x] Task: Complete `readonly-commands_20260502`
- [x] Task: Complete `download-safety_20260502`

## Phase 3: Release Readiness

- [x] Task: Complete `docs-release-readiness_20260502`
- [ ] Task: Review package boundaries for `mcp-server-roadmap_20260502`
- [ ] Task: Run final `$conductor-review`, apply fixes, re-run review, and write release review evidence
