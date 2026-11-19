# Plan: Repo Quality Automation

## Phase 1: Local Quality Commands

- [x] Task: Add repo management files for editor, Git attributes, lint config, and local check commands
- [x] Task: Add production-code anti-stub scanner

## Phase 2: Hosted CI

- [x] Task: Add GitHub Actions for CI, lint, security, and docs checks
- [x] Task: Add Renovate configuration for Go modules and GitHub Actions

## Phase 3: Review

- [x] Task: Run quality gates and anti-stub scan
- [x] Task: Run `$conductor-review`, apply fixes, re-run review, and write phase review evidence

## Phase 4: Bounded fuzzing campaigns (2026-11-20)

- [x] Task: Extend the five existing fuzz targets with native-JSON preservation and resumption-token invariants; all inputs synthetic and bounded.
- [x] Task: Move mutation campaigns to a path-scoped PR/scheduled/manual workflow with time, worker and memory budgets and retained reproducers.
- [x] Task: Execute the seven targets, ordinary/race tests, vet and anti-stub checks and record the exact toolchain/results in `phase-4-review.md`.
- [ ] Task: Verify hosted checks on the candidate head before admission.

No live provider validation, credentials, release action or toolchain upgrade is part of this phase. PR #152 separately owns the Go security toolchain update.
