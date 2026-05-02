# Plan: OSF API Client And Fixtures

## Phase 1: Fixture Contracts

- [x] Task: Add representative OSF JSON:API fixtures for users, nodes, components, contributors, files, pagination, and errors
- [x] Task: Add tests that define parsing and pagination expectations before client implementation

## Phase 2: Client

- [x] Task: Implement context-aware HTTP client with base URL and optional bearer token
- [x] Task: Implement pagination traversal and typed status-aware errors
- [x] Task: Add endpoint helpers needed by read-only commands

## Phase 3: Review

- [x] Task: Run quality gates and anti-stub scan
- [x] Task: Run `$conductor-review`, apply fixes, re-run review, and write phase review evidence
