# Zenodo Provider Roadmap

This roadmap tracks the staged provider-neutral repository layer for OSF and
Zenodo. The provider contract, public REST reads, and public OAI-PMH harvesting
are offline-tested; write and cross-provider workflows remain gated.

## API mapping

| Capability | OSF | Zenodo | Shared contract decision |
|---|---|---|---|
| Research container | Project/component/node | Published record/deposition | Preserve provider type and native ID; expose a common record envelope |
| Discovery | OSF search and node listings | Records search and OAI-PMH | Read-only first; provider query and pagination semantics remain visible |
| Files | WaterButler/storage files | Deposition bucket/files API | Shared transfer interface with provider-specific limits, checksums, and links |
| Metadata | Node/project metadata and contributors | Deposition metadata, creators, keywords, related identifiers | Map common fields but retain lossless provider metadata |
| Persistent identity | OSF URLs and DOI-related resources | DOI reservation, published DOI, version/record IDs | Never treat GUIDs, record IDs, and DOIs as interchangeable |
| Publication | OSF project visibility and registration workflows | Deposition publish/discard/new-version actions | Publication is an explicit provider capability requiring authorization |
| Authentication | `OSF_TOKEN` or documented fallback | Zenodo personal/OAuth token with scoped permissions | Environment/credential-store boundary; never project-local secrets |
| Rate limits | OSF response and retry behavior | Zenodo global/search/OAI limits and rate headers | Shared retry instrumentation with provider-specific budgets |

## Delivery sequence

| Stage | Child issue and track | Exit condition |
|---|---|---|
| 1 | [#102 API provenance](https://github.com/edithatogo/osf-cli-go/issues/102), `conductor/archive/zenodo-api-provenance_20260715/` | Authoritative sources, version policy, fixtures, and drift gate are reproducible |
| 2 | [#103 provider contract](https://github.com/edithatogo/osf-cli-go/issues/103), `conductor/archive/repository-provider-contract_20260715/` | Qualified IDs, capability negotiation, lossless metadata, and OSF compatibility pass |
| 3 | [#104 REST client](https://github.com/edithatogo/osf-cli-go/issues/104), `conductor/archive/zenodo-readonly-client_20260715/`; [#107 OAI-PMH](https://github.com/edithatogo/osf-cli-go/issues/107), `conductor/archive/zenodo-oai-pmh_20260715/` | Independent read adapters pass offline contract, fuzz, limit, and redaction tests |
| 4 | [#105 CLI](https://github.com/edithatogo/osf-cli-go/issues/105), `conductor/archive/provider-scoped-cli_20260715/`; [#106 MCP](https://github.com/edithatogo/osf-cli-go/issues/106), `conductor/archive/provider-scoped-mcp_20260715/` | Provider-scoped read UX is compatible and no deferred writes are advertised |
| 5 | [#108 transfers](https://github.com/edithatogo/osf-cli-go/issues/108), `conductor/archive/zenodo-sandbox-transfers_20260715/`; [#109 publication](https://github.com/edithatogo/osf-cli-go/issues/109), `conductor/tracks/zenodo-publication-state_20260715/` | Disposable sandbox evidence proves integrity, cleanup, and irreversible-state safety |
| 6 | [#110 cross-provider copy](https://github.com/edithatogo/osf-cli-go/issues/110), `conductor/tracks/cross-provider-provenance-transfer_20260715/` | Dry-run, provenance, idempotency, compensation, and failure recovery pass |
| 7 | [#111 release integration](https://github.com/edithatogo/osf-cli-go/issues/111), `conductor/tracks/multi-provider-release-integration_20260715/` | CI, release evidence, observability, docs, matrix, and registry claims agree |

## Architecture guardrails

- Start with concrete OSF and Zenodo adapters. Promote an operation into the
  shared interface only after both implementations demonstrate coherent
  semantics.
- Use provider-qualified identifiers such as provider plus native record ID;
  never infer equivalence from a DOI, URL, or similarly formatted value.
- Normalize common fields for workflows while retaining a lossless native
  metadata envelope and source provenance.
- Negotiate capabilities at runtime and represent partial or unsupported
  behavior explicitly in CLI, API, MCP, and JSON contracts.
- Keep OAI-PMH separate from REST because its schemas, resumption tokens,
  expiry, and rate limits form a different protocol contract.
- Model publication as an irreversible state machine with explicit access,
  embargo, license, scope, validation, confirmation, and audit requirements.
- Model cross-provider writes as idempotent sagas with mapping previews,
  checkpoints, partial results, compensation, and transformation provenance.
- Use separate `ZENODO_TOKEN` and `ZENODO_BASE_URL` inputs and never reuse,
  persist, print, or infer OSF credentials.
- Defer a product or binary rename. Provider-scoped commands preserve the 1.0
  OSF compatibility contract while real multi-provider usage is evaluated.

## Additional maturity controls

- Reusable provider conformance tests for capabilities, identity, metadata,
  pagination, errors, cancellation, observability, and validation levels.
- Parser fuzzing and bounded response, retry, concurrency, and rate budgets.
- Architecture decision records for public contract and dependency choices.
- Threat-model updates for token scopes, private records, metadata disclosure,
  confused-deputy writes, and cross-provider data residency.
- A sandbox janitor that identifies and cleans stale disposable resources
  without broad deletion permissions.
- Performance and reliability baselines for large records and file sets.
- Explicit schema evolution, deprecation, and migration rules before provider
  fields enter stable JSON or MCP contracts.

These controls belong to the ten child tracks rather than new standalone
issues: API drift is owned by #102, conformance by #103, fuzzing and budgets by
#104, lifecycle policy by #109, recovery by #110, and release-wide architecture,
security, performance, and cleanup evidence by #111.

The parent implementation epic is [#101](https://github.com/edithatogo/osf-cli-go/issues/101),
with all ten delivery issues attached as native GitHub subissues.
