# OSF API entity coverage

## Objective

Close the remaining read-only entity gap identified by GitHub issue #80.
Expose file versions, node add-ons, wiki pages, comments, logs, and identifiers
through stable typed API, CLI, and MCP contracts.

## Acceptance criteria

- API client methods return typed records for every covered endpoint.
- CLI commands provide deterministic table and JSON output.
- MCP tools expose bounded read-only structured results with stable identifiers.
- Unit tests cover endpoint paths, command contracts, tool registration, and
  structured output behavior.
- Published MCP inventories and the feature matrix match the implementation.
