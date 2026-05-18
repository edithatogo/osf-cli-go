# Plan: Preprints, Search, Registrations, and Add-ons

## Reconciliation Note

Status reconciled on 2026-05-16 against the current codebase. Preprint listing, provider filtering, typed OSF search results, registration draft creation, and add-on listing are implemented and offline-tested as of 2026-05-16.

## Phase 1: Preprint Listing

- [x] Task: Add `ListPreprints` method to the OSF API client with pagination
- [x] Task: Add preprint provider filter support
- [x] Task: Add `osf preprints list` CLI command with JSON and table output
- [x] Task: Add fixture-backed tests for preprint listing

## Phase 2: Search Across OSF

- [x] Task: Add `SearchOSF` method sending GET `/v2/search/?q=<query>`
- [x] Task: Map search result types (nodes, preprints, registrations, files) to typed responses
- [x] Task: Add `osf search <query>` CLI command
- [x] Task: Add fixture-backed tests for search results

## Phase 3: Registration Creation

- [x] Task: Add `CreateRegistration` method sending POST to `/v2/nodes/{id}/registrations/`
- [x] Task: Handle registration schema selection and draft registration parameters
- [x] Task: Add `osf registrations create <node-id>` CLI command with confirmation
- [x] Task: Add fixture-backed tests for registration creation

## Phase 4: Add-on Listing

- [x] Task: Add `ListNodeAddons` method sending GET `/v2/nodes/{id}/addons/`
- [x] Task: Add `osf files addons <node-id>` CLI command listing configured storage providers
- [x] Task: Add fixture-backed tests for add-on listing
- [x] Task: Run quality gates and `$conductor-review`, apply fixes, and write phase review evidence
