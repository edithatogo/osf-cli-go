# 1.0 Migration Guide

There is no breaking migration from the current 0.3.x contract to the planned
1.x contract. Existing CLI JSON fields, MCP tool names, and required MCP input
properties remain compatible under the documented policy.

Before upgrading automation:

1. Pin the release or container tag rather than an unversioned development
   checkout.
2. Treat unknown JSON fields as forward-compatible additions.
3. Keep `OSF_TOKEN` as the preferred credential and do not persist credentials
   in project files.
4. Review release notes for any explicitly deprecated command, field, or tool.

The additive 2026-07-15 contract introduces public Zenodo REST reads under
`osf zenodo records|files`, public OAI-PMH harvesting under `osf zenodo oai`,
and three `zenodo_oai_*` MCP tools. Existing OSF automation requires no
migration. OAI-PMH consumers should persist the returned opaque resumption
token and must not treat it as a Zenodo REST page URL.

The first breaking change, if one becomes necessary, will require a major
version, a replacement path, and at least one minor-release deprecation period
where practical.
