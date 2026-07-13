# MCP Schema Contract

The MCP server exposes read-only tools whose names and required input
properties are stable within a major version.

The current contract includes:

| Tool | Required inputs |
|---|---|
| `osf_whoami` | none |
| `osf_projects_list` | none |
| `osf_project_get` | `id` |
| `osf_components_list` | `id` |
| `osf_files_list` | `id`; optional `path` |
| `osf_contributors_list` | `id` |
| `osf_search` | `query`; optional `limit` |
| `osf_preprints_list` | optional `provider`, `limit` |
| `osf_preprints_search` | `query`; optional `provider`, `limit` |
| `osf_doi_resolve` | `identifier` |

`internal/mcpserver/server_test.go` asserts the complete tool set and input
property names. Tool output fields may be extended compatibly; required input
properties are not renamed or removed in a minor release. Any future write
tool requires an explicit authorization and rollback contract before exposure.
