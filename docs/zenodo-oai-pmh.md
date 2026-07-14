# Zenodo OAI-PMH harvesting

OSF CLI Go exposes Zenodo's public OAI-PMH endpoint as a protocol-specific
adapter. It is not REST search: metadata prefixes, sets, date selectors,
protocol errors, and opaque resumption tokens retain their OAI-PMH meaning.

## CLI

List one page and receive its continuation:

```console
osf zenodo oai harvest --metadata-prefix oai_dc --set user-example --output json
```

Resume a saved page or follow all pages within the client budgets:

```console
osf zenodo oai harvest --resume-token TOKEN --output json
osf zenodo oai harvest --metadata-prefix datacite --all --output json
```

Inspect selective-harvesting sets and available schemas:

```console
osf zenodo oai sets
osf zenodo oai formats --identifier oai:zenodo.org:12345
```

The JSON page contains `records` and `next`. Each non-deleted record retains its
native metadata as `application/xml`, plus response date, source endpoint,
metadata prefix, set, and datestamp provenance. Deleted records preserve their
header and optional `about` provenance without inventing metadata.

## MCP

The stdio server exposes three separate read-only tools:

- `zenodo_oai_records_list` returns one bounded page and its opaque continuation.
- `zenodo_oai_sets_list` lists all sets within the page budget.
- `zenodo_oai_formats_list` lists metadata schemas.

OAI-PMH is public and never receives `OSF_TOKEN` or a Zenodo REST token.
Transport retries are limited to transient network errors and HTTP 429, 502,
503, and 504 responses. XML bytes, pages, records, retry delays, redirects, and
context cancellation are bounded and fixture-tested. A token with a known
expiry is rejected locally; the typed `badResumptionToken` protocol response
remains authoritative for externally persisted tokens.
