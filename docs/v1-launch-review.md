# 1.0 Launch Review

Reviewed 2026-07-14. This is a readiness record, not a claim that `v1.0.0`
has been released.

| Gate | Evidence or waiver | Status |
|---|---|---|
| Stable contracts | Compatibility policy, CLI JSON contract, MCP schema contract, migration guide, and regression tests | ready locally |
| Reliability | Atomic writes, cancellation, pagination, retries, and race tests; resumable transfer is not yet implemented | partial; follow-up required |
| Security | Threat model, redaction tests, CodeQL, dependency review, govulncheck, SBOM, provenance, and Cosign workflows | ready locally; final review pending |
| Live behavior | Opt-in live validation matrix and tool; no credentials or disposable OSF project were configured for this review | waived pending live run |
| Quality | Go tests, race tests, vet, lint, anti-stub, review, registry, matrix, release-contract, and vulnerability gates | passed locally |
| Operations | Support policy and operations runbook; structured metrics/log export is not yet a release contract | partial; follow-up required |
| Documentation | Commands, usage, install, troubleshooting, architecture, migration, and support references | ready locally |
| Distribution | Cross-platform artifacts, signed provenance, SBOM, MCP registry, and plugin packages; several provider reviews remain external | partial; external gates remain |
| Governance | Compatibility, contribution, security, support, issue triage, and release policies are documented | ready locally |

## Decision

The repository is a stronger 1.0 release candidate, but not yet eligible for a
`v1.0.0` tag. The unresolved items are tracked in the feature matrix and
release evidence. This review does not waive live OSF testing, provider review,
resumable transfers, or structured observability; each requires a later dated
decision.
