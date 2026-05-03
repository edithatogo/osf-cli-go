# Preprints, Search, Registrations, and Add-ons

## Objective

Add GET `/v2/preprints/`, GET `/v2/search/`, POST registration creation, and GET `/v2/nodes/{id}/addons/` endpoints to the OSF API client.

## Acceptance Criteria

- Preprint listing supports pagination, filtering by provider, and returns preprint metadata
- Search endpoint queries across OSF nodes, preprints, and registrations with a query string
- Registration creation submits a registration plan for an existing node
- Add-on listing returns configured storage add-ons for a given node
- All endpoints are covered by fixture-backed unit tests

## Non-Goals

- Preprint file upload or preprint creation
- Full-text search indexing
- Registration approval workflow management
- Add-on configuration or credential management
