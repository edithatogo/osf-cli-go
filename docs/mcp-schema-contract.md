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
| `osf_file_versions_list` | `id` |
| `osf_addons_list` | `id` |
| `osf_wikis_list` | `id` |
| `osf_comments_list` | `id` |
| `osf_logs_list` | `id` |
| `osf_identifiers_list` | `id` |
| `osf_contributors_list` | `id` |
| `osf_search` | `query`; optional `limit` |
| `osf_preprints_list` | optional `provider`, `limit` |
| `osf_preprints_search` | `query`; optional `provider`, `limit` |
| `osf_doi_resolve` | `identifier` |
| `repository_capabilities_get` | `provider` (`osf` or `zenodo`) |
| `zenodo_records_search` | optional `query`, `limit` |
| `zenodo_record_get` | `id` |
| `zenodo_files_list` | `id` |
| `zenodo_oai_records_list` | optional `metadataPrefix`, `set`, `from`, `until`, `resumptionToken` |
| `zenodo_oai_sets_list` | none |
| `zenodo_oai_formats_list` | optional `identifier` |

`internal/mcpserver/server_test.go` asserts the complete tool set and input
property names. The Zenodo tools expose public reads only and accept native,
provider-qualified, or canonical record identities. `nativeMetadataJson` and
`nativeMetadataXml` retain provider payloads as strings. Tool output fields may be extended compatibly; required input
properties are not renamed or removed in a minor release. Any future write
tool requires an explicit authorization and rollback contract before exposure.
No Zenodo upload, delete, update, version, or publish tool is registered.
