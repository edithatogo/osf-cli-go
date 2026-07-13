# OSF API schema discovery phase review

Reviewed 2026-07-14.

## Specification audit

- Audited the official OSF developer repository and pinned
  `swagger-spec/swagger.yaml` to commit
  `52215db60cc8dd95f86841b467d2ed339fd67dec`.
- Benchmarked `hirakinii/osf-api-mcp` as a separate Apache-2.0 TypeScript MCP
  implementation without copying its generated or incompatible surfaces.

## Design and implementation

- Added the license-aware, dated source manifest and offline validator.
- Added an optional online fetch check for the pinned source URL.
- Deferred generic schema search and generated-client replacement to issue #46,
  with the rationale documented in `docs/osf-api-schema-discovery.md`.

## Closeout

- Release validation requires the source manifest and its pinned deferred
  decision fields.
- The feature matrix records schema discovery as prepared pending a separately
  reviewed generated or vendored contract.
- Full tests, race, vet, lint, vulnerability, stub, review, matrix, registry,
  release, and schema-manifest gates passed.
