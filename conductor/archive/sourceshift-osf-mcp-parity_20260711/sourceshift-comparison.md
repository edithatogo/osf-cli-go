# SourceShift OSF MCP comparison

Reviewed: 2026-07-13

## Evidence

- Repository: [SourceShift/osf-mcp-server](https://github.com/SourceShift/osf-mcp-server)
- README capability reference: [README.md](https://github.com/SourceShift/osf-mcp-server/blob/main/README.md)
- Project metadata checked on 2026-07-13: default branch `main`, latest push
  2025-12-27, one star, zero open issues, and no SPDX license detected by the
  GitHub repository API. The README states MIT, but this repository does not
  treat that statement as a substitute for a machine-readable upstream
  license file.
- Upstream dependencies and test commands are documented in
  [pyproject.toml](https://github.com/SourceShift/osf-mcp-server/blob/main/pyproject.toml).

## Capability comparison

| Capability | SourceShift | OSF CLI Go | Decision |
|---|---|---|---|
| Project/search discovery | `search_projects`, plus project lookup | `osf_search`, `osf_project_get` | Implemented in MCP parity work |
| Registration search | README advertises registration search | CLI API coverage includes registration listing and creation; MCP remains read-only | Defer to a dedicated MCP registration-read track |
| Preprint search/listing | `search_preprints` | `osf_preprints_list` | Implemented in MCP parity work |
| Project metadata writes | `create_project`, `update_project` | CLI supports writes; MCP is intentionally read-only | Rejected for this track; require explicit MCP write-safety design |
| File listing | `list_files` | `osf_files_list` | Already covered |
| File download | `download_file` | CLI has safe streaming download; MCP exposes metadata only | Defer until MCP download boundary and resource limits are approved |
| Cached file reading and PDF conversion | `read_file` and `pymupdf4llm` | No equivalent MCP tool | Defer to a content-extraction track; avoid implicit local caching in a stdio server |
| Wiki access | `get_wiki` | API client supports wiki listing; no MCP tool | Defer to read-only entity expansion |
| DOI resolution | `resolve_doi` | No equivalent | Track under DOI-oriented parity issue [#14](https://github.com/edithatogo/osf-cli-go/issues/14) |
| Authentication | Optional `OSF_TOKEN`; public-only mode | Token and username/password fallback with redaction | OSF CLI Go exceeds upstream authentication handling |
| Transfer safety | Local storage path and downloads | Path traversal protection, conflict policy, manifests, and streaming downloads | OSF CLI Go exceeds upstream safety contract |
| Tests and release maturity | Pytest/coverage commands; latest upstream push 2025-12-27 | Deterministic Go tests, race/vet/lint/vulnerability/release/registry gates, signed release artifacts | OSF CLI Go exceeds upstream maturity evidence |

## Scope result

The material low-risk MCP gaps were search and preprint listing. They are now
implemented as `osf_search` and `osf_preprints_list`, with bounded result
limits, input validation, structured output, offline tests, and synchronized
MCPB/registry metadata.

The remaining differences are intentionally not hidden as parity failures:
MCP write tools need an explicit confirmation and authorization model; file
content extraction needs resource limits and a cache contract; DOI resolution
belongs with the existing DOI-oriented roadmap; and wiki/registration tools
are separate read-only entity expansions.
