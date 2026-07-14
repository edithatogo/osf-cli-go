# Research plugins OSF parity

Last reviewed: 2026-07-14

This comparison uses the public [wentorai/research-plugins repository](https://github.com/wentorai/research-plugins), its [README](https://github.com/wentorai/research-plugins/blob/main/README.md), and the OSF API contract. It evaluates OSF-specific behavior separately from the project's much broader scholarly-research skills.

## Source maturity

| Signal | research-plugins | OSF CLI Go |
|---|---|---|
| Primary language | TypeScript/JavaScript agent plugin | Go CLI and stdio MCP server |
| License | MIT for original project content; third-party sources retain their licenses | Apache-2.0 |
| Distribution | OpenClaw plugin plus Agent Skills installation | Versioned CLI binaries, MCPB, OCI image, and coding-agent packages |
| Scope | 433 skills, 34 tools, 18 scholarly database modules, and 40+ agent-framework integrations | OSF API client, safe file operations, deterministic research validation, and MCP tools |
| OSF-specific surface | `search_osf_preprints` in the OSF Preprints module; broader OA/full-text tools use other scholarly providers | OSF preprint list/search, OSF search, metadata export, and explicit file download |
| Current source signals | Public repository snapshot shows 259 stars, 42 forks, and latest release v1.4.8 dated 2026-06-19 | Active CI, offline tests, race/vet/security/release gates, and release artifacts |

## Capability comparison

| Capability | research-plugins reference | OSF CLI Go behavior | Decision |
|---|---|---|---|
| OSF preprint discovery | `search_osf_preprints` | `osf preprints search` and `osf_preprints_search` with query/provider/limit bounds | Implemented |
| General scholarly discovery | OpenAlex, Crossref, arXiv, PubMed, Europe PMC, OpenCitations, and other modules | OSF search and preprint discovery only | Deliberately out of scope; use dedicated scholarly integrations rather than expanding the OSF client |
| Open-access/full-text discovery | `find_oa_version` and provider-specific article tools | Explicit OSF file/tree download when a file is identified; no implicit PDF retrieval | Deferred; issue #16 tracks an opt-in provenance and download contract |
| Authentication | Provider-specific public APIs; README advertises no-key scholarly APIs | `OSF_TOKEN` preferred, username/password fallback, redacted errors | Implemented for OSF |
| Agent automation | 433 progressive-loading `SKILL.md` files and OpenClaw/MCP-style tools | JSON CLI contract, shell completion, MCP schemas, and deterministic output | Implemented for OSF workflows |
| Tests and maintenance | Public tests/tools directories and active release cadence | Offline fixtures, unit/race tests, vet, anti-stub, registry, release, and MCP-quality gates | Implemented |

## Deferred full-text contract

Automatic full-text retrieval is not added to the OSF CLI or MCP server. A safe
implementation would need to distinguish OSF files from external OA links,
preserve DOI/provider provenance, enforce content-size and media-type limits,
avoid surprise downloads, define local extraction boundaries, and prevent
credential or private-file leakage. The current supported workflow is explicit:
discover an OSF preprint, inspect its metadata, then request a known OSF file or
tree download with the existing conflict and path-safety controls.

The deferred capability is tracked in [issue #16](https://github.com/edithatogo/osf-cli-go/issues/16) and the maintained [feature matrix](feature-matrix.md). No credentials, synthetic usage, or external full-text access claims are committed.
