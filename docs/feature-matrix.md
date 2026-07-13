# OSF CLI Go feature matrix

Last reviewed: 2026-07-13

Status meanings: **Implemented** is covered by repository code and tests;
**Prepared** means a distribution artifact exists but provider-side approval is
not evidenced; **Track** means the capability is intentionally planned; and
**External gate** means local work is complete but authentication or maintainer
action remains.

| Area | CLI | API client | MCP | Safety/quality contract | Status and next action |
|---|---|---|---|---|---|
| Authentication and identity | `auth login`, `auth whoami`, `whoami` | token and username/password credentials | `osf_whoami` | redaction, no credential persistence, explicit auth modes | Implemented; live validation remains opt-in |
| Projects: list/get | `projects list`, `projects get` | current projects, node get | `osf_projects_list`, `osf_project_get` | URL/id normalization, JSON output | Implemented |
| Projects: create/update/delete | `projects create`, `projects update`, `projects delete` | create, patch, delete node | intentionally not exposed | confirmation for destructive actions; MCP remains read-only | Implemented in CLI; MCP write-safety track needed |
| Components | `components list` | child node pagination | `osf_components_list` | deterministic pagination | Implemented |
| Contributors | API-backed CLI surface | contributor listing | `osf_contributors_list` | stable structured fields | Implemented |
| Search | `search` | OSF search pagination and limit | `osf_search` | required query, max limit 100 | Implemented; SourceShift parity landed |
| Preprints | `preprints list` | provider filter and limit | `osf_preprints_list` | bounded result sets | Implemented |
| Registrations | `registrations create` | draft registration creation | not exposed | authenticated write, no publish action | CLI implemented; MCP read/write expansion tracked |
| Storage listing | `files list` | provider and folder traversal | `osf_files_list` | traversal protection and structured metadata | Implemented |
| File download | `files download` | streaming download | not exposed | atomic writes, conflicts, manifests, path safety | CLI implemented; MCP download boundary tracked |
| File upload | `files upload` | WaterButler upload | not exposed | conflict policy, auth, explicit write | CLI implemented; MCP write track needed |
| Folder creation | `files mkdir` | WaterButler folder create | not exposed | path validation and auth | CLI implemented |
| File deletion | `files rm` | WaterButler delete | not exposed | confirmation and auth | CLI implemented |
| File versions | no dedicated CLI command | version listing | not exposed | stable typed model | API implemented; CLI/MCP exposure track |
| Add-ons | `files addons` | node add-on listing | not exposed | read-only | CLI/API implemented; MCP expansion track |
| Wiki pages | no dedicated CLI command | wiki listing | not exposed | read-only | API implemented; CLI/MCP expansion track |
| Comments/logs/identifiers | no dedicated CLI command | entity listing | not exposed | read-only | API implemented; entity parity tracks |
| Export | `export` | composed node snapshot | not exposed | deterministic JSON and pagination | CLI implemented; MCP export boundary tracked |
| Shell integration | `completion bash/zsh/fish/powershell`, `open` | n/a | stdio transport | supported-platform contract | Implemented |
| Output | table and `--json` | typed Go models | structured MCP content | stable schemas and errors | Implemented; compatibility harness tracked |
| Pagination/retries | transparent API pagination | collection helpers | bounded tool limits | cancellation and deterministic behavior | Implemented; fuzz/performance hardening tracked |
| Live OSF validation | opt-in validation tool | real API requests | real MCP calls | no committed credentials, safe writes only | Prepared; `OSF_TOKEN` and validation project required |
| Release artifacts | multi-platform binaries and checksums | dynamic versioning | OCI MCP image and MCPB | SBOM, provenance, Cosign, release gates | Implemented at 0.3.2; 1.0 hardening track |
| Package managers | Homebrew tap, Scoop bucket, WinGet PR | n/a | n/a | install/version verification | Homebrew/Scoop published; WinGet external review |
| Official MCP Registry | n/a | n/a | `server.json` and OCI package | registry contract and release workflow | Published; maintain on release |
| Smithery/Glama/MCP.Directory | n/a | n/a | MCPB/metadata packets | provider-specific metadata | Published or prepared; maintain evidence |
| Codex/OpenAI | plugin and MCP config | n/a | stdio MCP | marketplace validator and authenticated review | Prepared; track `codex-marketplace-adoption_20260706` |
| Claude/Anthropic | plugin and marketplace metadata | n/a | stdio MCP | official submission form and plugin validation | Prepared; submission pending |
| GitHub Copilot | plugin, skill, repository MCP | n/a | repository config | marketplace/repository requirements | Prepared; provider approval not evidenced |
| Cursor/Cline/LobeHub | configs, plugin artifacts, MCPB | n/a | stdio/remote options | client install validation and listing receipts | Tracks `cursor`, `cline`, and `lobehub` |
| Gemini/Qwen | extension manifests and packages | n/a | embedded MCP config | version alignment and gallery rules | Prepared/published artifacts; indexing evidence pending |
| Ecosystem parity | 13 source-tool tracks | n/a | competitor comparison | dated source evidence and explicit gap decisions | SourceShift partial complete; remaining tracks open |

## Matrix rules

1. A row cannot be marked implemented from documentation alone; it needs code
   and deterministic validation.
2. A registry cannot be marked published from a local packet; provider-side
   receipt or public listing evidence is required.
3. A feature that performs a write must document authorization, confirmation,
   rollback, and live-test cleanup before MCP exposure.
4. Every release updates this matrix, the feature inventory, and the registry
   scorecard together.
