# Plan: Preprints, Search, Registrations, and Add-ons

## Phase 1: Preprint Listing

- [ ] Task: Add `ListPreprints` method to the OSF API client with pagination and provider filter
- [ ] Task: Add `osf preprints list` CLI command with JSON and table output
- [ ] Task: Add fixture-backed tests for preprint listing

## Phase 2: Search Across OSF

- [ ] Task: Add `SearchOSF` method sending GET `/v2/search/?q=<query>`
- [ ] Task: Map search result types (nodes, preprints, registrations, files) to typed responses
- [ ] Task: Add `osf search <query>` CLI command
- [ ] Task: Add fixture-backed tests for search results

## Phase 3: Registration Creation

- [ ] Task: Add `CreateRegistration` method sending POST to `/v2/nodes/{id}/registrations/`
- [ ] Task: Handle registration schema selection and draft registration parameters
- [ ] Task: Add `osf registrations create <node-id>` CLI command with confirmation
- [ ] Task: Add fixture-backed tests for registration creation

## Phase 4: Add-on Listing

- [ ] Task: Add `ListNodeAddons` method sending GET `/v2/nodes/{id}/addons/`
- [ ] Task: Add `osf files addons <node-id>` CLI command listing configured storage providers
- [ ] Task: Add fixture-backed tests for add-on listing
- [ ] Task: Run quality gates and `$conductor-review`, apply fixes, and write phase review evidence
