# Auth Capability Matrix

Date: 2026-05-17

## Legend

- `supported-offline-fixture`: covered by local tests or request-signing
  fixtures.
- `supported-via-token-bootstrap`: works through the existing bearer-token path
  after the user supplies or exports a PAT.
- `public-only`: works without credentials only when OSF permits public access.
- `unknown-pending-live-validation`: username/password direct behavior needs
  opt-in live OSF validation.

## Command Matrix

| Command | Token auth | Username/password direct auth | Bootstrap path | Notes |
|---|---|---|---|---|
| `auth whoami` | supported-offline-fixture | supported-live-validated | supported-via-token-bootstrap | Basic auth header is fixture-tested; live username/password validation passed on 2026-05-17. |
| `projects list` | supported-offline-fixture | supported-live-validated | supported-via-token-bootstrap | Live username/password validation passed on 2026-05-17. |
| `projects get` | supported-offline-fixture | supported-live-validated | supported-via-token-bootstrap | Live username/password validation passed on 2026-05-17 against project `xj6qc`. |
| `projects create` | supported-offline-fixture | unknown-pending-live-validation | supported-via-token-bootstrap | Write command requires confirmation. |
| `projects update` | supported-offline-fixture | unknown-pending-live-validation | supported-via-token-bootstrap | Write command requires confirmation. |
| `projects delete` | supported-offline-fixture | unknown-pending-live-validation | supported-via-token-bootstrap | Write command requires confirmation. |
| `components list` | supported-offline-fixture | supported-live-validated | supported-via-token-bootstrap | Live username/password validation passed on 2026-05-17 against project `xj6qc`. |
| `files list` | supported-offline-fixture | supported-live-validated | supported-via-token-bootstrap | Live username/password validation passed on 2026-05-17 against project `xj6qc`; nested folder-ID listing also passed on 2026-05-17. |
| `files download` | supported-offline-fixture | unknown-pending-live-validation | supported-via-token-bootstrap | WaterButler request signing is shared. |
| `files upload` | supported-offline-fixture | unknown-pending-live-validation | supported-via-token-bootstrap | WaterButler write operation. |
| `files mkdir` | supported-offline-fixture | unknown-pending-live-validation | supported-via-token-bootstrap | WaterButler write operation. |
| `files rm` | supported-offline-fixture | unknown-pending-live-validation | supported-via-token-bootstrap | WaterButler write operation with confirmation. |
| `files addons` | supported-offline-fixture | supported-live-validated | supported-via-token-bootstrap | Live username/password validation passed on 2026-05-17 against project `xj6qc`. |
| `search` | supported-offline-fixture | public-only | supported-via-token-bootstrap | Bounded live validation passed on 2026-05-17 with `--limit 5`; this is public-context behavior, not credential-specific access. |
| `preprints list` | supported-offline-fixture | public-only | supported-via-token-bootstrap | Bounded live validation passed on 2026-05-17 with `--limit 5`; this is public-context behavior, not credential-specific access. |
| `registrations create` | supported-offline-fixture | unknown-pending-live-validation | supported-via-token-bootstrap | Write command requires confirmation and schema ID. |
| `export` | supported-offline-fixture | supported-live-validated | supported-via-token-bootstrap | Live username/password validation passed on 2026-05-17 against project `xj6qc`. |

## Live Validation Gate

Direct username/password support must remain documented as a fallback or
experimental compatibility mode until `tools/livevalidation` is run with:

- `OSF_LIVE_VALIDATION=1`
- `OSF_USERNAME`
- `OSF_PASSWORD`
- `OSF_VALIDATE_PROJECT`
- optionally `OSF_VALIDATE_DOWNLOAD`

Sanitized output from that run should update this matrix before public docs
claim live support for individual commands.
