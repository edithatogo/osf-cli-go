# MCPB Bundle Packaging

`manifest.json` is the source manifest for OSF CLI Go MCPB bundles.

Build locally:

```powershell
.\scripts\build-mcpb.ps1
```

Validate and pack with the official MCPB CLI when available:

```powershell
npm install -g @anthropic-ai/mcpb
mcpb validate dist\mcpb\osf-cli-go
mcpb pack dist\mcpb\osf-cli-go dist\mcpb\osf-cli-go-0.3.1-windows-amd64.mcpb
```

The bundle contains:

```text
manifest.json
server/osf-mcp(.exe)
```

The manifest declares only read-only OSF tools and expects credentials through
the client UI or environment-backed user configuration.
