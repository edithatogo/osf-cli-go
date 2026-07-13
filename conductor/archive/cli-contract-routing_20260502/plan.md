# Plan: CLI Contract And Routing

## Phase 1: Contract

- [x] Task: Document command vocabulary, global flags, output modes, and exit codes
- [x] Task: Add command-level acceptance examples for help, version, JSON mode, and unknown commands

## Phase 2: Cobra Routing

- [x] Task: Add Cobra dependency and root command
- [x] Task: Replace current flag parsing with Cobra command execution
- [x] Task: Add tests for help, version, unknown command, and output mode behavior

## Phase 3: Review

- [x] Task: Run quality gates and anti-stub scan
- [x] Task: Run `$conductor-review`, apply fixes, re-run review, and write phase review evidence
