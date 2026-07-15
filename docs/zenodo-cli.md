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

## Capability guidance

`osf zenodo capabilities` is generated from the reviewed provider contract.
Write-shaped invocations such as `zenodo records create`, `zenodo files upload`,
and `zenodo publish` fail with a typed partial-capability error and explain the
draft, sandbox, or irreversible lifecycle constraint. They never make a write
request. Sandbox transfers and publication remain owned by issues #108 and
#109.

OAI-PMH harvesting remains a separate subgroup documented in
[Zenodo OAI-PMH harvesting](zenodo-oai-pmh.md).
