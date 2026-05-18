# Plan: OSF API Coverage Extension

## Reconciliation Note

Status reconciled on 2026-05-17 against the current codebase. Planned endpoint coverage, research mapping, confirmation-gated write CLI coverage, and phase-review evidence are present and offline-tested.

## Phase 1: Research and Mapping

- [x] Task: Research osfclient, osfr, osf-project-exporter API coverage
- [x] Task: Fetch OSF API v2 docs and list all endpoints
- [x] Task: Create comparison matrix of supported endpoints across tools

## Phase 2: Read-Only Endpoint Extensions

- [x] Task: Add registrations listing endpoint
- [x] Task: Add wiki content endpoint
- [x] Task: Add comments listing endpoint
- [x] Task: Add node logs endpoint
- [x] Task: Add identifiers endpoint

## Phase 3: Write Endpoints

- [x] Task: Add node creation POST endpoint
- [x] Task: Add file upload endpoint
- [x] Task: Add CLI commands for write operations with explicit confirmation

## Phase 4: Review

- [x] Task: Run quality gates and tests
- [x] Task: Run `$conductor-review`, apply fixes, re-run review, and write phase review evidence
