# Zenodo OAI-PMH harvesting adapter

## Objective

Support standards-based metadata harvesting without conflating OAI-PMH
resumption, schema, and rate-limit semantics with Zenodo REST discovery.

## Requirements

- Support sets, metadata prefixes, resumption tokens, expiry, cancellation,
  provider limits, and typed protocol errors.
- Preserve source provenance and native metadata alongside normalized fields.
- Expose OAI-PMH as a separate negotiated capability.

## Completion evidence

Offline fixtures cover pagination, token expiry, malformed XML, retries,
cancellation, metadata preservation, and deterministic recovery.
