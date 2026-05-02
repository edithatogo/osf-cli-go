# Plan: Coverage Hardening

## Phase 1: Coverage Baseline

- [ ] Task: Capture current coverage and identify high-value uncovered branches
- [ ] Task: Add tests for CLI usage and error handling paths

## Phase 2: Package-Specific Risk Areas

- [ ] Task: Add auth redaction and missing-token edge case tests
- [ ] Task: Add OSF API error parsing and endpoint resolution tests
- [ ] Task: Add download failure, skip, and symlink/path tests where missing
- [ ] Task: Add output helper tests

## Phase 3: Review

- [ ] Task: Run build, quality gates, `$conductor-review`, apply fixes, re-run review, and write phase review evidence
