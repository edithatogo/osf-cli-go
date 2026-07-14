# Zenodo Provider Roadmap

This roadmap tracks a future provider-neutral repository layer for OSF and
Zenodo. It does not claim that Zenodo support is implemented.

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

1. Pin and map the official Zenodo API and sandbox.
2. Define shared domain and capability contracts.
3. Add offline Zenodo read-only fixtures without changing OSF behavior.
4. Add explicit CLI/MCP discovery only after compatibility review.
5. Validate sandbox transfers and publication actions with cleanup evidence.
6. Consider production write support and optional cross-provider transfer only
   after safety, provenance, and authorization review.

The implementation track and issue are [#101](https://github.com/edithatogo/osf-cli-go/issues/101).
