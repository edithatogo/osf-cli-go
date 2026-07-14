# Phase Review: Paperclip Research Retrieval Parity

## Benchmark

- Reviewed the public `matsjfunke/paperclip` repository and README on 2026-07-14.
- Confirmed that the repository is archived, its hosted endpoint is unavailable, and its source uses Python FastMCP over HTTP.
- Compared Paperclip's multi-provider search, metadata retrieval, DOI/download URL handling, PDF-to-Markdown tools, tests, and deployment model with the OSF CLI Go contracts.

## Capability Decisions

- Existing OSF preprint discovery, DOI resolution, and explicit file/tree download workflows cover the maintainable OSF-specific capabilities.
- No production code or new offline tests were warranted because no safe accepted gap was identified.
- Arbitrary PDF URL retrieval and automatic PDF parsing were deferred with explicit provenance, media-type, size, parser, privacy, and opt-in requirements.

## Closeout

- Added the dated comparison to `docs/paperclip-research-retrieval-parity.md`.
- Updated the OSF tooling landscape and generated feature matrix.
- Issue #45 is reconciled with the source-backed decision and deferred retrieval contract.
