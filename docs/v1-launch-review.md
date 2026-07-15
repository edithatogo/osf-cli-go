# 1.0 Launch Review

Reviewed 2026-07-15. This is a readiness record, not a claim that `v1.0.0`
has been released.

| Gate | Evidence or waiver | Status |
|---|---|---|
| Stable contracts | Compatibility policy, pinned OSF source manifest, CLI/MCP golden fixtures, migration guide, and CI compatibility gate | ready locally |
| Reliability | Atomic writes, cancellation, pagination, retries, and race tests; resumable transfer is tracked in [#95](https://github.com/edithatogo/osf-cli-go/issues/95) | partial; follow-up required |
| Security | Threat model, redaction tests, CodeQL, dependency review, govulncheck, SBOM, provenance, and Cosign workflows | ready locally; final review pending |
| Live behavior | Opt-in live validation matrix and tool; dated sanitized evidence in `docs/live-osf-validation-evidence.md` | passed 2026-07-15: authentication, reads, transfers, conflict handling, cancellation, MCP, and cleanup passed against a disposable private OSF project |
| Quality | Go tests, race tests, vet, lint, anti-stub, review, registry, matrix, release-contract, and vulnerability gates | passed locally |
| Operations | Support policy and operations runbook; structured observability is tracked in [#96](https://github.com/edithatogo/osf-cli-go/issues/96) | partial; follow-up required |
| Documentation | Commands, usage, install, troubleshooting, architecture, migration, and support references | ready locally |
| Distribution | Cross-platform artifacts, signed provenance, SBOM, MCP registry, and plugin packages; clean release-candidate verification is recorded in [#98](https://github.com/edithatogo/osf-cli-go/issues/98) | ready locally; external gates remain |
| Governance | Compatibility, contribution, security, support, issue triage, and release policies are documented | ready locally |

## Decision

The repository is a stronger 1.0 release candidate, but not yet eligible for a
`v1.0.0` tag. Live OSF validation passed with disposable-resource cleanup;
security review, hosted PR gates, provider review, and package-manager gates
remain external. This review does not waive maintainer/provider acceptance.
