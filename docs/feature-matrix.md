# OSF CLI Go feature matrix

Last reviewed: 2026-07-13

This generated matrix is backed by `docs/feature-matrix.json`. Status meanings: **external-gate** is Local work is complete, but authentication or maintainer action remains.; **implemented** is Covered by repository code and deterministic tests.; **prepared** is A local distribution artifact exists, but provider-side approval is not evidenced.; **track** is Intentionally planned and linked to follow-up work.

| Area | CLI | API client | MCP | Safety/quality contract | Status | Next action | Track | Issue |
|---|---|---|---|---|---|---|---|---|
| Authentication and identity | auth login, auth whoami, whoami | token and username/password credentials | osf_whoami | redaction; no credential persistence; explicit auth modes | implemented | Run opt-in live validation | `conductor/tracks/live-osf-validation_20260502/spec.md` |  |
| Projects: list/get | projects list, projects get | current projects; node get | osf_projects_list, osf_project_get | URL/id normalization; JSON output | implemented |  | `` |  |
| Projects: create/update/delete | projects create, projects update, projects delete | create; patch; delete node | intentionally not exposed | confirmation for destructive actions; MCP remains read-only | track | Design MCP write authorization and rollback | `conductor/tracks/mcp-server-roadmap_20260502/spec.md` | #21 |
| Components | components list | child node pagination | osf_components_list | deterministic pagination | implemented |  | `` |  |
| Contributors | API-backed CLI surface | contributor listing | osf_contributors_list | stable structured fields | implemented |  | `` |  |
| Search | search | OSF search pagination and limit | osf_search | required query; maximum limit 100 | implemented | Maintain SourceShift parity evidence on releases | `conductor/archive/sourceshift-osf-mcp-parity_20260711/spec.md` | #8 |
| Literature metadata export | search --bibtex | OSF search title, abstract, tags, year, URL, and stable ID | osf_search structured metadata | deterministic escaped BibTeX; no contributor lookup or PDF download side effect | implemented | Maintain bibliographic field compatibility | `conductor/archive/colrev-osf-parity_20260711/spec.md` | #17 |
| Preprints | preprints list | provider filter and limit | osf_preprints_list | bounded result sets | implemented |  | `` |  |
| Preprint discovery | preprints search | title and provider filters; publication metadata | osf_preprints_search | required query; limit 1-100; read-only structured results | implemented | Maintain OSF preprint filter compatibility | `conductor/archive/tooluniverse-osf-parity_20260711/spec.md` | #15 |
| Registrations | registrations create | draft registration creation | not exposed | authenticated write; no publish action | track | Add read-only MCP registration tools first | `conductor/tracks/osf-api-coverage_20260502/spec.md` | #20 |
| Storage listing | files list | provider and folder traversal | osf_files_list | traversal protection; structured metadata | implemented |  | `` |  |
| Storage integrity metadata | files list --json | optional attributes.extra.hashes.md5 | osf_files_list md5 | provider-supplied checksum only; no implicit download | implemented | Maintain checksum compatibility | `conductor/archive/jasp-osf-integration-parity_20260711/spec.md` | #18 |
| File download | files download | streaming download | not exposed | atomic writes; conflicts; manifests; path safety | track | Approve MCP download resource boundary | `conductor/tracks/mcp-server-roadmap_20260502/spec.md` | #21 |
| File upload | files upload | WaterButler upload | not exposed | conflict policy; auth; explicit write | track | Design MCP write confirmation | `conductor/tracks/mcp-server-roadmap_20260502/spec.md` | #21 |
| Folder creation | files mkdir | WaterButler folder create | not exposed | path validation and auth | implemented |  | `` |  |
| File deletion | files rm | WaterButler delete | not exposed | confirmation and auth | implemented |  | `` |  |
| File versions | no dedicated command | version listing | not exposed | stable typed model | track | Expose read-only versions consistently | `conductor/tracks/osf-api-coverage_20260502/spec.md` | #80 |
| Add-ons | files addons | node add-on listing | not exposed | read-only | track | Add MCP read-only add-on tool | `conductor/tracks/osf-api-coverage_20260502/spec.md` | #80 |
| Wiki pages | no dedicated command | wiki listing | not exposed | read-only | track | Add CLI and MCP wiki read tools | `conductor/tracks/osf-api-coverage_20260502/spec.md` | #80 |
| Comments, logs, identifiers | no dedicated commands | entity listing | not exposed | read-only | track | Prioritize entity parity by user demand | `conductor/tracks/osf-api-coverage_20260502/spec.md` | #80 |
| Export | export | composed node snapshot | not exposed | deterministic JSON and pagination | track | Define MCP export size and resource limits | `conductor/tracks/mcp-server-roadmap_20260502/spec.md` | #21 |
| OSF metadata validation | validate --profile research-output|preregistration | node, contributor, and storage metadata | not exposed | read-only deterministic findings; no LLM or scientific validity claim | implemented | Maintain finding schema compatibility | `conductor/archive/metacheck-osf-validation_20260711/spec.md` | #20 |
| Shell integration | completion bash/zsh/fish/powershell; open | n/a | stdio transport | supported-platform contract | implemented |  | `` |  |
| Output | table and --json | typed Go models | structured MCP content | stable schemas and errors | track | Add compatibility regression harness | `conductor/tracks/mcp-quality-evaluation-harness_20260713/spec.md` | #54 |
| Pagination, retries, cancellation | transparent | collection helpers | bounded tool limits | deterministic behavior and cancellation | track | Add fuzz and performance campaigns | `conductor/archive/v1-hardening-maturity_20260713/spec.md` | #52 |
| Live OSF validation | opt-in validation tool | real API requests | real MCP calls | no committed credentials; safe writes only | prepared | Set OSF_TOKEN and OSF_VALIDATE_PROJECT | `conductor/tracks/live-osf-validation_20260502/spec.md` |  |
| Release artifacts | multi-platform binaries and checksums | dynamic versioning | OCI MCP image and MCPB | SBOM; provenance; Cosign; release gates | implemented | Complete 1.0 launch gates | `conductor/archive/v1-hardening-maturity_20260713/spec.md` | #52 |
| Package managers | Homebrew tap; Scoop bucket; WinGet PR | n/a | n/a | install and version verification | external-gate | WinGet maintainer review | `conductor/archive/winget-adoption_20260706/spec.md` |  |
| Official MCP Registry | n/a | n/a | server.json and OCI package | registry contract and release workflow | implemented | Maintain on every release | `conductor/archive/official-mcp-github-registry-adoption_20260706/spec.md` |  |
| Smithery, Glama, MCP.Directory | n/a | n/a | MCPB and metadata packets | provider-specific metadata and receipts | prepared | Refresh evidence on release | `conductor/tracks/registry-submission-scorecard_20260713/spec.md` | #53 |
| Codex and OpenAI Cowork | plugin and MCP config | n/a | stdio MCP | marketplace validation and authenticated review | prepared | Submit through current OpenAI surface | `conductor/archive/codex-marketplace-adoption_20260706/spec.md` |  |
| Claude and Anthropic Cowork | plugin and marketplace metadata | n/a | stdio MCP | official submission form and plugin validation | prepared | Submit through official Claude form | `conductor/archive/claude-official-plugin-directory-adoption_20260706/spec.md` |  |
| GitHub Copilot | plugin; skill; repository MCP | n/a | repository config | marketplace and repository requirements | prepared | Verify provider-side approval | `conductor/tracks/registry-submission-scorecard_20260713/spec.md` | #53 |
| Cursor, Cline, LobeHub | configs; plugin artifacts; MCPB | n/a | stdio and remote options | clean-client install validation and receipts | track | Execute provider-specific submissions | `conductor/tracks/registry-submission-scorecard_20260713/spec.md` | #53 |
| Gemini and Qwen | extension manifests and packages | n/a | embedded MCP config | version alignment and gallery rules | prepared | Verify indexing evidence | `conductor/tracks/registry-submission-scorecard_20260713/spec.md` | #53 |
| DataLad interoperability | export; files download --tree | OSF project and storage primitives | read-only project/file tools | no implicit Git, DataLad, or git-annex state; safe local writes | track | Define optional git-annex and Git remote interoperability contract | `conductor/archive/datalad-osf-parity_20260711/spec.md` | #69 |
| DOI-to-OSF resolution | resolve | DOI redirect and OSF host validation | osf_doi_resolve | strict DOI forms; no non-OSF destinations; no download side effect | implemented | Maintain resolver redirect compatibility | `conductor/archive/datahugger-doi-parity_20260711/spec.md` | #14 |
| OSF ecosystem parity | 13 source-tool tracks | competitor comparison | dated source evidence | each gap implemented, deferred, or rejected | track | Complete parity tracks and close SourceShift | `conductor/archive/feature-matrix_20260713/spec.md` | #51 |

## Matrix rules

1. A row cannot be marked implemented from documentation alone; it needs code and deterministic validation.
2. A registry cannot be marked published from a local packet; provider-side receipt or public listing evidence is required.
3. A feature that performs a write must document authorization, confirmation, rollback, and live-test cleanup before MCP exposure.
4. Every release updates this matrix, the feature inventory, and the registry scorecard together.
