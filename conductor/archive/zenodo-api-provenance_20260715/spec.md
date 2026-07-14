# Zenodo API provenance, version policy, and drift monitoring

## Objective

Establish a reproducible source-of-truth boundary for Zenodo REST, sandbox,
OAI-PMH, authentication, limits, and lifecycle behavior before implementation.

## Requirements

- Record authoritative URLs, retrieval dates, terms, API generation/version,
  schemas, endpoint evidence, and sandbox differences.
- Store a machine-checkable capability snapshot without copied credentials or
  unstable generated noise.
- Detect contract drift in CI and identify the affected source and remediation.
- Distinguish REST and OAI-PMH evidence and pin fixtures to their provenance.

## Completion evidence

Offline provenance and drift checks pass and the reviewed snapshot links every
implemented Zenodo capability to authoritative evidence.
