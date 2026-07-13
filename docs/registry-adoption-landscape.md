# Registry adoption landscape

Last reviewed: 2026-07-13

This matrix distinguishes a public submission path from a listing that is
actually published. Status is based on the linked provider documentation or
submission surface.

| Target | Submission path | Current assessment | Next action |
|---|---|---|---|
| [Cline MCP Marketplace](https://github.com/cline/mcp-marketplace) | GitHub issue template | High-value, public, curated; requires 400x400 logo, reason, and Cline install validation | Track #47 |
| [LobeHub MCP Marketplace](https://market.lobehub.com/s/publish-mcp) | `@lobehub/market-cli` plus browser login and GitHub ownership connection | High-value, owner-authenticated, documented | Track #48 |
| [MCP.so](https://mcp.so/submit) | Public GitHub repository form | Public submission path confirmed; draft is published when saved | Existing packet; submit and record receipt |
| [mcpservers.org](https://mcpservers.org/submit) | Website form | Public free submission; premium review is optional and not required | Existing packet; submit without paid upgrade |
| [MCPize](https://mcpize.com) | Account or platform workflow | Active marketplace, but account and deployment requirements need confirmation | Existing adoption track |
| [Docker MCP Catalog](https://docs.docker.com/ai/mcp-catalog-and-toolkit/catalog/) | Pull request to `docker/mcp-registry` | Public repository workflow | Existing prepared packet |
| [PulseMCP](https://www.pulsemcp.com/submit) | Website form | Access denied from this environment; no submission claimed | Contact or retry in authenticated browser |
| MCP Market | Website availability observed, submission flow not verified | Low confidence until official submission documentation is found | Do not submit yet |
| MCP Central | Directory endpoint was not verifiable during this review | Low confidence | Do not submit yet |

## Additional ecosystem observations

- [Paperclip](https://github.com/matsjfunke/paperclip) is an archived MIT MCP
  server with useful multi-provider research retrieval patterns; it is tracked
  separately and is not treated as a current hosted service.
- [osf-api-mcp](https://github.com/hirakinii/osf-api-mcp) indexes 250 OSF API
  endpoints and 40 tags from an OpenAPI snapshot; it is tracked as a possible
  offline API-discovery capability rather than an OSF data client replacement.
- [Cline's submission README](https://github.com/cline/mcp-marketplace#how-to-submit-your-mcp-server)
  explicitly allows projects with few stars, but still applies project,
  installation, and security review.

Low-confidence or inaccessible targets remain in the matrix for follow-up, not
as claimed distribution channels.
