# Multi-provider CI, release, observability, and documentation integration

## Objective

Make capability claims, validation levels, operational evidence, docs, and
registry metadata agree before any multi-provider release.

## Requirements

- Gate offline contracts and API drift in CI and keep sandbox/live checks opt-in.
- Distinguish offline-tested, sandbox-validated, and production-validated claims.
- Emit provider-tagged, redacted, versioned operational events.
- Reconcile CLI/MCP fixtures, feature matrix, docs, release evidence, SBOM,
  provenance, and registry metadata.
- Record architecture decisions, threat-model deltas, performance budgets,
  sandbox-resource cleanup checks, and compatibility migration policy.

## Completion evidence

Release gates reject stale or unsupported claims and produce a dated,
reproducible multi-provider validation report.
