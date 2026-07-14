# Repository provider contract

The provider domain contract is an internal, versioned vocabulary for comparing
concrete OSF and Zenodo behavior. It does not replace the existing OSF client and
does not claim that a Zenodo client exists yet.

## Design rules

- Every identity carries provider, native resource kind, and native identifier.
- Provider/kind mismatches are rejected; DOI, OSF GUID, Zenodo record ID, concept
  DOI, and version ID are never treated as interchangeable.
- Common lifecycle fields support workflows, while the original provider state
  and metadata bytes remain available losslessly.
- Permission, links, checksums, native version IDs, DOI, and concept DOI remain
  separate fields rather than being inferred from one another.
- The complete capability vocabulary receives an explicit `supported`,
  `partial`, or `unsupported` decision for each concrete provider.
- Partial and unsupported decisions carry constraints suitable for CLI/MCP error
  guidance. A partial capability is not silently promoted to full support.
- There is deliberately no generic network client interface. Later adapters may
  promote only operations whose concrete semantics satisfy this contract.

## Capability map

| Capability | OSF | Zenodo | Non-equivalence |
|---|---|---|---|
| `files.delete` | supported | partial | Zenodo deletion is draft-only; OSF behavior depends on storage authorization |
| `files.download` | supported | supported | Native links, access policy, and checksums remain provider-specific |
| `files.list` | supported | supported | OSF storage and Zenodo draft/published file shapes remain distinct |
| `files.upload` | supported | partial | Zenodo uses draft buckets and record limits |
| `lifecycle.publish` | partial | partial | OSF visibility/registration is not Zenodo DOI publication |
| `metadata.update` | supported | partial | Zenodo published updates require lifecycle actions |
| `oai.harvest` | unsupported | supported | OAI-PMH is a separate Zenodo adapter, not generic REST pagination |
| `records.create` | supported | partial | OSF creates projects/components; Zenodo creates draft depositions |
| `records.delete` | partial | partial | Entity state and withdrawal rules differ |
| `records.get` | supported | supported | OSF GUIDs and Zenodo record/deposition IDs retain native kinds |
| `records.search` | supported | supported | Query syntax, result shapes, and pagination remain native |
| `records.update` | supported | partial | Zenodo mutability depends on draft/edit state |
| `versions.create` | unsupported | partial | Zenodo versions require the latest published identity; OSF projects do not share this model |

## Validation

`internal/repository/conformancetest` is reusable by future provider
descriptors. It verifies model version, vocabulary completeness, ordering,
support resolution, and JSON round trips. Fixture cases exercise all three
support levels. Existing CLI and MCP compatibility fixtures continue to guard
the public OSF surfaces independently.
