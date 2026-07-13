# Datahugger DOI comparison

Reviewed: 2026-07-13

## Evidence

- Repository: [J535D165/datahugger](https://github.com/J535D165/datahugger)
- User workflow: [README.md](https://github.com/J535D165/datahugger/blob/main/README.md)
- DOI parsing/resolution: [handles.py](https://github.com/J535D165/datahugger/blob/main/datahugger/handles.py)
- OSF downloader: [services.py](https://github.com/J535D165/datahugger/blob/main/datahugger/services.py)
- GitHub metadata checked on 2026-07-13: default branch `main`, latest push 2026-06-29, 89 stars, 19 open issues, and MIT license metadata.
- Datahugger advertises support for more than 377 repositories, including OSF, and its latest listed release is `v0.13` from 2024-09-30. It has benchmark, documentation, package-publish, and GitHub Pages workflows.

## Capability comparison

| Capability | Datahugger | OSF CLI Go | Decision |
|---|---|---|---|
| DOI parsing | Accepts bare DOI, `doi:` prefix, and DOI URLs | New strict DOI parser accepts those forms | Implemented |
| DOI resolution | Follows DOI redirects to identify the repository | `osf resolve` and `osf_doi_resolve` resolve and require an OSF destination | Implemented for OSF-safe resolution |
| OSF file retrieval | Resolves OSF datasets and downloads files, including recursive folders | Existing `files download`, tree downloads, manifests, conflict policy, and path protection | Covered with a stronger local-write safety contract |
| Cross-repository dispatch | Routes DOI/URL resources across hundreds of repositories | Intentionally OSF-only | Reject broad dispatch; it would expand scope and provider-specific maintenance without improving OSF correctness |
| Metadata and dry-run | Exposes metadata and print-only download mode | Structured JSON export and read-only MCP tools; no DOI download side effect | Covered for OSF discovery; download planning remains a separate bounded workflow |
| Checksums and size limits | Optional checksum and maximum-file-size download controls | Safe streaming and manifests, with typed file metadata | Existing safety model exceeds the minimum; broader checksum policy remains separate |
| Tests and release maturity | Active Python project with package, docs, and benchmark workflows | Deterministic Go tests, race/vet/lint/vulnerability/registry/release gates | Both have mature evidence in different layers |

## Scope result

The material OSF-specific gap was DOI-to-OSF resolution. It is implemented as
a network opt-in resolver with strict input validation, redirect handling,
non-OSF destination rejection, deterministic offline tests, a CLI command, and
the read-only MCP tool `osf_doi_resolve`. Datahugger's multi-repository
dispatch, download planning, and provider-specific metadata remain outside
this track and are not copied into the OSF client.
