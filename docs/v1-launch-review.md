# 1.0 Launch Review

Reviewed 2026-07-14. This is a readiness record, not a claim that `v1.0.0`
has been released.

| Gate | Evidence or waiver | Status |
|---|---|---|
| Stable contracts | Compatibility policy, CLI JSON contract, MCP schema contract, migration guide, and regression tests | ready locally |
| Reliability | Atomic writes, cancellation, pagination, retries, and race tests; resumable transfer is tracked in [#95](https://github.com/edithatogo/osf-cli-go/issues/95) | partial; follow-up required |
| Security | Threat model, redaction tests, CodeQL, dependency review, govulncheck, SBOM, provenance, and Cosign workflows | ready locally; final review pending |
| Live behavior | Opt-in live validation matrix and tool; release-candidate execution is tracked in [#97](https://github.com/edithatogo/osf-cli-go/issues/97) | blocked 2026-07-14: `OSF_TOKEN` or username/password and `OSF_VALIDATE_PROJECT` are not present in the validation environment |
| Quality | Go tests, race tests, vet, lint, anti-stub, review, registry, matrix, release-contract, and vulnerability gates | passed locally |
| Operations | Support policy and operations runbook; structured observability is tracked in [#96](https://github.com/edithatogo/osf-cli-go/issues/96) | partial; follow-up required |
| Documentation | Commands, usage, install, troubleshooting, architecture, migration, and support references | ready locally |
| Distribution | Cross-platform artifacts, signed provenance, SBOM, MCP registry, and plugin packages; clean release-candidate verification is tracked in [#98](https://github.com/edithatogo/osf-cli-go/issues/98) | partial; external gates remain |
| Governance | Compatibility, contribution, security, support, issue triage, and release policies are documented | ready locally |

## Decision

The repository is a stronger 1.0 release candidate, but not yet eligible for a
`v1.0.0` tag. The remaining local gates are tracked in [#95](https://github.com/edithatogo/osf-cli-go/issues/95), [#96](https://github.com/edithatogo/osf-cli-go/issues/96), [#97](https://github.com/edithatogo/osf-cli-go/issues/97), [#98](https://github.com/edithatogo/osf-cli-go/issues/98), and [#99](https://github.com/edithatogo/osf-cli-go/issues/99). This review does not waive live OSF testing, provider review, resumable transfers, structured observability, or compatibility verification; each requires a dated decision.
