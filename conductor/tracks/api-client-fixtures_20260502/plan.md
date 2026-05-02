# Plan: OSF API Client And Fixtures

## Phase 1: Fixture Contracts

- [ ] Task: Add representative OSF JSON:API fixtures for users, nodes, components, contributors, files, pagination, and errors
- [ ] Task: Add tests that define parsing and pagination expectations before client implementation

## Phase 2: Client

- [ ] Task: Implement context-aware HTTP client with base URL and optional bearer token
- [ ] Task: Implement pagination traversal and typed status-aware errors
- [ ] Task: Add endpoint helpers needed by read-only commands

## Phase 3: Review

- [ ] Task: Run quality gates and anti-stub scan
- [ ] Task: Run `$conductor-review`, apply fixes, re-run review, and write phase review evidence
