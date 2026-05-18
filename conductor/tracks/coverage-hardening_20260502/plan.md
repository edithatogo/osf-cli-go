# Plan: Coverage Hardening

## Reconciliation Note

Status reconciled on 2026-05-16 against the current codebase. The targeted auth, API, download, output, and CLI error-path tests are present. A refreshed coverage baseline, coverage-risk summary, and phase-review evidence were recorded on 2026-05-16.

## Phase 1: Coverage Baseline

- [x] Task: Capture current coverage and identify high-value uncovered branches
- [x] Task: Add tests for CLI usage and error handling paths

## Phase 2: Package-Specific Risk Areas

- [x] Task: Add auth redaction and missing-token edge case tests
- [x] Task: Add OSF API error parsing and endpoint resolution tests
- [x] Task: Add download failure, skip, and symlink/path tests where missing
- [x] Task: Add output helper tests

## Phase 3: Review

- [x] Task: Run build, quality gates, `$conductor-review`, apply fixes, re-run review, and write phase review evidence
