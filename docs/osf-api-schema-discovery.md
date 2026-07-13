# OSF API schema discovery

Reviewed 2026-07-14. The source manifest is
[`osf-api-schema-source.json`](osf-api-schema-source.json) and is validated by
`go run ./tools/checkapischema`.

## Source comparison

| Source | Evidence | Decision |
|---|---|---|
| Official OSF API v2 documentation | [CenterForOpenScience/developer.osf.io](https://github.com/CenterForOpenScience/developer.osf.io), pinned to commit `52215db60cc8dd95f86841b467d2ed339fd67dec` and `swagger-spec/swagger.yaml` | Authoritative discovery source; use the manifest and optional online refresh check |
| `osf-api-mcp` | [hirakinii/osf-api-mcp](https://github.com/hirakinii/osf-api-mcp), Apache-2.0, TypeScript MCP server with its own `schema/` and `src/` surfaces | Benchmark only; do not copy incompatible implementation or replace the Go client |

## Decision

The repository keeps its typed Go client and explicit MCP contract as the
runtime surface. It does not expose a generic schema-search tool or generate a
new client from the whole Swagger document in this track. That would create a
large, unstable public surface and make the release contract dependent on
unreviewed generated code.

The source manifest records provenance, license, retrieval date, pinned commit,
tracked resource tags, and the linked follow-up issue. `checkapischema` validates
that manifest offline; `-online` additionally fetches the pinned source and
checks that it is a non-empty Swagger document. Updating the pinned commit is a
deliberate reviewable change.
