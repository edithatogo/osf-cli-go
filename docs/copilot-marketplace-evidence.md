# GitHub Copilot Marketplace Evidence

Last reviewed: 2026-07-11

## Public repository marketplace

The repository-hosted GitHub Copilot marketplace is defined by
`.github/plugin/marketplace.json` and points to `plugins/github-copilot-osf`.
The public source is:

<https://github.com/edithatogo/osf-cli-go/tree/master/.github/plugin>

Install from that marketplace with:

```text
copilot plugin marketplace add edithatogo/osf-cli-go
copilot plugin marketplace browse osf-cli-go
copilot plugin install osf-cli-go@osf-cli-go
```

The direct repository path remains available for clients that do not support
marketplace registration:

```text
copilot plugin install edithatogo/osf-cli-go:plugins/github-copilot-osf
```

## Local validation

The release contract validates the marketplace JSON, its plugin source path,
and version alignment with `server.json` and all other client manifests:

```text
go run ./tools/checkreleasecontract
```

With GitHub Copilot CLI 1.0.69 and an isolated temporary `HOME`, the local
marketplace was added, browsed, and installed successfully:

```text
Marketplace "osf-cli-go" added successfully.
Plugins in "osf-cli-go":
  • osf-cli-go - Open Science Framework tools for GitHub Copilot CLI and coding agent.
Plugin "osf-cli-go" installed successfully. Installed 1 skill.
Installed plugins:
  • osf-cli-go@osf-cli-go (v0.3.1)
```

This validates the repository marketplace contract without changing the
developer's normal Copilot configuration.

## Exact external status

The repository-hosted marketplace is available. No evidence currently shows
acceptance into a GitHub-maintained default marketplace or provider approval
outside this repository. The project does not describe the package as
approved without dated provider-side evidence.
