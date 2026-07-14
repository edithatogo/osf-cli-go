# LobeHub MCP Marketplace review

## Evidence

- LobeHub user authentication completed through Chrome on 2026-07-14.
- GitHub account connection completed through the LobeHub OAuth flow.
- Repository import accepted for `https://github.com/edithatogo/osf-cli-go`.
- LobeHub identifier: `edithatogo-osf-cli-go`.
- Public listing: <https://market.lobehub.com/s/plugins/edithatogo-osf-cli-go>.
- Manifest endpoint: <https://market.lobehub.com/api/v1/plugins/edithatogo-osf-cli-go/manifest>.
- `lhm plugin list --output json` reports `published`, `isClaimed: true`, and latest version `0.3.2`.

## Validation

The release-aligned manifest was parsed locally and the repository quality
gates passed after publication. No credentials are stored in the repository.

## Closeout

LobeHub publication is complete. Future releases must rerun the manifest
publication and public listing verification; provider publication does not
replace the repository's release and security gates.
