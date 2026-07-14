# Cross-provider transfer contract

Cross-provider copy is an explicit one-way operation, never a mirror. The
planner requires a qualified source identity, the exact OSF-to-Zenodo or
Zenodo-to-OSF direction, a different destination provider, either a new record
or an existing draft ID, authorization, a conflict policy, and a publication
intent.

## Dry-run mapping

Every report accounts for the same source fields in stable order: title,
description, upload type, creators, keywords, access, license, embargo,
identifiers, version, native metadata, and files. Each field is marked as:

- `exact` when the reviewed destination contract preserves its meaning;
- `transformed` when the target meaning differs and the reason is recorded;
- `preserved_native` when no stable target field exists but provenance retains
  the source value; or
- `blocked` when execution would require an unsafe inference.

A private OSF source has no implicit Zenodo access equivalent. The report stays
non-executable until the caller chooses an explicit open, embargoed, restricted,
or closed Zenodo policy. Open and embargoed Zenodo targets also require an
explicit license. Zenodo open access maps to OSF public visibility; embargoed,
restricted, and closed access map only to private visibility, with the richer
access, license, identifier, and version semantics preserved in provenance.

## Identity and replay

Source and destination identities remain provider-qualified. Source identifiers
may become related metadata but never replace a destination's native identity.
The report includes a SHA-256 digest of lossless provider-native metadata and a
list of every non-exact transformation.

The `xfer-v1-...` idempotency key is derived from direction, source snapshot,
file sizes/checksums, destination, mapped target metadata, conflict policy, and
publication intent. Capture time is excluded, so an unchanged request produces
the same key while any content or semantic change produces a different key.

`draft_only` and `publish_after_copy` are planning declarations. A copy executor
must complete and verify the draft before a separately confirmed publication
action; the mapping report itself performs no network I/O.

## Checkpoints and compensation

Versioned checkpoints bind the mapping idempotency key to destination,
conflict, publication intent, and deterministic ordered steps. New-destination
workflows create the draft, apply metadata, copy files in path order, verify the
draft, and finalize it. A requested publication is a separate final step marked
as confirmation-required and irreversible.

Only the next pending or failed step can advance. Every attempt and redacted
failure is recorded, and replay is refused when the report's idempotency key
differs. Partial results classify every file as completed, failed, or pending
without claiming publication.

Compensation is planned in reverse order for completed draft mutations: copied
files are deleted, replaced files/metadata require caller-provided rollback
references, and a newly created draft is discarded last. Once publication has
completed, compensation is rejected because the provider boundary is
irreversible.
