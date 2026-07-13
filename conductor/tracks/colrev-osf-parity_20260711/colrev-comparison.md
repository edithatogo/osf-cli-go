# CoLRev Comparison

Reviewed 2026-07-13 against the public `main` branch of
[CoLRev-Environment/colrev](https://github.com/CoLRev-Environment/colrev).

## Dated Evidence

| Dimension | Evidence | Assessment |
|---|---|---|
| OSF search | [`colrev/packages/osf/README.md`](https://github.com/CoLRev-Environment/colrev/blob/main/colrev/packages/osf/README.md) documents OSF API node searches with title, ID, type, category, year, description, tags, and creation-date filters | The material reusable gap is bibliographic export from OSF search results |
| Record construction | [`osf_api.py`](https://github.com/CoLRev-Environment/colrev/blob/main/colrev/packages/osf/src/osf_api.py) maps OSF records to ID, title, abstract, keywords, year, and URL, then resolves bibliographic contributors | Go search results now preserve the same core fields; contributor lookup remains out of scope to avoid hidden N+1 requests |
| Workflow | [`osf.py`](https://github.com/CoLRev-Environment/colrev/blob/main/colrev/packages/osf/src/osf.py) integrates search history, pagination, reruns, and BibTeX loading | `osf search --bibtex` provides a deterministic interchange boundary; full review state management remains outside this CLI |
| PDF handling | CoLRev's README and comparison table document PDF retrieval and preparation | Existing OSF CLI download primitives remain available; this track does not add implicit downloads or PDF preparation |
| Maturity | GitHub metadata on 2026-07-13 showed 4,664 commits, 43 stars, 29 open issues, MIT license, and latest release 0.16.2 published 2026-02-24 | CoLRev is the more mature literature-review environment; this project adopts only the stable OSF-to-bibliography contract |

## Adopted Contract

- `osf search <query> --bibtex` emits one deterministic `@misc` entry per
  result.
- Fields are emitted in stable order: title, abstract, keywords, year, and URL.
- BibTeX values escape braces, backslashes, and newlines.
- `osf_search` MCP results expose keywords and year alongside existing fields.
- No contributor lookup, PDF download, search-history mutation, deduplication,
  screening, or synthesis occurs implicitly.

## Deferred Or Rejected Scope

CoLRev's Git-backed review projects, search-history files, contributor
resolution, BibDedupe integration, PDF retrieval/preparation, screening,
quality appraisal, and PRISMA reporting are separate workflows rather than
OSF API primitives. They remain deferred to future focused tracks; embedding a
Python literature-review runtime would violate this repository's CLI-first,
cross-platform, offline-testable boundary.
