# osf-cli-go

`osf-cli-go` is a Go command-line client for the Open Science Framework (OSF). It is designed for researchers, research software engineers, and data stewards who need predictable terminal workflows for OSF projects, files, registrations, and exports.

## Current Capability

- Authenticate with an OSF personal access token from `OSF_TOKEN`.
- Inspect the authenticated account with `osf auth whoami`.
- List and inspect projects, components, and OSF Storage files.
- Download files and folder trees with conservative conflict handling.
- Upload files, create OSF Storage folders, and remove files through explicit WaterButler commands.
- Search OSF, list preprints, create draft registrations, and export node snapshots.
- Emit table output for humans and JSON output for automation.

## Start Here

- [Install](install.md) explains local and release-based installation.
- [Usage](usage.md) covers authentication, output modes, and common workflows.
- [Commands](commands.md) lists the current CLI command surface.
- [Examples](examples.md) provides copyable workflow examples.
- [Development](development.md) documents local validation and repository hardening expectations.
- [1.0 launch review](v1-launch-review.md) records dated readiness evidence and explicit waivers.

## Status

The CLI is offline-tested for the implemented command surface. Live OSF validation remains opt-in and requires explicit environment variables and disposable OSF resources.
