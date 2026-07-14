# Zenodo provider abstraction

GitHub issue: [#101](https://github.com/edithatogo/osf-cli-go/issues/101)

## Objective

Add an opt-in Zenodo provider and a provider-neutral research repository layer
that supports comparable OSF and Zenodo workflows without erasing provider-
specific identity, metadata, permission, transfer, DOI, and publication rules.

## Official API boundary

The initial source of truth is the [official Zenodo REST API documentation](https://developers.zenodo.org/).
The track must pin the reviewed source URL, retrieval date, API version or
ref when available, license/terms, and sandbox assumptions in repository
evidence. The documented surface includes depositions, published-record
search, file upload/download, DOI reservation and publication actions, OAuth
scopes, rate limits, and OAI-PMH discovery.

## Requirements

- Map Zenodo records, depositions, files, metadata, identifiers, publication
  state, authentication scopes, pagination, rate limits, and sandbox behavior.
- Define shared provider interfaces and typed domain models for records,
  artifacts, files, metadata, identifiers, transfers, and capabilities.
- Preserve provider-specific IDs, DOI/version semantics, permissions,
  licenses, checksums, publication state, and links.
- Implement fixture-backed read-only Zenodo discovery before any write path.
- Keep upload, DOI reservation, publish, discard, deletion, and mirroring
  behind explicit authorization, confirmation, and sandbox/live validation.
- Reuse the repository's redaction, retry, cancellation, resumable-transfer,
  observability, compatibility, and release-contract patterns.
- Add OSF/Zenodo capability mapping and document equivalent, partial, and
  provider-specific behavior in the feature matrix.

## Safety and non-goals

- No silent OSF-to-Zenodo mirroring or automatic publication.
- No assumption that an OSF project and Zenodo record share an identifier.
- No production Zenodo write claim without disposable sandbox evidence or an
  explicitly authorized live-validation decision.
- No project-local credential persistence; Zenodo tokens and OSF credentials
  remain environment or approved credential-store inputs.

## Completion evidence

- Official API source and mapping evidence are dated and reproducible.
- Offline provider contract tests pass without credentials.
- Sandbox tests cover writes, transfers, DOI actions, cleanup, and failure
  handling before production support is advertised.
- CLI/MCP compatibility fixtures and release checks prevent unsupported
  provider operations from being exposed as implemented.

## Child delivery tracks

| Order | Track | GitHub |
|---|---|---|
| 1 | [Zenodo API provenance](../../archive/zenodo-api-provenance_20260715/spec.md) | [#102](https://github.com/edithatogo/osf-cli-go/issues/102) |
| 2 | [Repository provider contract](../../archive/repository-provider-contract_20260715/spec.md) | [#103](https://github.com/edithatogo/osf-cli-go/issues/103) |
| 3 | [Zenodo read-only REST client](../../archive/zenodo-readonly-client_20260715/spec.md) | [#104](https://github.com/edithatogo/osf-cli-go/issues/104) |
| 4 | [Provider-scoped CLI](../../archive/provider-scoped-cli_20260715/spec.md) | [#105](https://github.com/edithatogo/osf-cli-go/issues/105) |
| 4 | [Provider-scoped MCP](../../archive/provider-scoped-mcp_20260715/spec.md) | [#106](https://github.com/edithatogo/osf-cli-go/issues/106) |
| 3 | [Zenodo OAI-PMH](../../archive/zenodo-oai-pmh_20260715/spec.md) | [#107](https://github.com/edithatogo/osf-cli-go/issues/107) |
| 5 | [Zenodo sandbox transfers](../zenodo-sandbox-transfers_20260715/spec.md) | [#108](https://github.com/edithatogo/osf-cli-go/issues/108) |
| 5 | [Zenodo publication state](../zenodo-publication-state_20260715/spec.md) | [#109](https://github.com/edithatogo/osf-cli-go/issues/109) |
| 6 | [Cross-provider provenance transfer](../cross-provider-provenance-transfer_20260715/spec.md) | [#110](https://github.com/edithatogo/osf-cli-go/issues/110) |
| 7 | [Multi-provider release integration](../multi-provider-release-integration_20260715/spec.md) | [#111](https://github.com/edithatogo/osf-cli-go/issues/111) |

The GitHub issues are native subissues of #101. The parent closes only after
each child has its own review evidence and the final integration gate passes.
