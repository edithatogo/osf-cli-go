# Glama Submission Evidence

Date: 2026-07-11

## Discovery Audit

- Search URL: `https://glama.ai/mcp/servers?query=osf-cli-go`
- Glama reported 53,995 indexed servers at the time of submission.
- No result matched `OSF CLI Go`, `edithatogo/osf-cli-go`, or the official registry server identity.
- The visible OSF connector belonged to `io.github.pipeworx-io` and is unrelated.
- No listing URL, install command, quality grade, maintenance grade, or score existed for this repository.

## Authenticated Submission

The `Add Server` action required a Glama account. After user login, Chrome submitted the open-source server form with:

- Name: `OSF CLI Go`
- Description: `Read-only MCP tools for authenticated Open Science Framework projects, components, files, and contributors.`
- Repository: `https://github.com/edithatogo/osf-cli-go`
- Submission type: `Server`, not `Connector`, because this is a repository-backed stdio MCP server rather than a public remote endpoint.

The form states that public submissions are reviewed before becoming visible. Submission closed the dialog without an error and returned to the server search. The listing was not immediately public, so the current status is `pending_review`.

## Score Boundary

Glama provides license, quality, and maintenance grades after indexing. No grades or warnings exist before the listing is created, so no score-driven repository change can be justified yet. Existing canonical metadata, license, repository, tool schemas, auth declarations, and package identity remain covered by `go run ./tools/checkregistries`.

The external follow-up is Glama review/indexing, followed by a fresh score audit and claim check once a public listing URL exists.
