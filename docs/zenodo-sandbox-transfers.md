# Zenodo sandbox transfers

The `internal/zenodotransfer` adapter implements authenticated transfers only
against `https://sandbox.zenodo.org/api/` or a loopback fixture server. It
rejects production Zenodo and all unapproved origins before sending an
authorization header.

## Configuration

- `ZENODO_TOKEN` is a dedicated sandbox token and is never inferred from
  `OSF_TOKEN`.
- `ZENODO_BASE_URL` defaults to the Zenodo sandbox API. The live harness also
  verifies that this value identifies the sandbox before creating a draft.
- Tokens are sent only in the `Authorization: Bearer` header and are redacted
  from returned errors and evidence.

## Transfer contract

- Uploads accept regular files only, enforce configurable byte and file-count
  limits, and require an explicit `fail`, `skip`, or `overwrite` conflict
  policy.
- The documented Zenodo bucket upload is a whole-file `PUT`. A transient retry
  rewinds the source and starts again at byte zero. The result therefore never
  reports a partially resumed upload.
- Upload completion requires the server-reported size and MD5 checksum to match
  the local source. A mismatch remains a failure and retains the non-secret
  checkpoint for diagnosis or retry.
- Downloads use an atomic partial file, validate byte-range continuation,
  verify size and checksum, and retain checkpoints after interruption or
  cancellation.
- Draft deletion is a separate, explicit cleanup operation. The opt-in harness
  always attempts it after draft creation, including after transfer failure.

This package is an internal sandbox validation surface. No Zenodo write command
or MCP tool is advertised. Publication is separately governed by the
publication-state track and cannot be reached through this adapter.

## Disposable validation

Generate a credential-free dry-run evidence record with:

```sh
go run ./tools/zenodosandboxvalidation
```

Run the live sandbox proof only with a dedicated sandbox token and explicit
opt-in:

```sh
ZENODO_TOKEN="..." \
ZENODO_BASE_URL="https://sandbox.zenodo.org/api/" \
ZENODO_SANDBOX_VALIDATION=1 \
go run ./tools/zenodosandboxvalidation -live
```

The harness creates one unpublished draft, uploads a generated file, verifies a
complete download, deliberately interrupts and resumes a second download, and
then deletes the draft even when a transfer step fails. Its evidence contains
only the sandbox host, byte counts, checksums, status, and cleanup outcome.
