# Agent distribution status

Last reviewed: 2026-07-13

| Surface | Local artifact | External status | Track |
|---|---|---|---|
| OpenAI Codex | `plugins/codex-osf` and `.agents/plugins/marketplace.json` | Prepared; public marketplace review not yet evidenced | `codex-marketplace-adoption_20260706` |
| OpenAI Cowork | MCP server and MCPB package | Prepared; no separate approved public listing evidenced | `codex-marketplace-adoption_20260706` |
| Anthropic Claude Code/Cowork | `plugins/claude-osf` and `.claude-plugin/marketplace.json` | Package validated; official directory submission blocked by authentication on 2026-07-13 | `claude-official-plugin-directory-adoption_20260706` |
| GitHub Copilot CLI/cloud agent | `plugins/github-copilot-osf`, `.github/plugin/marketplace.json`, skill, and MCP configs | Repository-hosted marketplace available; GitHub-maintained default marketplace approval not evidenced | `github-copilot-plugin-publication_20260711` |
| Gemini CLI | root `gemini-extension.json` and `plugins/gemini-osf` release package | Public repository is gallery-discoverable via `gemini-cli-extension`; gallery indexing/approval not yet evidenced | `gemini-extension-gallery-publication_20260711` |
| Qwen Code | root `qwen-extension.json` and `plugins/qwen-osf` release package | Public Git/local/archive installation available; Qwen has no separate maintained gallery evidenced, with Claude/Gemini channels documented | `qwen-extension-publication_20260711` |
| Cursor, Cline, Roo, Windsurf, VS Code, Zed | `integrations/`, `.cursor/mcp.json`, `.roo/mcp.json`, and `.vscode/mcp.json` | Standard MCP configuration templates available; provider gallery listings not evidenced | `coding-agent-ecosystem-publication_20260711` |

An artifact is not described as submitted or approved until dated provider-side
evidence exists. GitHub Copilot supports repository or marketplace installation
of packages containing skills and MCP configuration; the dedicated Copilot
plugin package follows that structure.

## Registry priorities

The authoritative MCP Registry, Smithery, and Glama are already published. The
next high-value sweep covers Docker MCP Catalog, mcp.so,
`awesome-mcp-servers`, PulseMCP, MCP.Directory, MCPize, and maintained
successors. Lower-quality directories should be skipped when they require broad
permissions, obscure provenance, or provide no stable listing evidence.

See issue [#26](https://github.com/edithatogo/osf-cli-go/issues/26).

The current requirements and evidence scoring contract is documented in
[`docs/registry-scorecard.md`](registry-scorecard.md). Codex and Claude remain
explicit external submission gates even though their local plugin packages are
validated.
