# LobeHub MCP Marketplace submission packet

Status: prepared; authentication and provider publication are pending.

## Package

- Manifest: [`lhm.plugin.json`](../lhm.plugin.json)
- Repository: <https://github.com/edithatogo/osf-cli-go>
- Release version: `0.3.2`
- MCP transport: stdio
- Credentials: `OSF_TOKEN` preferred, with username/password fallback

The manifest is intentionally release-aligned and contains no credential
values. It describes the read-only OSF tools and points to the public source
repository and logo.

## Validation

The manifest parses as JSON and the current LobeHub CLI recognizes it as a
publish input. Device registration also succeeded and stored M2M credentials
locally. The publish command was run on 2026-07-14 and stopped at the user
publisher authentication boundary:

```text
npx --yes @lobehub/market-cli plugin publish --dir .
Not logged in. Run `lhm login` first.
```

M2M device authentication is available, but it is insufficient for publishing
owned plugins. This is an authentication blocker, not a package-validation
failure. After user login, run the publish command from the repository root and
record the returned identifier/version and public listing URL here. Do not
claim publication from the local manifest alone.
