# OSF CLI Go

Use the `osf` MCP server for authenticated, read-only Open Science Framework
inspection. Prefer `OSF_TOKEN`; public read operations may work without a
token. Never expose credentials and require explicit user intent for writes.

Install the packaged extension locally with:

```text
qwen extensions install ./plugins/qwen-osf
```

For public distribution, use a generated `qwen-osf-<version>-<runtime>.zip`
release archive. The repository also contains Claude marketplace metadata, so
whole-repository installation resolves that marketplace before this Qwen
manifest.
