# Provider Environment Evidence

- Verified: 2026-07-15 (Australia/Sydney)
- Repository: `edithatogo/osf-cli-go`
- Source: GitHub Environments REST API using the authenticated `gh` CLI

| Environment | Protection rule | Intended boundary |
|---|---|---|
| `provider-sandbox` | 1-minute wait timer | Disposable Zenodo transfer and cross-provider copy |
| `provider-sandbox-publication` | 5-minute wait timer | Irreversible Zenodo Sandbox publication |
| `provider-production` | 10-minute wait timer | Explicit OSF production validation |

The environments exist in GitHub and match
`.github/workflows/provider-validation.yml`. Their jobs are manual and false by
default. This evidence does not claim that provider credentials are stored or
that any production validation has run. Sandbox credentials used for earlier
live evidence were one-use and revoked.
