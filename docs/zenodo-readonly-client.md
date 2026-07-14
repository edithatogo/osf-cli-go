# Zenodo read-only REST client

`internal/zenodoapi` is the first concrete consumer of the provider contract.
It supports published-record search, record retrieval, and file discovery. It
does not yet add CLI or MCP commands and contains no deposition, upload,
publication, deletion, or OAI-PMH behavior.

## Configuration and authentication

- Production default: `https://zenodo.org/api/`.
- Approved alternative: `https://sandbox.zenodo.org/api/`.
- Plain HTTP is accepted only for localhost tests.
- Public record reads work without credentials.
- An optional token is sent only as `Authorization: Bearer`; URL credentials,
  token query parameters, foreign hosts, and cross-origin redirects are rejected.

The later CLI/MCP tracks will map environment variables into this constructor.
This internal track does not persist credentials or change existing OSF auth.

## Reliability and safety

- Response bytes, page count, simultaneous requests, retry count, and retry
  delay are bounded.
- Automatic retries apply only to idempotent GET requests and retryable network,
  `429`, `502`, `503`, and `504` failures.
- Pagination and redirects must remain under the configured API origin and path.
- Context cancellation interrupts semaphore waits, requests, and retry delays.
- `X-RateLimit-*` and `Retry-After` values are available as typed state.
- API errors retain method, path, status, rate-limit state, and a redacted bounded
  message. Queries and tokens are excluded from errors and operational events.
- Observability uses the existing versioned event schema with
  `provider=zenodo` and a low-cardinality `zenodo_api` endpoint class.

## Data fidelity

Typed fields cover identifiers, DOI/concept DOI, discovery metadata, creators,
access, licenses, files, checksums, and links. The original record JSON is also
retained and converted into the lossless provider envelope. Both current
`files.entries` and documented legacy file-array records are accepted.

Synthetic fixtures are tied to the dated provenance manifest and deliberately
include unknown metadata to prove it survives conversion. Routine tests do not
refresh fixtures from live user records.

## Validation status

The package is offline-tested, including pagination, limits, current/legacy
shapes, redaction, rate limits, retries, malformed responses, cancellation,
concurrency, redirects, fuzz seeds, and native metadata preservation. Targeted
statement coverage at the track review is 90.8%. A direct public API probe on
2026-07-15 returned a non-JSON intermediary response, so no live API claim or
fixture was derived from it.
