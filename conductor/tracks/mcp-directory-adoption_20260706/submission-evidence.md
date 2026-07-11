# MCP.Directory Submission Evidence

Date: 2026-07-11

## Current Requirements

- Submission URL: `https://mcp.directory/submit`
- Required field: public GitHub repository URL.
- Optional fields: npm package, PyPI package, description (100 characters), and email.
- MCP.Directory auto-pulls repository metadata and states that review normally occurs within 24 hours.
- Entries ingested from the Official MCP Registry can be claimed by emailing `hello@mcp.directory`.

## Submission Result

Chrome submission used:

- Repository: `https://github.com/edithatogo/osf-cli-go`
- Description: `Read-only MCP tools for Open Science Framework projects, components, files, and contributors.`
- npm, PyPI, and email: omitted because they are not required and do not describe the OCI/Go distribution route.

The form returned: `This repository has already been submitted. We'll review it soon!`

This proves a pending submission, not a published listing. No public listing URL or score was exposed, so no score can currently be improved. The external gate is MCP.Directory review and indexing.

## Package Evidence

- Official MCP Registry version: `0.2.0`.
- OCI package: `ghcr.io/edithatogo/osf-cli-go-osf-mcp:0.2.0`.
- Git tag `v0.2.0` exists locally and on GitHub.
- Registry publication workflow run `26027015142` completed successfully.
- GitHub does not currently expose a Release object for `v0.2.0`; this does not block MCP.Directory's repository-only form but does block robust package-manager adoption.
