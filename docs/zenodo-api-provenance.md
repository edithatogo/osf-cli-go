# Zenodo API provenance and version policy

Last reviewed: 2026-07-15

The canonical machine-readable snapshot is
`docs/zenodo-api-source.json`. It records the API behavior that this project may
implement; it is not a claim that every listed capability is already available.

## Authority and version decision

The supported source is the official [Zenodo developer documentation](https://developers.zenodo.org/),
supplemented by official [sandbox guidance](https://help.zenodo.org/docs/get-started/),
[terms of use](https://about.zenodo.org/terms/), and [repository policies](https://about.zenodo.org/policies/).
The developer documentation does not publish a semantic API version or a pinned
OpenAPI schema for the documented depositions API. The project therefore uses a
`documentation-date` policy: behavior is pinned to the reviewed capability
snapshot and its SHA-256 digest. Upstream structural changes require review,
fixture updates, and a new digest before implementation claims change.

The snapshot names this surface `documented-depositions-rest` to distinguish it
from newer Zenodo/InvenioRDM concepts that are visible in the service but are not
fully documented as the stable integration contract on the developer page.

## Protocol boundaries

- REST production base: `https://zenodo.org/api/`.
- REST sandbox base: `https://sandbox.zenodo.org/api/`. Sandbox accounts and
  credentials are separate from production and all write validation starts here.
- OAI-PMH base: `https://zenodo.org/oai2d`. OAI-PMH remains a separate adapter
  because resumption tokens, XML metadata formats, and expiry differ from REST.
- Public published-record reads do not require a token. Depositions and actions
  require a bearer token. Write and action scopes are distinct.
- Tokens are sent only through the `Authorization: Bearer` header. Although the
  documentation describes a query parameter, this project rejects that transport
  because URLs are routinely retained in logs, histories, and intermediary data.

## Limits and lifecycle evidence

The reviewed documentation reports guest, authenticated, search, and OAI-PMH
rate limits and exposes `X-RateLimit-*` response headers. OAI-PMH returns 50
records per page and its resumption tokens expire after two minutes. File limits
are recorded as 50 GB total/per file and 100 files per record for the newer
bucket upload path described by the official quickstart.

Publishing is treated as irreversible: published files cannot be added, changed,
or removed through the normal workflow. Terms and repository policies also make
the uploader responsible for rights, privacy, and licensing. These facts are
inputs to the later publication-state and cross-provider tracks, not write
authorization in this track.

## Drift checks

Run the deterministic offline check in every pull request:

```sh
go run ./tools/checkzenodoapi
```

The checker validates source authority, ordering, unique identifiers, required
core capabilities, safety classifications, limits, and the canonical snapshot
digest. A scheduled/manual workflow additionally runs:

```sh
go run ./tools/checkzenodoapi -online
```

The online mode fetches only the four official public sources and verifies the
reviewed structural markers. It sends no credentials and performs no repository
API operations. A failure is evidence to review upstream documentation; it must
not be bypassed by weakening markers without updating this provenance record.
