# Parallel Agent Work Plan

## Dependency Order

1. `cli-contract-routing_20260502`
2. `api-client-fixtures_20260502`
3. `auth-public-access_20260502`
4. `readonly-commands_20260502` and `download-safety_20260502`
5. `docs-release-readiness_20260502`

`repo-quality-automation_20260502` and `quality-review-automation_20260502` can proceed in parallel with the CLI and API contract tracks.

## Agent Ownership

- Agent A owns CLI routing and command contracts: `cmd/osf`, `internal/cli`, CLI tests, command contract docs.
- Agent B owns OSF API client and fixtures: `internal/osfapi`, API fixtures, API client tests.
- Agent C owns auth: `internal/auth`, auth tests, token docs.
- Agent D owns read-only commands after A and B stabilize their contracts.
- Agent E owns download safety: `internal/download`, download tests, download docs.
- Agent F owns documentation and release readiness: README, contributing docs, release checklist.
- Agent G owns repo quality automation: CI workflows, lint config, Renovate, local check scripts.
- Agent H owns review automation: anti-stub checks, phase review templates, workflow enforcement.

## Conflict Rules

- Agents must not rewrite another agent's owned package without coordinating in the track plan.
- Shared contract changes go through `cli-contract-routing_20260502` or `api-client-fixtures_20260502`.
- Every phase exit must run the review-fix-continue protocol from `conductor/workflow.md`.
