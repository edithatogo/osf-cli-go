# Version 1.0 launch roadmap

Version 1.0 means the project is prepared to maintain stable automation
contracts, not merely that it has accumulated enough features.

## Launch gates

| Gate | Required evidence |
|---|---|
| Stable contracts | Versioned CLI JSON and MCP schemas, compatibility tests, deprecation policy, and migration guide |
| Reliability | Resumable transfers, deterministic manifests, pagination and retry tests, cancellation, and atomic writes |
| Security | Threat model, credential-store decision, redaction tests, dependency and container scans, provenance, SBOM, signatures, and documented write approvals |
| Live behavior | Opt-in validation across public reads, private reads, safe writes, destructive confirmation, large transfers, and rate-limit handling |
| Quality | CI on supported operating systems, race/lint/vet/vulnerability gates, high-risk path coverage, and no unresolved release-blocking findings |
| Operations | Structured logs, dynamic versions, support policy, issue triage, release cadence, rollback procedure, and incident process |
| Documentation | Complete command and MCP references, install/upgrade/uninstall paths, troubleshooting, examples, and architecture decisions |
| Distribution | Signed multi-platform release, package-manager channels, official MCP registry, validated plugins, and evidence-backed directory listings |
| Governance | Maintainer expectations, compatibility window, contribution policy, code of conduct, and explicit support boundaries |

The implementation source of truth is issue
[#21](https://github.com/edithatogo/osf-cli-go/issues/21) and its Conductor
track. A `v1.0.0` tag must not be created until every gate is either satisfied
or explicitly waived in a release-candidate review.

The working evidence set is maintained in the [feature
matrix](feature-matrix.md), [registry scorecard](registry-scorecard.md), and
the [1.0 hardening and maturity track](https://github.com/edithatogo/osf-cli-go/blob/master/conductor/archive/v1-hardening-maturity_20260713/spec.md).

## Recommended sequence

1. Complete the source-tool parity tracks and update the feature matrix.
2. Freeze candidate CLI and MCP contracts and publish a release candidate.
3. Run the hardening, quality-harness, security, reliability, live-OSF, and cross-platform campaigns.
4. Complete package-manager, plugin, and registry distribution using the scorecard.
5. Publish migration, support, and governance documents.
6. Run a clean release review and publish `v1.0.0` with signed provenance.

## Tacit marketing

Discoverability should come from useful technical artifacts rather than claims:

- searchable registry and agent-marketplace listings;
- reproducible comparison tables and benchmark fixtures;
- installable extensions with concise workflow examples;
- documentation that answers real OSF automation problems;
- release notes with measured compatibility and quality evidence;
- a software paper or methods preprint only if the comparative evaluation and
  reproducible workflow constitute a defensible contribution.
