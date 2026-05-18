# OSF API Coverage Research Matrix

## Contract

This matrix is the repo-local contract for the OSF API coverage track. It is deliberately conservative: it maps stable API areas needed by this CLI, records whether comparable tools expose those areas, and avoids live OSF calls during validation. Future endpoint additions should update this file before implementation so gaps are explicit.

Evidence basis:

- `internal/osfapi/client.go` and fixture-backed tests in `internal/osfapi/client_test.go`
- Current CLI surfaces in `internal/cli`
- Publicly documented behavior of osfclient, osfr, and osf-project-exporter as research inputs for coverage shape, not as executable dependencies

Legend:

- `Y`: supported in the tool or this repo
- `Partial`: supported for a narrower workflow or with less metadata
- `N`: no comparable first-class support identified
- `Out of scope`: excluded by this track's spec

## Endpoint Matrix

| OSF API area | Endpoint family | osf-cli-go | osfclient | osfr | osf-project-exporter | Notes |
| --- | --- | --- | --- | --- | --- | --- |
| Current user | `/v2/users/me/` | Y | Y | Y | N | CLI exposes `auth whoami` and `whoami`. |
| User profile | `/v2/users/{id}/` | Y | Partial | Partial | N | API client support exists; no dedicated CLI command. |
| User nodes | `/v2/users/me/nodes/` | Y | Y | Y | Partial | CLI exposes authenticated `projects list`. |
| Node metadata | `/v2/nodes/{id}/` | Y | Y | Y | Y | CLI exposes `projects get` and export. |
| Node children | `/v2/nodes/{id}/children/` | Y | Y | Y | Y | CLI exposes `components list` and export. |
| Node contributors | `/v2/nodes/{id}/contributors/` | Y | Y | Partial | Y | Export uses this surface. |
| Node files | `/v2/nodes/{id}/files/osfstorage/` | Y | Y | Y | Y | CLI exposes `files list`, `files download`, and export. |
| File metadata | `/v2/files/{id}/` | Y | Y | Y | Partial | CLI download resolves file IDs through metadata. |
| File versions | `/v2/files/{id}/versions/` | Y | Partial | Partial | Partial | API client support exists; no dedicated CLI command. |
| Registrations list | `/v2/nodes/{id}/registrations/` | Y | Partial | Partial | Partial | API client support exists; registration creation has CLI coverage. |
| Wiki pages | `/v2/nodes/{id}/wikis/` | Y | Partial | Partial | Partial | API client support exists; no dedicated CLI command. |
| Comments | `/v2/nodes/{id}/comments/` | Y | N | N | N | API client support exists; no dedicated CLI command. |
| Logs | `/v2/nodes/{id}/logs/` | Y | N | N | Partial | API client support exists; no dedicated CLI command. |
| Identifiers | `/v2/nodes/{id}/identifiers/` | Y | N | Partial | Partial | API client support exists; no dedicated CLI command. |
| Storage add-ons | `/v2/nodes/{id}/addons/` | Y | Partial | Partial | Partial | CLI exposes `files addons`. |
| Preprints | `/v2/preprints/` | Y | N | Partial | N | Existing adjacent implementation, although preprints are a non-goal for this track. |
| Search | `/v2/search/` | Y | N | Partial | N | Existing adjacent implementation. |
| Create registration | `/v2/nodes/{id}/registrations/` POST | Y | N | Partial | N | CLI requires typed confirmation unless `--yes` is supplied. |
| Create node | `/v2/nodes/` POST | Y | Partial | Y | N | CLI requires typed confirmation unless `--yes` is supplied. |
| Update node | `/v2/nodes/{id}/` PATCH | Y | Partial | Y | N | CLI preserves omitted fields and requires typed confirmation unless `--yes` is supplied. |
| Delete node | `/v2/nodes/{id}/` DELETE | Y | Partial | Y | N | CLI requires typed confirmation unless `--yes` is supplied. |
| Upload file | WaterButler `PUT` | Y | Y | Y | N | CLI exists through `files upload`; live writes remain opt-in only. |
| Create folder | WaterButler `PUT kind=folder` | Y | Y | Y | N | CLI exists through `files mkdir`; live writes remain opt-in only. |
| Delete file | WaterButler `DELETE` | Y | Y | Y | N | CLI requires typed confirmation unless `--yes` is supplied. |
| OAuth/auth flows | OAuth/token management | Out of scope | Y | Y | N | Excluded by spec non-goals. |
| Bulk upload/sync | Batch write workflows | Out of scope | Partial | Partial | N | Excluded by spec non-goals. |
| Meetings API | OSF Meetings | Out of scope | N | N | N | Excluded by spec non-goals. |

## Implementation Contract

- New API endpoints must have fixture-backed `internal/osfapi` tests.
- New write CLI commands must require typed `yes` confirmation, with `--yes` reserved for explicit automation.
- Live OSF validation is not part of this contract; use mock servers or fake clients unless a later track explicitly approves a disposable live project and token.
- CLI docs must document any new command, confirmation behavior, and destructive behavior before the track is marked complete.
