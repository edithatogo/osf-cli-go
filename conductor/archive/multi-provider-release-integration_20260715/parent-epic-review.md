# Parent Epic Review: Zenodo provider abstraction #101

- Reviewed: 2026-07-15
- Parent: [#101](https://github.com/edithatogo/osf-cli-go/issues/101)
- Integration child: [#111](https://github.com/edithatogo/osf-cli-go/issues/111)

## Native subissues

| Issue | Delivery | GitHub state at review | Conductor evidence |
|---|---|---|---|
| #102 | API provenance and drift | closed | `conductor/archive/zenodo-api-provenance_20260715/` |
| #103 | Provider contract | closed | `conductor/archive/repository-provider-contract_20260715/` |
| #104 | Read-only REST client | closed | `conductor/archive/zenodo-readonly-client_20260715/` |
| #105 | Provider-scoped CLI | closed | `conductor/archive/provider-scoped-cli_20260715/` |
| #106 | Provider-scoped MCP | closed | `conductor/archive/provider-scoped-mcp_20260715/` |
| #107 | OAI-PMH adapter | closed | `conductor/archive/zenodo-oai-pmh_20260715/` |
| #108 | Sandbox transfers | closed | `conductor/archive/zenodo-sandbox-transfers_20260715/` |
| #109 | Publication state | closed | `conductor/archive/zenodo-publication-state_20260715/` |
| #110 | Cross-provider copy | closed | `conductor/archive/cross-provider-provenance-transfer_20260715/` |
| #111 | Release integration | active closeout | this track and the dated provider report |

## Acceptance review

- Official Zenodo REST, sandbox, OAI-PMH, terms, and version-policy evidence is dated and drift-checked.
- The shared contract preserves provider-qualified IDs, native metadata, DOI/version semantics, permissions, licenses, checksums, publication state, and links.
- REST and OAI-PMH reads are fixture-backed and bounded; existing OSF contracts remain frozen.
- Public CLI and MCP surfaces expose provider-scoped reads and capability discovery only.
- Internal sandbox transfer, publication, and cross-provider state machines have dated live evidence, cleanup or retained-record dispositions, scoped revoked credentials, and failure recovery tests.
- CI and tagged release workflows gate API drift, provider contracts, observability, compatibility, claims, release reports, SBOM, and provenance.
- Registry metadata and the frozen MCP fixture advertise no Zenodo write tool.
- The generated feature matrix and validation report agree with these boundaries.

## Decision

The parent acceptance criteria are satisfied once #111 completes its clean
phase review and archive. The implementation is `offline-tested` for public
read contracts and `sandbox-validated` for the internal write workflows. It
makes zero production-validated claims and does not authorize production
Zenodo writes or public MCP write tools.
