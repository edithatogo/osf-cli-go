# OSF API Client And Fixtures

## Objective

Build a typed, fixture-backed OSF API v2 client that handles authentication headers, JSON:API pagination, relationship traversal, and status-aware errors.

## Acceptance Criteria

- Unit tests use `httptest.Server`; routine tests do not call live OSF.
- Fixture contracts cover current user, projects/nodes, components, contributors, OSF Storage files, pagination, download links, and error responses.
- Client errors preserve status code, endpoint, and OSF error details.
