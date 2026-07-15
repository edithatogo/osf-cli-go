# Zenodo REST fixtures

These redacted fixtures model the published-record response shapes documented
at <https://developers.zenodo.org/> and reviewed in
`docs/zenodo-api-source.json` on 2026-07-15.

- `records_page1.json` uses the current `hits.hits` search envelope and
  `files.entries` collection described by the official bucket-oriented examples.
- `records_page2.json` includes the documented legacy file-array representation
  so the read adapter remains tolerant while preserving the native record JSON.
- `record_1001.json` includes an unknown provider extension to prove that typed
  discovery fields do not discard upstream metadata.

The values are synthetic and contain no account, token, private record, or live
user data. Routine tests must not refresh these fixtures from the network.
