# Zenodo provider CLI

Zenodo reads are explicit under `osf zenodo`; existing OSF commands retain
their names and behavior. Published REST records are public, so these commands
do not require or reuse `OSF_TOKEN`.

## Record identifiers

Commands that select a published record accept exactly these forms:

- Native decimal ID: `12345`
- Provider-qualified ID: `zenodo:record:12345`
- Canonical URL: `https://zenodo.org/records/12345`
- Sandbox URL: `https://sandbox.zenodo.org/records/12345`

DOIs, deposition URLs, cross-provider IDs, URL queries/fragments, and path-like
values are rejected rather than guessed.

## Read commands

```console
osf zenodo records search "open methods" --limit 10
osf zenodo records search "open methods" --output json
osf zenodo records get zenodo:record:12345
osf zenodo records get https://zenodo.org/records/12345 --json
osf zenodo files list 12345
osf zenodo capabilities --json
```

PowerShell uses the same argument contract:

```powershell
osf.exe zenodo records search "open methods" --limit 10 --output json
osf.exe zenodo files list "zenodo:record:12345"
```

Record JSON includes `qualifiedId`, provider and kind, native identifiers,
normalized discovery fields, links, and the lossless `nativeMetadata` JSON.
File JSON includes the parent record's qualified ID, provider-native file ID,
key, size, checksum, links, and preferred public download URL.

## Guarded write commands

Zenodo writes use only `ZENODO_TOKEN`; the CLI never reuses `OSF_TOKEN`. They
default to `https://sandbox.zenodo.org/api/`. Production writes require all of
`--production`, `ZENODO_BASE_URL=https://zenodo.org/api/`, `--execute`, a
dedicated token with the required deposit scopes, and an exact action-specific
confirmation. The production flag alone never changes the endpoint.

```powershell
$env:ZENODO_TOKEN = '<dedicated sandbox token>'

osf.exe zenodo deposits create --execute --json
osf.exe zenodo deposits get 12345 --json
osf.exe zenodo deposits metadata 12345 --metadata metadata.json
osf.exe zenodo deposits metadata 12345 --metadata metadata.json --execute
osf.exe zenodo files draft-list 12345
osf.exe zenodo files upload 12345 .\dataset.csv --execute
osf.exe zenodo files delete 12345 file-id --confirm 'zenodo:delete-file:12345:file-id'
osf.exe zenodo deposits reserve-doi 12345
osf.exe zenodo deposits new-version 12345
osf.exe zenodo deposits new-version 12345 --execute
osf.exe zenodo deposits discard 12345
osf.exe zenodo deposits discard 12345 --execute --confirm 'zenodo:discard:12345:discarded'
osf.exe zenodo publish 12345 --metadata metadata.json
osf.exe zenodo publish 12345 --metadata metadata.json --execute --confirm 'zenodo:publish:12345:published'
```

For production, include `--production` and the required confirmation. For
example, production draft creation is:

```powershell
$env:ZENODO_BASE_URL = 'https://zenodo.org/api/'
$env:ZENODO_TOKEN = '<dedicated production token>'

osf.exe zenodo deposits create --production --execute `
  --confirm 'zenodo:production:create-draft' --json
```

Production draft reads use the same explicit target selection:

```powershell
osf.exe zenodo deposits get 12345 --production --json
osf.exe zenodo files draft-list 12345 --production --json
```

Metadata updates and publication consume a strict JSON object. Unknown fields,
trailing JSON values, incomplete creators, invalid access policies, and stale
embargo dates fail locally before a client or authenticated request is created:

```json
{
  "title": "Dataset title",
  "description": "Dataset description",
  "uploadType": "dataset",
  "creators": [{"name": "Doe, Jane"}],
  "keywords": ["reproducibility"],
  "access": "open",
  "license": "cc-by-4.0"
}
```

Draft creation and upload require `--execute`. Metadata update emits a validated
preview without `--execute`. Lifecycle commands emit a dry-run plan by default;
publish and discard execution additionally require the exact `confirmation`
value from that plan. Production draft creation, upload, and metadata update
also require the action-specific confirmation shown by their validation errors.
File deletion prints its deterministic confirmation in the validation error and
performs no request until the exact value is supplied.

The implementation follows Zenodo's official depositions, bucket upload, and
deposition-actions API documented at <https://developers.zenodo.org/>. Uploads
are whole-file PUTs with explicit `fail`, `skip`, or `overwrite` conflict policy.

## Capability guidance

`osf zenodo capabilities` is generated from the reviewed provider contract.
The capability contract reports write operations as partial because only
unpublished depositions are accepted and lifecycle constraints remain
provider-specific. `zenodo records create` and `zenodo records update` remain
unsupported aliases; writes use the explicit `zenodo deposits` and draft-file
commands above. MCP writes remain unavailable.

OAI-PMH harvesting remains a separate subgroup documented in
[Zenodo OAI-PMH harvesting](zenodo-oai-pmh.md).
