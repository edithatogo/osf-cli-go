# Operations Runbook

## Release

Run the repository quality gates, build the cross-platform artifacts, verify
checksums and signatures, inspect the SBOM and provenance, then publish release
notes with compatibility and known-gate status. Do not tag `v1.0.0` while a
release-blocking finding lacks a dated waiver.

## Incident response

1. Triage the report without requesting or reproducing with credentials or
   private research files.
2. Determine whether the issue affects credentials, local writes, OSF writes,
   releases, or only presentation.
3. For credential or supply-chain risk, use private GitHub vulnerability
   reporting and pause affected release automation.
4. For a release regression, mark the release status, publish a correction or
   rollback guidance, and preserve the failing evidence.
5. Add a redacted regression test before closing the incident.

## Observability

Set `OSF_EVENT_LOG` to a private JSONL file for local diagnosis. Correlate
`operationId` and `requestId` across CLI, API, transfer, and MCP events. Event
logs are opt-in, local-only, redacted, and owner-readable; apply an operator
retention or rotation policy and inspect them before sharing. Never set the
destination to stdout because command JSON output must remain unpolluted.

## Rollback

Rollback means directing users to the previous verified release or container
tag and, when necessary, withdrawing the affected release assets. OSF data is
not automatically reverted by the client; any remote write remediation must be
explicit, authenticated, confirmed, and documented separately.

## Support and cadence

The default branch receives active fixes. The latest stable release receives
security and release-blocking fixes through the next minor-release window.
Use GitHub Issues for reproducible defects and the security policy for private
vulnerability reports.
