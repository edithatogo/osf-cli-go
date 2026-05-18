# Plan: Documentation Overhaul

## Reconciliation Note

Status reconciled on 2026-05-16 against the current codebase. Install, usage, architecture, command reference, examples, developer docs, package-level docs, example tests, and phase-review evidence are present as of 2026-05-16.

## Phase 1: Go Doc Comments

- [x] Task: Add/verify Go doc comments on all public types and functions in internal/auth
- [x] Task: Add/verify Go doc comments on all public types and functions in internal/osfapi
- [x] Task: Add/verify Go doc comments on all public types and functions in internal/download
- [x] Task: Add/verify Go doc comments on all public types and functions in internal/output
- [x] Task: Add package-level doc comments for all internal packages

## Phase 2: User Documentation

- [x] Task: Update README with comprehensive usage guide and quickstart
- [x] Task: Add docs/install.md with installation instructions
- [x] Task: Add docs/commands.md with full CLI command reference
- [x] Task: Add docs/examples.md with worked examples
- [x] Task: Add example test files (*_test.go with Example functions)

## Phase 3: Developer Documentation

- [x] Task: Add docs/architecture.md describing package layout and design decisions
- [x] Task: Add docs/contributing.md with development workflow and conventions
- [x] Task: Add CI check for documentation freshness and broken links

## Phase 4: Review

- [x] Task: Run quality gates and tests
- [x] Task: Run `$conductor-review`, apply fixes, re-run review, and write phase review evidence
