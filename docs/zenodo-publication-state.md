# Zenodo publication lifecycle

Zenodo DOI and publication actions are modeled independently from public record
reads and sandbox file transfers. The lifecycle package rejects production
hosts and validates the entire operation before an authenticated request is
created.

## States and transitions

| Current state | Action | Result | Scope | Confirmation |
|---|---|---|---|---|
| `draft` | verify automatic DOI reservation | `doi_reserved` | authenticated read | explicit authorization |
| `draft` | publish | `published` | `deposit:write`, `deposit:actions` | exact challenge; irreversible |
| `draft` | discard | `discarded` | `deposit:write` | exact challenge; destructive |
| `doi_reserved` | publish | `published` | `deposit:write`, `deposit:actions` | exact challenge; irreversible |
| `doi_reserved` | discard | `discarded` | `deposit:write` | exact challenge; destructive |
| `published` | new version | `version_draft` | `deposit:actions` | explicit authorization |
| `version_draft` | publish | `published` | `deposit:write`, `deposit:actions` | exact challenge; irreversible |
| `version_draft` | discard | `published` | `deposit:write` | exact challenge; destructive |

All other state/action pairs fail locally. `discarded` is terminal. A published
record cannot be discarded through this workflow. Discarding an unpublished
new-version draft leaves the preceding published version available.

## Publication metadata

Publication requires a non-empty title, description, upload type, and at least
one named creator. Access policy is validated as follows:

- `open` requires a license and rejects embargo or access-condition fields.
- `embargoed` requires a license and a future embargo date.
- `restricted` requires access conditions and rejects an embargo date.
- `closed` rejects embargo and access-condition fields.

The caller supplies Zenodo's reviewed license identifier. The package checks
that policy requires one where appropriate; Zenodo remains authoritative for
whether a particular identifier is accepted.

## Safety gates

Every action requires a provider-specific record ID, explicit authorization,
and every token scope represented in the plan. Publication requires both write
and action scopes because validated metadata is applied before publication;
the irreversible action is never retried automatically. Dry-run returns the expected
target state and deterministic confirmation challenge without performing HTTP.
Publish and discard execute only when that exact challenge is supplied.

Lifecycle actions are not retried automatically. If metadata update,
publication, new-version creation, or discard returns an ambiguous failure, the
error directs the caller to inspect the current draft/latest-draft state before
retrying. A publication-action failure after metadata update is a typed partial
result: the draft may contain the validated metadata even though publication
was not confirmed.

Audit evidence records the record ID, transition, outcome, dry-run status, and
irreversible/destructive flags. It omits metadata, token values, and scope
inventories, and redacts token-shaped values from errors. No publication write
is exposed through MCP or against production Zenodo. The stable CLI exposes the
sandbox lifecycle through dry-run-first commands, exact confirmation challenges,
and the same state-machine validation used by the live validation harness.

The current sandbox automatically reserves a DOI when a deposition is created
and returns it in `metadata.prereserve_doi`. The reservation transition verifies
that value through an authenticated read. It deliberately does not send the
legacy top-level `prereserve_doi` field, which the 2026-07-15 sandbox rejects.

The opt-in harness is `go run ./tools/zenodopublicationvalidation -live` with
`ZENODO_PUBLICATION_VALIDATION=1`, the sandbox base URL, and a dedicated token.
It writes redacted evidence to `docs/zenodo-publication-validation-evidence.md`.
The 2026-07-15 proof used only `deposit:write` and `deposit:actions`, excluded
`user:email`, and revoked the one-use token after public-record verification.
