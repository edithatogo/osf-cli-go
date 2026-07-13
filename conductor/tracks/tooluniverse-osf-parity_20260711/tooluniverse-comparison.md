# ToolUniverse Comparison

Reviewed 2026-07-13 against the public `main` branch of
[mims-harvard/ToolUniverse](https://github.com/mims-harvard/ToolUniverse).

## Dated Evidence

| Dimension | Evidence | Assessment |
|---|---|---|
| OSF capability | [`osf_preprints_tool.py`](https://github.com/mims-harvard/ToolUniverse/blob/main/src/tooluniverse/osf_preprints_tool.py) and [`OSF_search_preprints.py`](https://github.com/mims-harvard/ToolUniverse/blob/main/src/tooluniverse/tools/OSF_search_preprints.py) use the OSF API v2 preprints endpoint with title and provider filters | Material OSF gap is dedicated preprint search, not general agent orchestration |
| Result contract | [`osf_preprints_tools.json`](https://github.com/mims-harvard/ToolUniverse/blob/main/src/tooluniverse/data/osf_preprints_tools.json) specifies query, provider, max 100 results, title, date, published state, DOI, URL, and source | Implemented in Go API, CLI, and MCP with bounded read-only output and source provenance |
| Automation surface | README documents the `tu` CLI, Python SDK, native MCP server, compact mode, tool composition, caching, and agent skills | General ToolUniverse runtime is intentionally not copied; this repository remains an OSF-focused CLI/MCP server |
| Authentication | ToolUniverse's OSF preprint example is public API access and does not require a token | Go implementation preserves anonymous public discovery and existing optional credentials |
| Tests and quality | Repository currently exposes a large Python test/integration surface and active CI; GitHub metadata on 2026-07-13 showed 1,564 stars, 601 commits, Apache-2.0, and a v1.3.1 release published 2026-07-02 | ToolUniverse has broader ecosystem maturity; Go parity is limited to the material OSF contract and is covered by offline tests, race tests, vet, lint, and registry gates |

## Adopted Contract

- `osf preprints search <query> [--provider <provider>] [--limit 1-100]`
- `osf_preprints_search` MCP tool with the same query, provider, and limit inputs
- API method `SearchPreprints`, using OSF `filter[title]` and `filter[provider]`
- Stable fields: title, publication date, published state, DOI, OSF HTML URL,
  and `OSF Preprints` source provenance
- No writes, downloads, credential persistence, or network-dependent unit tests

## Deferred or Rejected Scope

ToolUniverse's general-purpose AI-Tool Interaction Protocol, 1,000-plus-tool
registry, asynchronous task framework, composition engine, compact mode, cache
system, literature search federation, and agent skills are outside this
OSF-focused product boundary. Reimplementing them would add a second agent
runtime rather than improve OSF access. Future research-agent parity belongs in
the dedicated `research-plugins-osf-parity` track and should be evaluated there
with separate acceptance criteria.
