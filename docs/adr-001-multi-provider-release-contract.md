# ADR 001: Multi-provider release contract

- Status: accepted
- Date: 2026-07-15
- Decision owners: repository maintainers

## Context

OSF CLI Go now has a shared repository contract, public OSF and Zenodo read
surfaces, and internal sandbox-only Zenodo write harnesses. A release must not
turn fixture coverage or sandbox execution into an unsupported production
claim, expose internal write machinery through MCP, or lose the relationship
between evidence and the source revision that produced an artifact.

## Decision

1. Provider capabilities are reported at one of three levels:
   `offline-tested`, `sandbox-validated`, or `production-validated`.
2. Every claim is dated and bound to repository evidence by SHA-256. Production
   claims additionally require a public HTTPS production record.
3. Sandbox claims record whether resources were deleted or, for irreversible
   publication, retained at an explicit public sandbox URL.
4. Provider writes remain internal validation capabilities until a separate
   compatibility and safety review approves a public CLI or MCP surface.
5. OSF, Zenodo, and cross-provider events use the versioned redacted
   observability envelope and a normalized top-level provider field.
6. Tagged artifacts are gated by the provider claim checker. The generated
   report is released beside binaries, while SBOM and provenance generation is
   bound to the same validated source revision.
7. Registry descriptions advertise only the public read-only MCP surface.

## Consequences

- Sandbox validation can increase confidence without implying production use.
- A stale file, digest, resource disposition, workflow boundary, or provider
  claim fails CI or the release job.
- Irreversible sandbox records remain visible and attributable rather than
  being falsely reported as cleaned up.
- New provider writes require additive compatibility fixtures, migration notes,
  threat-model review, and a separately approved public interface.

## Alternatives rejected

- A single `live-validated` label was rejected because it conflates sandbox and
  production behavior.
- Free-form release notes were rejected because they cannot detect digest or
  claim drift.
- Automatic scheduled write validation was rejected because it increases
  credential and orphan-resource risk.
