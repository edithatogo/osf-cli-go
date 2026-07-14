# Paperclip research retrieval parity

Last reviewed: 2026-07-14

This comparison uses the archived [matsjfunke/paperclip repository](https://github.com/matsjfunke/paperclip), its [README](https://github.com/matsjfunke/paperclip/blob/main/README.md), source, and tests. Paperclip is evaluated as a historical MCP research-retrieval reference; its hosted service is not treated as available.

## Source maturity

| Signal | Paperclip | OSF CLI Go |
|---|---|---|
| License | MIT | Apache-2.0 |
| Maintenance | GitHub repository archived 2025-12-16; 63 commits, 28 stars, and 10 forks in the reviewed snapshot | Active repository with release, security, registry, anti-stub, test, race, and vet gates |
| Runtime and transport | Python FastMCP HTTP server bound to `0.0.0.0:8000`; README says the hosted endpoint is unavailable and users must self-host | Go CLI plus stdio MCP server with versioned package and release artifacts |
| Provider scope | arXiv, OpenAlex, OSF, and many OSF-hosted preprint providers | OSF API and OSF Storage, with bounded OSF preprint and search surfaces |
| Retrieval tools | Multi-provider search, metadata by ID, paper-by-ID PDF retrieval, and direct PDF URL retrieval converted to Markdown | Structured OSF metadata, DOI-to-OSF resolution, explicit file/tree download, and deterministic output |
| Tests | Public metadata and PDF retrieval tests use live-looking provider identifiers and expected remote content | Offline HTTP fixtures and deterministic CLI/MCP tests; live OSF validation remains opt-in |

## Capability comparison

| Capability | Paperclip reference | OSF CLI Go behavior | Decision |
|---|---|---|---|
| Multi-provider discovery | One `search_papers` tool combines arXiv, OpenAlex, and OSF/preprint-provider results | OSF search and bounded OSF preprint search | Deliberately out of scope; adding unrelated scholarly providers would expand the OSF client boundary |
| OSF metadata retrieval | `get_paper_metadata_by_id` retrieves OSF metadata by preprint ID | `preprints list/search`, structured search output, and `osf_preprints_list/search` | Implemented |
| DOI and URL resolution | Metadata includes provider download URLs but no OSF-host validation contract | `resolve` and `osf_doi_resolve` validate DOI forms and reject non-OSF destinations without downloading | Implemented with stronger OSF-specific safety |
| Full-text retrieval | `get_paper_by_id` downloads and parses PDFs; `get_paper_content_by_url` accepts arbitrary PDF URLs | Explicit OSF file/tree download only; no implicit PDF retrieval or extraction | Deferred; issue #45 requires an opt-in provenance, media-type, size, parser, and privacy contract |
| Authentication and privacy | README documents an unauthenticated self-hosted HTTP endpoint | OSF token or username/password auth with redaction; private files require explicit authenticated access | Implemented for OSF |
| Automation and transport | HTTP MCP endpoint and broad provider tool surface | CLI-first JSON contracts and stdio MCP tools with bounded inputs | Implemented for OSF workflows |

## Deferred retrieval contract

Paperclip demonstrates useful retrieval ergonomics, but importing arbitrary URL PDF
fetching or automatic OSF PDF parsing would create uncontrolled network, content-size,
media-type, parser, and private-file risks. A future OSF full-text feature must first
define an explicit file-selection workflow, preserve OSF and DOI provenance, enforce
bounded downloads, validate content types, define parser behavior and failure modes,
and prevent credential or private-content leakage. Until then, users can discover
metadata and request a known OSF file or tree through the existing explicit download
path.

No Paperclip code was copied, no hosted-service availability is claimed, and no
network-dependent tests or credentials were added.
