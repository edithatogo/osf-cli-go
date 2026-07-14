# PresQT OSF parity

Last reviewed: 2026-07-14

This comparison uses the maintained [PresQT documentation](https://presqt.readthedocs.io/), especially its [service overview](https://presqt.readthedocs.io/en/latest/), [transfer metadata contract](https://presqt.readthedocs.io/en/latest/web_services.html), and [target integration contract](https://presqt.readthedocs.io/en/latest/target_integration.html). The upstream GitHub page was not used as the sole evidence source because it was unavailable during review.

## Capability comparison

| Capability | PresQT reference | OSF CLI Go behavior | Decision |
|---|---|---|---|
| OSF collection/search/detail/download/upload | REST service exposes OSF target collection, search, detail, download, and upload actions | OSF project/search, file listing, explicit download, and upload commands/API surfaces | Existing OSF capabilities implemented; no service wrapper added |
| Cross-repository transfer | Transfers between OSF and GitHub, Zenodo, GitLab, Figshare, and other targets | OSF-only client boundary | Deliberately out of scope; separate provider integrations would materially expand scope |
| Fixity | PresQT documents OSF checksums and target-specific hash behavior | OSF provider-supplied checksum metadata and explicit transfer manifests | Existing integrity contract sufficient for OSF-only operations |
| Keyword mapping | Maps target-specific tags/topics/keywords during transfers | No cross-target metadata mapping | Deferred; requires a provider-neutral metadata model and explicit write semantics |
| FTS metadata | Writes/extends `PRESQT_FTS_METADATA.json` with action history, source/destination, timestamps, and keywords | Deterministic transfer manifests, but no cross-repository provenance document | Deferred; issue #19 tracks a future provenance/metadata contract |
| FAIR and preservation services | FAIRness evaluation, keyword enhancement, EaaSI, and preservation workflows | OSF metadata validation without claiming scientific or FAIR certification | Deliberately out of scope; preserve validation boundary |
| Authentication and deployment | REST target tokens and service deployment model | OSF token/password fallback, redaction, explicit local CLI/MCP execution | Implemented for OSF; no hosted PresQT dependency |

## Decision

PresQT is a preservation and interoperability service, not a drop-in OSF CLI
competitor. Its useful OSF-specific fixity and transfer-history concepts are
represented by existing OSF checksums and explicit manifests. Cross-repository
transfer, target-specific keyword mapping, and `PRESQT_FTS_METADATA.json` are
deferred because adopting them would require a new provider-neutral metadata and
write contract. No credentials, hosted-service claims, network tests, or synthetic
preservation results were added.
