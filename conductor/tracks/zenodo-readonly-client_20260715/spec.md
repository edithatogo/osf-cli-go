# Fixture-backed read-only Zenodo REST client

## Objective

Implement records search, record inspection, and file discovery behind the
reviewed provider contract before exposing writes or OAI-PMH harvesting.

## Requirements

- Support public unauthenticated reads, pagination, cancellation, rate-limit
  signals, typed errors, and redacted authenticated failures.
- Use official-shape fixtures with dated provenance and no routine live network.
- Emit provider-tagged structured events without tokens or sensitive metadata.
- Fuzz parsers and pagination boundaries and enforce bounded response, retry,
  and concurrency budgets against malformed or unexpectedly large responses.

## Completion evidence

Offline tests cover success, pagination, malformed responses, retries,
cancellation, limits, and redaction across supported read operations.
