# Qwen Code Extension Evidence

Last reviewed: 2026-07-11

## Authoritative requirements

- [Qwen Code extensions](https://qwenlm.github.io/qwen-code-docs/en/users/extension/introduction/)
- [Qwen Code configuration and MCP settings](https://github.com/QwenLM/qwen-code/blob/main/docs/users/configuration/settings.md)

Qwen Code supports extensions from Git repositories, local paths, archives,
npm, Claude Code marketplaces, and the Gemini CLI extensions gallery. Its
extension contract requires a `qwen-extension.json` file in the extension root;
the current settings format uses an array with explicit `envVar` and
`sensitive` fields.

## Public repository readiness

- Repository: <https://github.com/edithatogo/osf-cli-go>
- Root manifest: `qwen-extension.json`
- Root context: `QWEN.md`
- Version: `0.3.1`, aligned with `server.json` and client manifests.
- Self-contained release package: `plugins/qwen-osf`, produced by
  `scripts/build-plugin-archives.ps1` with `qwen-extension.json` at archive
  root and a bundled `bin/osf-mcp` binary.

## Local validation

The release contract validates both root and packaged Qwen manifests:

```text
go run ./tools/checkreleasecontract
```

Qwen Code 0.15.6 was exercised from a clean temporary home. Whole-repository
installation resolves the repository's existing Claude marketplace before the
root Qwen manifest, so the supported unambiguous route is the packaged
extension directory or a release archive. The packaged-directory attempt
reached the Qwen settings flow but could not complete on this macOS host
because the system keychain was unavailable. No extension installation is
claimed from that run; JSON validation and archive construction remain
reproducible local gates.

Use the packaged directory during local development:

```text
qwen extensions install ./plugins/qwen-osf
```

Use a generated `qwen-osf-<version>-<runtime>.zip` archive for distribution.

## Exact external status

The packaged repository directory and archive are ready for Qwen installation.
The whole multi-client repository is not advertised as a direct Qwen install
source because its Claude marketplace is selected first. No separate
Qwen-maintained public gallery or provider approval was identified in the
current official documentation. Claude marketplace and Gemini gallery routes
remain available separately and are not claimed here as Qwen approval.
