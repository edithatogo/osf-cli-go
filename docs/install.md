# Installation

## System Requirements

- Go 1.26 or later
- Git (for `go install`)
- Make (optional, for build targets)

## From Source

```bash
go install github.com/edithatogo/osf-cli-go/cmd/osf@latest
```

This places the `osf` binary in `$GOPATH/bin` (defaults to `~/go/bin`).

## From Releases

Pre-built binaries are available on the [GitHub releases page](https://github.com/edithatogo/osf-cli-go/releases).

1. Download the archive for your platform.
2. Extract the `osf` (or `osf.exe`) binary.
3. Place it in a directory on your `PATH`.

The current asset-backed release is `v0.3.2`. Verify the matching
`SHA256SUMS-<platform>-<architecture>.txt` file before installing.

## Package Managers

The submit-ready package definitions are kept under `packaging/`:

- Homebrew formula: `packaging/homebrew/osf-cli-go.rb`
- Scoop manifest: `packaging/scoop/osf-cli-go.json`
- WinGet manifests: `packaging/winget/`

The Homebrew formula will be available through the project tap after the tap
submission is accepted. Scoop and WinGet submissions use the Windows release
assets and their recorded SHA-256 hashes.

## Cross-Platform Builds

Use the Makefile to build for multiple platforms:

```bash
make build # builds for the current platform
```

Release binaries are built by GoReleaser for:

| Platform   | Architecture |
|------------|-------------|
| Linux      | amd64, arm64 |
| macOS      | amd64, arm64 |
| Windows    | amd64        |

## Verifying Installation

```bash
osf version
```

If the token is set, you can verify connectivity:

```bash
osf auth whoami
```
