# OSF API Coverage Extension

## Objective

Research all existing OSF tools (osfclient Python, osfr R, osf-project-exporter) and OSF API v2 docs, then extend the osfapi client to match or exceed their endpoint coverage.

## Acceptance Criteria

- Research all OSF API v2 endpoints documented at developer.osf.io
- Map endpoints: what osfclient, osfr, osf-project-exporter support vs what we support
- Implement missing read-only endpoints: registrations, wikis, comments, logs, identifiers
- Implement write endpoints: create/update/delete nodes, upload files
- Add CLI commands for new endpoints
- All new endpoints have fixture-backed tests
- Coverage requirement: each endpoint handler has >= 85% test coverage

## Non-Goals

- Authentication changes or OAuth flow modification
- OSF preprints or OSF Meetings API support
- Batch/bulk upload operations
