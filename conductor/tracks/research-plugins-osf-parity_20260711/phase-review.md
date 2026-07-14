# Phase Review: Research Plugins OSF Parity

## Evidence and Gap Analysis

- Reviewed the public `wentorai/research-plugins` repository and README on 2026-07-14.
- Recorded the source-backed maturity and capability comparison in `docs/research-plugins-osf-parity.md`.
- Confirmed that OSF preprint discovery overlaps with existing `preprints list` and `preprints search` CLI and MCP surfaces.
- Classified general scholarly search and provider-specific open-access discovery as out of scope for an OSF client.

## Test-Driven Parity Work

- No accepted implementation gap was identified, so no new offline test or production code was warranted.
- Existing bounded search, explicit file download, authentication, provenance, and deterministic MCP/CLI tests remain the applicable contract.
- Documented the deferred full-text workflow and its provenance, media-type, size, and extraction requirements in the parity document and feature matrix.

## Validation and Closeout

- Updated `docs/osf-tooling-landscape.md` and generated `docs/feature-matrix.md` from `docs/feature-matrix.json`.
- Reconciled issue #16 with the source-backed findings and deferred contract.
- Validation commands and results are recorded in the implementation commit and final closeout note.
