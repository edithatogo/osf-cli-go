# Plan: Files Download CLI

## Phase 1: Command Contract

- [x] Task: Define `osf files download` flags, arguments, conflict policy names, and output contracts
- [x] Task: Add CLI tests for help, argument validation, conflict policy parsing, and JSON/human output

## Phase 2: API And Download Integration

- [x] Task: Add API client support for resolving file metadata and opening download streams with fixture-backed tests
- [x] Task: Wire the command to `internal/download` for single-file and folder-tree downloads
- [x] Task: Ensure path traversal, symlink escapes, failed streams, and existing-file behavior are covered

## Phase 3: Docs And Review

- [x] Task: Update README and docs with runnable offline examples and opt-in live examples
- [x] Task: Run build, quality gates, `$conductor-review`, apply fixes, re-run review, and write phase review evidence
