# Multi-provider validation levels

`docs/multi-provider-validation.json` is the release claim source of truth. Each
claim is dated and bound to repository evidence by SHA-256. Ordinary CI runs
`go run ./tools/checkproviderrelease`; stale evidence, digest drift, duplicate
capabilities, unknown levels, and unsupported production claims fail the build.

Every sandbox claim also records its resource disposition. `deleted` requires
that no resource URL remain. `published-retained` is reserved for irreversible
publication and requires a public `https://sandbox.zenodo.org/...` record. This
prevents a successful API call from masking an orphaned or intentionally
retained resource.

## Levels

| Level | Meaning | Required boundary |
|---|---|---|
| `offline-tested` | Deterministic fixtures, contracts, and local quality gates passed without a provider request | Repository evidence with a matching digest |
| `sandbox-validated` | The current behavior ran against a provider sandbox or disposable non-production service | Dated sandbox evidence, cleanup result, and matching digest |
| `production-validated` | The current behavior ran against the production provider | Dated evidence plus a public HTTPS production receipt; sandbox URLs are rejected |

`live-validated` is not a release level because it does not identify the target
environment. Existing harness evidence may use that execution phrase, but the
release manifest must classify it as sandbox or production explicitly.

## Opt-in workflow

`.github/workflows/provider-validation.yml` has only a manual
`workflow_dispatch` trigger. Every provider job defaults to disabled and uses a
protected GitHub environment. Sandbox transfer, sandbox publication,
cross-provider copy, and OSF production validation have separate boolean inputs
and credential names. Pull requests, pushes, and schedules cannot start these
jobs.

The offline report job always renders `provider-validation-report.md` from the
validated manifest. Selected live jobs upload only their sanitized Markdown
evidence. Credentials remain environment-only and are never included in the
report or artifacts.

Run
`go run ./tools/checkproviderrelease -report docs/multi-provider-validation-report.md`
to render the dated report deterministically. Tagged binary releases publish
that report beside the binaries, and the container SBOM/provenance workflow is
gated on the same claim manifest. The current manifest makes no
production-validated claim.
