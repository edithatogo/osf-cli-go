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
publish input. The publish command was run on 2026-07-14 and stopped at the
provider authentication boundary:

```text
npx --yes @lobehub/market-cli plugin publish --dir .
Not logged in. Run `lhm login` first.
```

This is an authentication blocker, not a package-validation failure. After
login, run the publish command from the repository root and record the returned
identifier/version and public listing URL here. Do not claim publication from
the local manifest alone.
