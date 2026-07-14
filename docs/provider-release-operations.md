# Provider Release Operations

This policy defines the performance, cleanup, evidence, and migration controls
for OSF, Zenodo, and cross-provider release validation.

## Performance budgets

| Surface | Default budget | Release check |
|---|---|---|
| Provider claim validation and report rendering | 10 seconds in CI | `provider-validation.yml` offline job timeout |
| Zenodo REST reads | 30-second request timeout, 8 MiB response, 100 pages, 4 concurrent requests | client tests and race tests |
| Zenodo OAI-PMH | 30-second request timeout, 8 MiB response, 100 pages, 5,000 records, 2 concurrent requests | adapter tests and bounded-harvest fixtures |
| Zenodo sandbox transfer | 2-minute request timeout, 8 MiB control response, 50 GiB file, 100 files, 2 retries | transfer boundary tests and sandbox evidence |
| Live validation jobs | 10 minutes for sandbox; 15 minutes for OSF production | workflow job timeouts |

Increasing a default requires tests, an operational rationale, and review of
memory, service-rate, and CI-runtime impact. The release claim checker itself
must remain network-free and deterministic.

## Sandbox cleanup

- Transfer and cross-provider validations create unpublished drafts and must
  delete them through cancellation-resistant compensation.
- Publication validation is irreversible. Its published sandbox record is
  retained and identified by a public sandbox URL; any unpublished new-version
  draft must be discarded.
- Every `sandbox-validated` manifest claim must set `resourceDisposition` to
  `deleted` or `published-retained`. A retained publication must include
  `resourceRecord`; a deleted resource must not.
- A failed live job is investigated from redacted evidence before rerun. Never
  retry an ambiguous create or publish action without inspecting provider state.
- Validation credentials are one-use or narrowly scoped and are revoked after
  the run. Credentials and authenticated payloads are not release artifacts.

## Compatibility and migration

Public provider reads are part of the additive CLI/API/MCP baseline. Internal
Zenodo transfer, publication, and cross-provider packages are validation
machinery, not public compatibility promises. Promoting them requires a new
track, explicit command/tool fixtures, deprecation analysis, migration guidance,
and confirmation and rollback semantics appropriate to the provider.

## Release procedure

1. Run the offline provider checker and render the report.
2. Run only the necessary opt-in live jobs with protected environments.
3. Update evidence digests and resource dispositions after reviewing cleanup.
4. Run the full compatibility, security, registry, feature-matrix, and release
   gates from a clean checkout.
5. Tag only when the release review records every unrun live or external gate as
   a blocker or explicit waiver.
