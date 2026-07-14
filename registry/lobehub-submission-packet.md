# LobeHub MCP Marketplace submission packet

Status: published and verified.

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
publish input. Device registration and user authentication succeeded. The
repository was imported, claimed, and published on 2026-07-14.

```text
npx --yes @lobehub/market-cli plugin publish --dir .
Published edithatogo-osf-cli-go (1.0.0 -> 0.3.2)
```

## Public receipt

- Identifier: `edithatogo-osf-cli-go`
- Status: `published`
- Claimed: `true`
- Latest version reported by `lhm plugin list`: `0.3.2`
- Public listing: <https://market.lobehub.com/s/plugins/edithatogo-osf-cli-go>
- Manifest: <https://market.lobehub.com/api/v1/plugins/edithatogo-osf-cli-go/manifest>

Both public URLs returned HTTP 200 during verification. The marketplace search
result reports the listing as validated and public.
