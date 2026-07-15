# Live Validation Matrix

Live checks are opt-in and require explicit credentials or a disposable test
project. They must never run as ordinary unit tests or in pull requests.

| Provider | Area | Required scenario | Evidence level | Current status |
|---|---|---|---|---|
| OSF | Authentication and private reads | `whoami`, projects, and contributors with a disposable token | production-validated only with a public receipt; otherwise redacted run evidence | opt-in, not currently production-validated |
| OSF | Public reads and search | Get a public node, search projects, and list preprints | redacted request shape, counts, and HTTP status | opt-in |
| OSF | Files and safe writes | List, download, verify, upload, and clean a fixture | checksums, manifest, project scope, and cleanup | opt-in |
| OSF | Destructive writes and registrations | Delete only a disposable fixture; create but never publish a draft registration | explicit confirmation and cleanup evidence | opt-in |
| Zenodo | Public REST and OAI-PMH reads | Search/get records and perform bounded harvests | offline-tested contract plus optional public response evidence | offline-tested |
| Zenodo | Draft file transfer | Create, upload, verify/resume, and delete an unpublished sandbox draft | sandbox-validated digest and `deleted` disposition | sandbox-validated |
| Zenodo | Publication lifecycle | Reserve, confirm, publish, create a new version, and discard its draft | sandbox-validated retained record plus discarded draft | sandbox-validated |
| Cross-provider | OSF-qualified to Zenodo draft copy | Map metadata, copy files, verify provenance, replay, and compensate | sandbox-validated sidecar, checksums, and `deleted` disposition | sandbox-validated |
| All | Failure behavior | Invalid token, rate limit, missing resource, timeout, and ambiguous mutation response | typed redacted errors; fixtures plus explicit live evidence | fixture + opt-in |

The release checklist must identify which rows were run for a release candidate
and which remain unrun because credentials, permissions, or a disposable OSF
fixture were unavailable.

## Repeatable invocation

Set credentials and a disposable project reference only in the shell running
the validation. Then opt in explicitly:

```sh
OSF_LIVE_VALIDATION=1 \
OSF_VALIDATE_WRITES=1 \
OSF_TOKEN="$OSF_TOKEN" \
OSF_VALIDATE_PROJECT="<disposable-project-id-or-url>" \
go run ./tools/livevalidation -live
```

Set `OSF_VALIDATE_DOWNLOAD` to a known disposable fixture file reference to
enable the download row. The tool writes sanitized evidence to the current
`docs/live-osf-validation-evidence.md` report by default and never writes
credential values. Omit `OSF_VALIDATE_WRITES=1` to keep upload, conflict, and
cleanup scenarios pending even when live read validation is enabled.

Zenodo and cross-provider jobs are invoked only through the manual
`.github/workflows/provider-validation.yml` workflow. All inputs default to
false, use protected environments, and follow the cleanup contract in
`docs/provider-release-operations.md`.
