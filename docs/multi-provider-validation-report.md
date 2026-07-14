# Multi-provider validation report

- Schema: 1
- Generated: 2026-07-14T18:14:28Z
- Opt-in workflow: `.github/workflows/provider-validation.yml`
- Production-validated claims: 0

| Provider | Capability | Level | Validated | Resource disposition | Evidence |
|---|---|---|---|---|---|
| cross-provider | repository.contract.v1 | offline-tested | 2026-07-15 | not-applicable | `docs/repository-provider-contract.md` (sha256:c3920b60c955) |
| osf | cli.api.mcp.compatibility | offline-tested | 2026-07-15 | not-applicable | `docs/compatibility-policy.md` (sha256:42fb35523f58) |
| zenodo | rest.oai.contract | offline-tested | 2026-07-15 | not-applicable | `docs/zenodo-api-source.json` (sha256:468eca003f23) |
| cross-provider | operational.events.v1 | offline-tested | 2026-07-15 | not-applicable | `docs/observability.md` (sha256:6daff0663d84) |
| zenodo | draft.files.transfer | sandbox-validated | 2026-07-15 | deleted | `docs/zenodo-sandbox-validation-evidence.md` (sha256:60b3181b8fc0) |
| zenodo | deposit.publication.lifecycle | sandbox-validated | 2026-07-15 | published-retained (https://sandbox.zenodo.org/records/565256) | `docs/zenodo-publication-validation-evidence.md` (sha256:21340a7634a9) |
| cross-provider | osf.zenodo.copy.saga | sandbox-validated | 2026-07-15 | deleted | `docs/cross-provider-sandbox-validation-evidence.md` (sha256:1b7542e8dd90) |
| cross-provider | release.governance.v1 | offline-tested | 2026-07-15 | not-applicable | `docs/provider-release-operations.md` (sha256:5ae6391c0dd6); `docs/adr-001-multi-provider-release-contract.md` (sha256:74e4fcacb374); `docs/provider-environment-evidence.md` (sha256:c55023589c71) |
| zenodo | registry.readonly.surface | offline-tested | 2026-07-15 | not-applicable | `server.json` (sha256:3e1c406908ae); `internal/mcpserver/testdata/compatibility/mcp-tools.json` (sha256:add329e4be13) |
