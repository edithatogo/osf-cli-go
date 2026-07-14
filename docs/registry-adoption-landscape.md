# Registry adoption landscape

Last reviewed: 2026-07-14

This matrix records public submission routes separately from local readiness,
provider review, and publication. Costs are the observed submission cost for
the documented route; hosting, paid promotion, and provider-specific plans are
not implied. No row is treated as published without provider-side evidence.

| Target | Source and route | Auth/ownership | Cost | Review/publication | Listing URL | Status | Next action |
|---|---|---|---|---|---|---|---|
| [Cline MCP Marketplace](https://github.com/cline/mcp-marketplace) | GitHub issue template; requires logo, reason, and Cline install validation | GitHub account; maintainer review | Free | Curated issue review | Not listed | blocked: packet incomplete | Track #47: add the 400x400 PNG, install transcript, and submit |
| [LobeHub MCP Marketplace](https://market.lobehub.com/s/publish-mcp) | `@lobehub/market-cli`; `lhm plugin submit`/`publish` | LobeHub login plus GitHub ownership connection | Free route observed | Ownership claim and marketplace review | Not listed | blocked: manifest incomplete | Track #48: add `lhm.plugin.json`, authenticate, submit, and verify |
| [MCP.so](https://mcp.so/submit) | Public submission form/issue route | Public repository; account may be requested at submission | Free route observed | Provider indexing/review | Not listed | prepared | Submit the existing packet and record the receipt |
| [mcpservers.org](https://mcpservers.org/submit) | Website submission form | Public repository; account requirements not evidenced | Free route observed; paid review is optional | Directory review; paid promotion is separate | Not listed | prepared | Submit without purchasing promotion and record the listing |
| [MCPize](https://docs.mcpize.com/) | Account or GitHub App deployment workflow | Account and deployment ownership required | Not established | Platform deployment/review | Not listed | blocked: account/deployment gate | Authenticate only when ready; confirm deployment and pricing first |
| [Docker MCP Catalog](https://docs.docker.com/ai/mcp-catalog-and-toolkit/catalog/) | Pull request to [`docker/mcp-registry`](https://github.com/docker/mcp-registry) | GitHub account; repository CI and maintainer review | Free | Automated checks plus maintainer review | Not listed | prepared | Open the upstream PR and record CI/review evidence |
| [PulseMCP](https://www.pulsemcp.com/submit) | Website form or official-registry ingestion | Access/auth requirement unresolved | Not established | Provider ingestion/review | Not listed | blocked: access denied | Use the published contact route or retry in an authenticated browser |
| MCP Market | Website availability observed; official submission contract not verified | Unknown | Unknown | Unknown | Not verified | deprioritized | Do not submit until an official route and ownership model are verified |
| [MCP Central](https://mcpcentral.io/servers) | Aggregator endpoint observed; submission contract not verified | Unknown | Unknown | Unknown | Not verified | watch | Reassess only after an official submission/API contract appears |
| [mcp-reg.com](https://mcp-reg.com/) | Self-hosted registry template, not a downstream public marketplace | Repository owner operates the instance | Self-hosting cost | Instance-owner controlled | Not applicable | deprioritized | Do not spend submission effort; it is useful only for private registry deployments |
| [Official MCP Registry](https://modelcontextprotocol.io/registry/about) | `mcp-publisher` metadata publication | Namespace authentication; GitHub-backed namespace | Free registry route | Schema/namespace validation | [OSF entry](https://registry.modelcontextprotocol.io/v0.1/servers?search=io.github.edithatogo%2Fosf-cli-go) | published | Republish and verify on every release |
| [Smithery](https://smithery.ai/docs/build/publish) | MCPB or hosted-server publish flow | Account; deployment ownership | Free publication route observed; hosting plan may apply | Automated build/health checks and provider review | [OSF release](https://smithery.ai/servers/edithatogo/osf-cli-go/releases) | published | Maintain deployment and tool inventory |
| [Glama](https://glama.ai/) | GitHub indexing/claim flow | GitHub connection for ownership claim | Free listing; paid hosting is separate | Automated indexing and quality/security scans | [OSF listing](https://glama.ai/mcp/servers/edithatogo/osf-cli-go) | published | Refresh release and usage evidence without fabricating usage |
| Microsoft Commercial Marketplace | [Marketplace Ingestion MCP](https://learn.microsoft.com/en-us/partner-center/marketplace-offers/ingestion-mcp) and Partner Center offer workflow | Partner Center account, seller tenant, and required role | Commercial marketplace terms apply; not an open MCP directory | Offer schema validation, certification, and publication | Not applicable | out of scope | Revisit only if OSF CLI Go becomes a commercial offer; this is not an open MCP listing route |

## Prioritization

High-value open submission targets with dedicated work are Cline (#47),
LobeHub (#48), and Docker MCP Catalog (the archived catalog sweep and its
prepared packet). MCP.so and mcpservers.org are worthwhile public follow-ups.
MCPize and PulseMCP require external access or deployment decisions. MCP Market,
MCP Central, and mcp-reg.com are explicitly deprioritized until their public
submission contracts become verifiable.

The matrix intentionally contains no credentials, usage events, synthetic
receipts, or approval claims.

## Packet and track map

| Target | Local packet/evidence | Dedicated track or follow-up |
|---|---|---|
| Cline | `integrations/cline/`, release MCPB artifacts; packet still incomplete | `conductor/tracks/cline-mcp-marketplace-adoption_20260713/` and issue #47 |
| LobeHub | `registry/directory-submissions.json`; `lhm.plugin.json` still missing | `conductor/tracks/lobehub-mcp-marketplace-adoption_20260713/` and issue #48 |
| Docker MCP Catalog | `registry/docker-mcp-registry/`, `docs/mcp-catalog-discoverability-evidence.md` | archived catalog sweep; submit upstream PR |
| MCP.so and mcpservers.org | `docs/mcp-catalog-discoverability-evidence.md`, canonical metadata packet | submit manually and record receipts |
| MCPize | canonical metadata and Docker/MCPB assets | authenticate and confirm deployment requirements before submission |
| MCP.Directory | archived submission evidence and `registry/directory-submissions.json` | retain pending-review state |
| PulseMCP | official registry metadata and access-denied evidence | contact provider or retry authenticated route |
