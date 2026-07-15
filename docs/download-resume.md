# Resumable transfers

`osf files download --file` and `--tree` use checkpointed atomic writes. If a
download is interrupted, the partial file is retained beside the destination
as `<destination>.part` and its state is recorded in
`<destination>.resume.json`.

On the next invocation, the client validates the source URL, destination, and
known size before requesting the remaining byte range. If the provider ignores
the range request, the client safely restarts from byte zero. A completed
transfer verifies the known size and any OSF provider checksum before renaming
the partial file into place. Checkpoints are removed only after successful
finalization.

Resume state is automatic and does not change the existing `--conflict`
policies. `fail`, `skip`, and `overwrite` continue to apply to a completed
destination. Resume checkpoint files contain source and upload identity
fingerprints plus transfer state, never raw URLs or credentials.

## Provider-supported uploads

The WaterButler upload endpoint currently exposes a one-shot PUT contract, so
the CLI does not pretend that it can resume an upload that the provider cannot
acknowledge in chunks. The reusable `download.ResumeFileUpload` contract
supports providers that expose acknowledged chunk offsets and keeps a
fingerprinted checkpoint for those integrations.
