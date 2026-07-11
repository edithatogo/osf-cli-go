# Live Validation Matrix

Live checks are opt-in and require explicit credentials or a disposable test
project. They must never run as ordinary unit tests or in pull requests.

| Area | Required scenario | Evidence | Status |
|---|---|---|---|
| Authentication | `whoami` with a disposable token | Redacted command/result and account-independent status | opt-in |
| Public read | Get a known public node without a token | Node metadata and HTTP status | opt-in |
| Private read | List owned projects and contributors | Redacted IDs and result counts | opt-in |
| Files | List, download, and verify a fixture file/tree | Checksums and manifest | opt-in |
| Safe writes | Upload or create a folder in a disposable project | Project-scoped evidence and cleanup | opt-in |
| Destructive writes | Delete only a disposable fixture after explicit confirmation | Confirmation and cleanup evidence | opt-in |
| Search | Search OSF and list preprints | Query, pagination, and result-shape evidence | opt-in |
| Registrations | Create a draft only, never publish | Draft identifier and cleanup decision | opt-in |
| Failure behavior | Invalid token, rate limit, missing node, and network timeout | Redacted typed errors | fixture + opt-in |

The release checklist must identify which rows were run for a release candidate
and which remain unrun because credentials, permissions, or a disposable OSF
fixture were unavailable.
