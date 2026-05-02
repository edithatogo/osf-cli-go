# Plan: Live OSF Validation

## Phase 1: Harness

- [x] Task: Add an opt-in live validation script with strict environment variable checks and token redaction
- [x] Task: Add dry-run/self-test coverage for missing environment variables and evidence file creation

## Phase 2: Command Coverage

- [x] Task: Cover authenticated identity and read-only listing commands
- [x] Task: Add `files download` live validation after the download CLI track lands

## Phase 3: Evidence And Review

- [x] Task: Document required environment variables and safe evidence handling
- [x] Task: Run build, quality gates, `$conductor-review`, apply fixes, re-run review, and write phase review evidence
