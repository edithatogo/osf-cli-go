# Provider capabilities and lossless repository contracts

## Objective

Derive shared contracts from concrete OSF and Zenodo adapters without reducing
either provider to a misleading lowest-common-denominator model.

## Requirements

- Use typed provider-qualified identifiers and explicit capability negotiation.
- Preserve native metadata, links, permissions, checksums, versions, and states.
- Represent supported, partial, and unsupported operations explicitly.
- Keep the frozen OSF CLI, API, and MCP behavior compatible.
- Avoid a broad generic interface until concrete adapters prove each boundary.
- Provide a reusable provider conformance suite so future repositories must
  satisfy the same capability, identity, metadata, and error contracts.

## Completion evidence

Contract fixtures demonstrate lossless round trips, capability-aware failures,
and unchanged OSF compatibility behavior.
