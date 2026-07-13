# Registry and coding-agent scorecard

Last reviewed: 2026-07-13

This is an execution scorecard, not a claim that any provider has awarded a
score. The target is 100% of applicable requirements; `N/A` must have a written
rationale. Provider review and approval remain external facts.

## Universal scoring rubric

| Criterion | Weight | Evidence required |
|---|---:|---|
| Correct public repository, license, ownership, and release version | 15 | public URLs and version alignment check |
| Install path works from a clean environment | 20 | rerunnable install transcript for the provider |
| Metadata, descriptions, categories, and tool inventory | 15 | provider packet plus schema checks |
| Security and permissions disclosure | 15 | auth variables, read/write classification, redaction and threat-model links |
| Tests, support, compatibility, and maintenance evidence | 15 | CI, release, support policy, issue/response links |
| User-facing assets and examples | 10 | logo/screenshots/examples where required |
| Provider submission and public verification | 10 | receipt, review state, listing URL, or exact blocker |

## Target matrix

| Target | Local readiness | External action | Score target |
|---|---|---|---:|
| OpenAI Codex/Cowork | plugin, MCP config, release archives | authenticate and submit through the current OpenAI surface | 100% |
| Anthropic Claude/Cowork | validated Claude plugin and marketplace metadata | submit at official Claude form; record review/listing | 100% |
| GitHub Copilot | plugin, skill, repository MCP config | verify repository marketplace and any maintained default marketplace route | 100% |
| Cursor | `.cursor/mcp.json`, MCPB, README install path | submit/verify Cursor Directory | 100% |
| Cline | MCPB, logo, README/`llms-install.md` validation | submit official Cline issue | 100% |
| LobeHub | `lhm.plugin.json`, release metadata | browser login, GitHub connect, submit/publish with `lhm` | 100% |
| Gemini CLI | extension manifests and version checks | verify gallery indexing and record listing | 100% |
| Qwen Code | extension manifests and package | verify supported gallery or document repository install | 100% |
| Official MCP Registry | `server.json`, OCI image, release workflow | maintain each release publication | 100% |
| Smithery | MCPB, published deployment, MCP URL | maintain release and tool inventory | 100% |
| Glama | `glama.json`, cross-links, release evidence | claim/update listing and verify usage/listing | 100% |
| MCP.Directory | prepared metadata and submission packet | authenticate/manual submission and verify | 100% |
| Docker MCP Catalog | Docker image and registry packet | open/maintain upstream PR | 100% |
| MCP.so | public GitHub repository | submit form and complete draft | 100% |
| mcpservers.org | public repository and category metadata | free submission and verify listing | 100% |
| MCPize | public package and deployment metadata | confirm account/deployment requirements before submit | 100% |
| PulseMCP/MCP Central/MCP Market | packet only | verify official route before spending effort | N/A until verified |

## Submission rules

- Never fabricate usage, score, approval, or listing state.
- Use provider-side validation whenever available, then store a dated receipt.
- Use browser authentication only for the final submission step and ask the
  user to log in when the session is unavailable.
- Re-run the scorecard after every release and after every provider response.
